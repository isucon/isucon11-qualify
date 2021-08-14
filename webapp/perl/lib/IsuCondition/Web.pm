package IsuCondition::Web;
use v5.34;
use warnings;
use utf8;

use experimental qw(isa try signatures);

use Kossy;

use DBIx::Sunny;
use File::Slurp qw(read_file);
use HTTP::Status qw/:constants/;
use Log::Minimal;
use JSON::MaybeXS;
use Cpanel::JSON::XS::Type;

local $Log::Minimal::LOG_LEVEL = "DEBUG";

my $MYSQL_CONNECTION_DATA = {
    host     => $ENV{MYSQL_HOST}   # '127.0.0.1',
    port     => $ENV{MYSQL_PORT}   # '3306',
    user     => $ENV{MYSQL_USER}   # 'isucon',
    dbname   => $ENV{MYSQL_DBNAME} # 'isucondition',
    password => $ENV{MYSQL_PASS}   # 'isucon',
};


use constant Isu => {
    id           => JSON_TYPE_INT,
    jia_isu_uuid => JSON_TYPE_STRING,
    name         => JSON_TYPE_STRING,
    image        => undef,
    character    => JSON_TYPE_STRING,
    jia_user_id  => undef,
    created_at   => undef,
    updated_at   => undef,
};

use constant InitializeResponse => {
    language => JSON_TYPE_STRING,
};

use constant GetMeResponse => {
    jia_user_id => JSON_TYPE_STRING,
};

use constant ConditionsPercentage => {
    sitting       => JSON_TYPE_INT,
    is_broken     => JSON_TYPE_INT,
    is_dirty      => JSON_TYPE_INT,
    is_overweight => JSON_TYPE_INT,
};

use constant GraphDataPoint => {
    score      => JSON_TYPE_INT,
    percentage => ConditionsPercentage,
};

use constant GraphResponse => {
    start_at             => JSON_TYPE_INT,
    end_at               => JSON_TYPE_INT,
    data                 => GraphDataPoint,
    condition_timestamps => json_type_arrayof(JSON_TYPE_INT),
};

use constant GetIsuConditionResponse => {
    jia_isu_uuid    => JSON_TYPE_STRING,
    isu_name        => JSON_TYPE_STRING,
    timestamp       => JSON_TYPE_INT,
    is_sitting      => JSON_TYPE_BOOL,
    condition       => JSON_TYPE_STRING,
    condition_level => JSON_TYPE_STRING,
    message         => JSON_TYPE_STRING,
};

use constant GetIsuListResponse => {
    id                   => JSON_TYPE_INT,
    jia_isu_uuid         => JSON_TYPE_STRING,
    name                 => JSON_TYPE_STRING,
    character            => JSON_TYPE_STRING,
    latest_isu_condition => GetIsuConditionResponse,
};

use constant TrendCondition => {
    isu_id    => JSON_TYPE_INT,
    timestamp => JSON_TYPE_INT,
};

use constant TrendResponse => {
    character => JSON_TYPE_STRING,
    info      => json_type_arrayof(TrendCondition),
    warning   => json_type_arrayof(TrendCondition),
    critical  => json_type_arrayof(TrendCondition),
};



sub get_user_id_from_session($c) {
    my $jia_user_id = $c->req->session->get('jia_user_id');

    if (!$jia_user_id) {
        $c->halt_no_content(HTTP_UNAUTHORIZED, "you are not signed in");
    }

    my $count = $self->dbh->select_one("SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = ?", $jia_user_id);
    if ($count == 0) {
        $c->halt_no_content(HTTP_UNAUTHORIZED, "you are not signed in");
    }

    return $jia_user_id;
}

sub get_jia_service_url($dbh) {
    my $config = $dbh->select_row("SELECT * FROM `isu_association_config` WHERE `name` = ?", "jia_service_url");
    if (!$config) {
        return $DEFAULT_JIA_SERVICE_URL;
    }
    return $config->{url};
}

# POST /initialize
# サービスを初期化
post '/initialize' => [qw/allow_json_request/] => sub ($self, $c) {
    if (!$c->req->parameters->{'jia_service_url'}) {
        $c->halt_text(HTTP_BAD_REQUEST, "bad request body");
    }

    my $e = system("../sql/init.sh");
    if ($e) {
        warnf("exec init.sh error: %s", $e);
        $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
    }

    $self->query(
        "INSERT INTO `isu_association_config` (`name`, `url`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)",
        "jia_service_url",
        $c->req->parameters->{'jia_service_url'},
    );

    return $c->res_json({
        language => "perl",
    }, InitializeResponse);
};

# POST /api/auth
# サインアップ・サインイン
post '/api/auth' => sub ($self, $c) {
    my $req_jwt;
    # TODO
    #reqJwt := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

    my $jwt;
    try {
        my $token = $jwt->parse($req_jwt);
    }
    catch ($e) {
    }
    # TODO
    #token, err := jwt.Parse(reqJwt, func(token *jwt.Token) (interface{}, error) {
	#	if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
	#		return nil, jwt.NewValidationError(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]), jwt.ValidationErrorSignatureInvalid)
	#	}
	#	return jiaJWTSigningKey, nil
	#})
    #}
    #catch ($e) {
    #    #if (case *jwt.ValidationError:) {
    #    if (0) {
    #        return $c->render_text(HTTP_FORBIDDEN);
    #    }
    #    else {
    #        warnf($e);
    #        return $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
    #    }
    #}
    #warnf("invalid JWT payload")

    my $jia_user_id;

    $self->query("INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)", $jia_user_id);

    $c->req->session->set(jia_user_id => $jia_user_id);

    return $c->halt_no_content(HTTP_OK);
};

# POST /api/signout
# サインアウト
post '/api/signout' => sub ($self, $c) {
    my $jia_user_id = get_user_id_from_session($c);
    $c->req->session->clear();

    return $c->halt_no_content(HTTP_OK);
};

# GET /api/user/me
# サインインしている自分自身の情報を取得
get '/api/user/me' => sub ($self, $c) {
    my $jia_user_id = get_user_id_from_session($c);

    return $c->render_json({
        jia_user_id => $jia_user_id,
    }, GetMeResponse);
};

# GET /api/isu
# ISUの一覧を取得
get '/api/isu' => sub ($self, $c) {
    my $jia_user_id = get_user_id_from_session($c);

    my $isu_list = $self->dbh->select_all(
		"SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC",
		$jia_user_id
    );

    my $response_list = [];
    for my $isu ($isu_list->@*) {
        my $found_last_condition = !!1;
        my $last_condition = $self->dbh->select_row(
		    "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` DESC LIMIT 1",
            $isu->{jia_isu_uuid}
        );
        if (!$last_condition) {
            $found_last_condition = !!0;
        }

        my $formatted_condition;
        if ($found_last_condition) {
			my ($condition_level, $e) = calculate_condition_level($last_condition->{condition});
            if ($e) {
                warnf($e);
                return $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
			}

			$formatted_condition = {
				jia_isu_uuid =>      $last_condition->{jia_isu_uuid},
				isu_name =>         $isu->{name},
				timestamp =>       $last_condition->{timestamp}, # TODO unix timestamp
				is_sitting =>       $last_condition->{is_sitting},
				condition =>       $last_condition->{condition},
				condition_level =>  $condition_level,
				message =>         $last_condition->{message},
			}
        }

        my $res = {
            id => $isu->{id},
            jia_isu_uuid => $isu->{jia_isu_uuid},
            name => $isu->{name},
            character => $isu->{character},
            latest_isu_condition => $formatted_condition,
        };
        push $response_list->@* => $res;
    }

    return $c->render_json($response_list, json_type_arrayof(GetIsuListResponse));
};

# POST /api/isu
# ISUを登録
post '/api/isu' => sub ($self, $c) {
    my $jia_user_id = get_user_id_from_session($c);

    my $use_default_image = !!0;

    my $jia_isu_uuid = $c->req->parameters->{'jia_isu_uuid'};
    my $isu_name = $c->req->parameters->{'isu_name'};

    my $file = $c->req->uploads->{'image'};
    if (!$file) {
        $use_default_image = !!1;
    }

    my $image;
    if ($use_default_image) {
        $image = read_file($DEFAULT_ICON_PATH, binmode => ':raw');
    }
    else {
        $image = read_file($file, binmode => ':raw');
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope;
    try {
        $dbh->query("INSERT INTO `isu`"+
		"	(`jia_isu_uuid`, `name`, `image`, `jia_user_id`) VALUES (?, ?, ?, ?)",
		$jia_isu_uuid, $isu_name, $image, $jia_user_id)

        my $target_url = get_jia_service_url($dbh) + "/api/activate";
        my $body = {
            target_base_url => $POST_ISUCONDITION_TARGET_BASE_URL,
            isu_uuid => $jia_isu_uuid,
        };

        # TODO
        my $furl = Furl->new;;
        my $res = $furl->post(
            $target_url,
            $body, # as json payload

            # reqJIA.Header.Set("Content-Type", "application/json")
        );

        if (!$res->is_success) {
            warnf("failed to request to JIAService: %s", $res); # TODO
            $self->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
        }

        if ($res->status != HTTP_ACCEPTED) {
            warnf("JIAService returned error: status code %s, message: %s", $res->status, $res->body)
            return $c->halt_text($res->status, "JIAService returned error");
        }

        my $isu_from_jia = $res->decoded_content;

        $self->dbh->query("UPDATE `isu` SET `character` = ? WHERE  `jia_isu_uuid` = ?", $isu_from_jia->{character}, $jia_isu_uuid);

        my $isu = $self->dbh->select_row(
            "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
            $jia_user_id, $jia_isu_uuid);

        $txn->commit;
    }
    catch ($e) {
        $txn->rollback;
        warnf("db error: %s", $e);
        $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
    }

    return $c->render_json(HTTP_CREATED, $isu, Isu);
}

# GET /api/isu/:jia_isu_uuid
# ISUの情報を取得
get '/api/isu/{jia_isu_uuid}' => sub ($self, $c) {
    my $jia_user_id = get_user_id_from_session($c);

	my $jia_isu_uuid = $c->req->parameters->{'jia_isu_uuid'};

    my $isu = $self->dbh->select_row(
        "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
		$jia_user_id, $jia_isu_uuid);

    if (!$isu) {
        $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
    }

    return $c->render_json(HTTP_OK, $isu, Isu);
};

# GET
# ISUのアイコンを取得
get '/api/isu/:jia_isu_uuid/icon' => sub ($self, $c) {
    my $jia_user_id = get_user_id_from_session($c);

	my $jia_isu_uuid = $c->req->parameters->{'jia_isu_uuid'};

    my $image = $self->dbh->select_one(
        "SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
		$jia_user_id, $jia_isu_uuid);

    if (!$isu) {
        $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
    }

    # TODO
    # return c.Blob(http.StatusOK, "", image)
    $c->res->status(HTTP_OK);
    $c->res->body($image);
    return $c->res;
}

# GET /api/isu/:jia_isu_uuid/graph
# ISUのコンディショングラフ描画のための情報を取得
get '/api/isu/{jia_isu_uuid}/graph' => sub ($self, $c) {
    my $jia_user_id = get_user_id_from_session($c);

	my $jia_isu_uuid = $c->req->parameters->{'jia_isu_uuid'};
    my $datetime = $c->req->parameters->{'datetime'};

    # TODO
    #	jiaIsuUUID := c.Param("jia_isu_uuid")
    #	datetimeStr := c.QueryParam("datetime")
    #	if datetimeStr == "" {
    #		return c.String(http.StatusBadRequest, "missing: datetime")
    #	}
    #	datetimeInt64, err := strconv.ParseInt(datetimeStr, 10, 64)
    #	if err != nil {
    #		return c.String(http.StatusBadRequest, "bad format: datetime")
    #	}
    #	date := time.Unix(datetimeInt64, 0).Truncate(time.Hour)

    my $count = $self->select_one(
        "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
        $jia_user_id, $jia_isu_uuid);

    if ($count == 0) {
        return $c->halt_text(HTTP_NOT_FOUND, "not found: isu")
    }

	my $res = generateIsuGraphResponse(tx, $jia_isu_uuid, $date);

    return $c->render_json($res, json_type_arrayof(GraphResponse));
}

# グラフのデータ点を一日分生成
sub generateIsuGraphResponse(tx *sqlx.Tx, jiaIsuUUID string, graphDate time.Time) ([]GraphResponse, error) {
	dataPoints := []GraphDataPointWithInfo{}
	conditionsInThisHour := []IsuCondition{}
	timestampsInThisHour := []int64{}
	var startTimeInThisHour time.Time
	var condition IsuCondition

	rows, err := tx.Queryx("SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC", jiaIsuUUID)
	if err != nil {
		return nil, fmt.Errorf("db error: %v", err)
	}

	for rows.Next() {
		err = rows.StructScan(&condition)
		if err != nil {
			return nil, err
		}

		truncatedConditionTime := condition.Timestamp.Truncate(time.Hour)
		if truncatedConditionTime != startTimeInThisHour {
			if len(conditionsInThisHour) > 0 {
				data, err := calculateGraphDataPoint(conditionsInThisHour)
				if err != nil {
					return nil, err
				}

				dataPoints = append(dataPoints,
					GraphDataPointWithInfo{
						JIAIsuUUID:          jiaIsuUUID,
						StartAt:             startTimeInThisHour,
						Data:                data,
						ConditionTimestamps: timestampsInThisHour})
			}

			startTimeInThisHour = truncatedConditionTime
			conditionsInThisHour = []IsuCondition{}
			timestampsInThisHour = []int64{}
		}
		conditionsInThisHour = append(conditionsInThisHour, condition)
		timestampsInThisHour = append(timestampsInThisHour, condition.Timestamp.Unix())
	}

	if len(conditionsInThisHour) > 0 {
		data, err := calculateGraphDataPoint(conditionsInThisHour)
		if err != nil {
			return nil, err
		}

		dataPoints = append(dataPoints,
			GraphDataPointWithInfo{
				JIAIsuUUID:          jiaIsuUUID,
				StartAt:             startTimeInThisHour,
				Data:                data,
				ConditionTimestamps: timestampsInThisHour})
	}

	endTime := graphDate.Add(time.Hour * 24)
	startIndex := 0
	endNextIndex := len(dataPoints)
	for i, graph := range dataPoints {
		if startIndex == 0 && !graph.StartAt.Before(graphDate) {
			startIndex = i
		}
		if endNextIndex == len(dataPoints) && graph.StartAt.After(endTime) {
			endNextIndex = i
		}
	}

	filteredDataPoints := []GraphDataPointWithInfo{}
	if startIndex < endNextIndex {
		filteredDataPoints = dataPoints[startIndex:endNextIndex]
	}

	responseList := []GraphResponse{}
	index := 0
	thisTime := graphDate

	for thisTime.Before(graphDate.Add(time.Hour * 24)) {
		var data *GraphDataPoint
		timestamps := []int64{}

		if index < len(filteredDataPoints) {
			dataWithInfo := filteredDataPoints[index]

			if dataWithInfo.StartAt.Equal(thisTime) {
				data = &dataWithInfo.Data
				timestamps = dataWithInfo.ConditionTimestamps
				index++
			}
		}

		resp := GraphResponse{
			StartAt:             thisTime.Unix(),
			EndAt:               thisTime.Add(time.Hour).Unix(),
			Data:                data,
			ConditionTimestamps: timestamps,
		}
		responseList = append(responseList, resp)

		thisTime = thisTime.Add(time.Hour)
	}

	return responseList, nil
}

# 複数のISUのコンディションからグラフの一つのデータ点を計算
sub calculateGraphDataPoint(isuConditions []IsuCondition) (GraphDataPoint, error) {
	conditionsCount := map[string]int{"is_broken": 0, "is_dirty": 0, "is_overweight": 0}
	rawScore := 0
	for _, condition := range isuConditions {
		badConditionsCount := 0

		if !isValidConditionFormat(condition.Condition) {
			return GraphDataPoint{}, fmt.Errorf("invalid condition format")
		}

		for _, condStr := range strings.Split(condition.Condition, ",") {
			keyValue := strings.Split(condStr, "=")

			conditionName := keyValue[0]
			if keyValue[1] == "true" {
				conditionsCount[conditionName] += 1
				badConditionsCount++
			}
		}

		if badConditionsCount >= 3 {
			rawScore += scoreConditionLevelCritical
		} else if badConditionsCount >= 1 {
			rawScore += scoreConditionLevelWarning
		} else {
			rawScore += scoreConditionLevelInfo
		}
	}

	sittingCount := 0
	for _, condition := range isuConditions {
		if condition.IsSitting {
			sittingCount++
		}
	}

	isuConditionsLength := len(isuConditions)

	score := rawScore / isuConditionsLength

	sittingPercentage := sittingCount * 100 / isuConditionsLength
	isBrokenPercentage := conditionsCount["is_broken"] * 100 / isuConditionsLength
	isOverweightPercentage := conditionsCount["is_overweight"] * 100 / isuConditionsLength
	isDirtyPercentage := conditionsCount["is_dirty"] * 100 / isuConditionsLength

	dataPoint := GraphDataPoint{
		Score: score,
		Percentage: ConditionsPercentage{
			Sitting:      sittingPercentage,
			IsBroken:     isBrokenPercentage,
			IsOverweight: isOverweightPercentage,
			IsDirty:      isDirtyPercentage,
		},
	}
	return dataPoint, nil
}

# GET /api/condition/:jia_isu_uuid
# ISUのコンディションを取得
get '/api/condition/:jia_isu_uuid' => sub {
    my $jia_user_id = get_user_id_from_session($c);
    my $jia_isu_uuid = $c->req->parameters->{'jia_isu_uuid'};
    if (!$jia_isu_uuid) {
        $c->halt_text(HTTP_BAD_REQUEST, "missing: jia_isu_uuid");
    }

    my $end_time = $c->req->parameters->{'end_time'};
    if (!$end_time) { #TODO
        #endTimeInt64, err := strconv.ParseInt(c.QueryParam("end_time"), 10, 64)
        $c->halt_text(HTTP_BAD_REQUEST, "bad format: end_time");
    }

    my $condition_level_csv = $c->req->parameters->{'condition_level'};
    if (!$condition_level_csv) {
        $c->halt_text(HTTP_BAD_REQUEST, "missing: condition_level");
    }
    my $condition_level = {};
    for my $level (split $condition_level_csv, ",") {
        $condition_level->{$level} = {};
    }
    # TODO
    #conditionLevel := map[string]interface{}{}
	#for _, level := range strings.Split(conditionLevelCSV, ",") {
	#	conditionLevel[level] = struct{}{}
	#}

    my $start_time = $c->req->parameters->{'start_time'};
    if (!$start_time) { #TODO
        #startTimeInt64, err := strconv.ParseInt(c.QueryParam("start_time"), 10, 64)
        $c->halt_text(HTTP_BAD_REQUEST, "bad format: start_time");
    }


    my $isu_name = $self->dbh->select_one(
		"SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ?",
		$jia_isu_uuid, $jia_user_id);

    if (!$isu_name) {
        $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
    }

	my $conditions_response = get_isu_conditions_from_db(db, jiaIsuUUID, endTime, conditionLevel, startTime, conditionLimit, isuName);
    return $c->render_json($conditions_response);
};

# ISUのコンディションをDBから取得
sub getIsuConditionsFromDB(db *sqlx.DB, jiaIsuUUID string, endTime time.Time, conditionLevel map[string]interface{}, startTime time.Time,
	limit int, isuName string) ([]*GetIsuConditionResponse, error) {

	conditions := []IsuCondition{}
	var err error

	if startTime.IsZero() {
		err = db.Select(&conditions,
			"SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?"+
				"	AND `timestamp` < ?"+
				"	ORDER BY `timestamp` DESC",
			jiaIsuUUID, endTime,
		)
	} else {
		err = db.Select(&conditions,
			"SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?"+
				"	AND `timestamp` < ?"+
				"	AND ? <= `timestamp`"+
				"	ORDER BY `timestamp` DESC",
			jiaIsuUUID, endTime, startTime,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("db error: %v", err)
	}

	conditionsResponse := []*GetIsuConditionResponse{}
	for _, c := range conditions {
		cLevel, err := calculateConditionLevel(c.Condition)
		if err != nil {
			continue
		}

		if _, ok := conditionLevel[cLevel]; ok {
			data := GetIsuConditionResponse{
				JIAIsuUUID:     c.JIAIsuUUID,
				IsuName:        isuName,
				Timestamp:      c.Timestamp.Unix(),
				IsSitting:      c.IsSitting,
				Condition:      c.Condition,
				ConditionLevel: cLevel,
				Message:        c.Message,
			}
			conditionsResponse = append(conditionsResponse, &data)
		}
	}

	if len(conditionsResponse) > limit {
		conditionsResponse = conditionsResponse[:limit]
	}

	return conditionsResponse, nil
}

# ISUのコンディションの文字列からコンディションレベルを計算
sub calculateConditionLevel(condition string) (string, error) {
	var conditionLevel string

	warnCount := strings.Count(condition, "=true")
	switch warnCount {
	case 0:
		conditionLevel = conditionLevelInfo
	case 1, 2:
		conditionLevel = conditionLevelWarning
	case 3:
		conditionLevel = conditionLevelCritical
	default:
		return "", fmt.Errorf("unexpected warn count")
	}

	return conditionLevel, nil
}

# GET /api/trend
# ISUの性格毎の最新のコンディション情報
get '/api/trend' => sub ($self, $c) {

	characterList := []Isu{}
    my $character_list = $self->select_all(
	    "SELECT `character` FROM `isu` GROUP BY `character`")

    my $trend_response = [];
    for my $character ($character_list->@*) {
        my $isu_list = $self->dbh->select_all(
			"SELECT * FROM `isu` WHERE `character` = ?",
			$character->{character},
        );

        my $character_isu_conditions = {};

        for my $isu ($isu_list->@*) {
            my $conditions = $self->dbh->select_all(
				"SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY timestamp DESC",
				$isu->{jia_isu_uuid},
			)

            if ($conditions->@* > 0) {
                my $isu_last_condition = $conditions->[0];
				my $condition_level = calculateConditionLevel($isu_last_condition->{condition});
                my $trend_condition = {
                    id => $isu->{id},
                    timestamp => $isu->{timestamp},
                }
                push $character_isu_conditions->{$condition_level}->@* => $trend_condition;
            }
        }

        for my $level (keys $character_isu_conditions->%*) {
            my $conditions = $character_isu_conditions->{$level};
            my @sorted_conditions = sort { $a->{timestamp} > $b->{timestamp} } $conditions->@*;
            $character_isu_conditions->{$level} = \@sorted_conditions;
        }

        push $trend_response->@* => {
            character => $character->{character},
            info      => $character_isu_conditions->{info},
            warning   => $character_isu_conditions->{warning},
            critical  => $character_isu_conditions->{critical},
        };
    }

    return $c->render_json($trend_response, json_type_arrayof(TrendResponse));
}

# POST /api/condition/:jia_isu_uuid
# ISUからのコンディションを受け取る
post '/api/condition/:jia_isu_uuid' => [qw/allow_json_request/] => sub ($self, $c) {
	# TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき
	my $drop_probability = 0.9;
	if (rand <= $drop_probability) {
		warnf("drop post isu condition request")
        $c->halt_no_content(HTTP_SERVICE_UNAVAILABLE);
	}

    my $jia_isu_uuid = $c->req->parameters->{'jia_isu_uuid'};
    if (!$jia_isu_uuid) {
        $c->halt_text(HTTP_BAD_REQUEST, "missing: jia_isu_uuid");
    }

    # TODO
    #req := []PostIsuConditionRequest{}
	#err := c.Bind(&req)
	#if err != nil {
	#	return c.String(http.StatusBadRequest, "bad request body")
	#} else if len(req) == 0 {
	#	return c.String(http.StatusBadRequest, "bad request body")
	#}

	tx, err := db.Beginx()
	if err != nil {
		warnf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()


    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope;
    try {
        my $count = $self->dbh->select_one("SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?", $jia_isu_uuid);
        if ($count == 0) {
            $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
        }
        # TODO
        #	for _, cond := range req {
        #		timestamp := time.Unix(cond.Timestamp, 0)
        #
        #		if !isValidConditionFormat(cond.Condition) {
        #			return c.String(http.StatusBadRequest, "bad request body")
        #		}
        #
        #		_, err = tx.Exec(
        #			"INSERT INTO `isu_condition`"+
        #				"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)"+
        #				"	VALUES (?, ?, ?, ?, ?)",
        #			jiaIsuUUID, timestamp, cond.IsSitting, cond.Condition, cond.Message)
        #		if err != nil {
        #			warnf("db error: %v", err)
        #			return c.NoContent(http.StatusInternalServerError)
        #		}
    }
    catch($e) {
        $txn->rollback;
		warnf("db error: %s", $e);
        $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
    }

    return $c->halt_no_content(HTTP_CREATED);
}

# ISUのコンディションの文字列がcsv形式になっているか検証
sub is_valid_condition_format($condition_str) {

	keys := []string{"is_dirty=", "is_overweight=", "is_broken="}
	const valueTrue = "true"
	const valueFalse = "false"

	idxCondStr := 0

	for idxKeys, key := range keys {
		if !strings.HasPrefix(conditionStr[idxCondStr:], key) {
			return false
		}
		idxCondStr += len(key)

		if strings.HasPrefix(conditionStr[idxCondStr:], valueTrue) {
			idxCondStr += len(valueTrue)
		} else if strings.HasPrefix(conditionStr[idxCondStr:], valueFalse) {
			idxCondStr += len(valueFalse)
		} else {
			return false
		}

		if idxKeys < (len(keys) - 1) {
			if conditionStr[idxCondStr] != ',' {
				return false
			}
			idxCondStr++
		}
	}

	return (idxCondStr == len(conditionStr))
}

get "/" => sub ($self, $c) {
    my $file = $FRONTEND_CONTENTS_PATH + "/index.html";
    # FIXME
    return $c->xxxx($file);
}

filter 'allow_json_request' => sub {
    my $app = shift;
    return sub {
        my ($self, $c) = @_;
        $c->env->{'kossy.request.parse_json_body'} = 1;
        $app->($self, $c);
    };
};

sub dbh {
    my $self = shift;
    $self->{_dbh} ||= do {
        my ($host, $port, $user, $dbname, $password) = $MYSQL_CONNECTION_DATA->@{qw/host port user dbname password/};
        my $dsn = "dbi:mysql:database=$dbname;host=$host;port=$port";
        DBIx::Sunny->connect($dsn, $user, $password, {
            mysql_enable_utf8mb4 => 1,
            mysql_auto_reconnect => 1,
            Callbacks => {
                connected => sub {
                    my $dbh = shift;
                    # XXX $dbh->do('SET SESSION sql_mode="STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"');
                    return;
                },
            },
        });
    };
}

# XXX hack Kossy::Connection
{

    my $orig = \&Kossy::Exception::response;
    *Kossy::Exception::response = sub {
        my $self = shift;
        if ($self->{my_response}) {
            return $self->{my_response};
        }
        return $orig->();
    };

    *Kossy::Connection::halt_text = sub {
        my ($c, $status, $text) = @_;
        $c->res->status($status);
        $c->res->content_type('text/plain');
        $c->res->body($text);
        $c->halt($status, my_response => $c->res);
    };

    *Kossy::Connection::halt_no_content = sub {
        my ($c, $status) = @_;
        $c->res->headers->remove_content_headers;
        $c->res->content_length(0);
        $c->res->code($status);
        $c->halt($status, my_response => $c->res);
    };

    # override
    my $_JSON = JSON::MaybeXS->new()->allow_blessed(1)->convert_blessed(1)->ascii(1);
    *Kossy::Connection::render_json = sub {
        my ($c, $obj, $json_spec) = @_;

        my $body = $_JSON->encode($obj, $json_spec); # Cpanel::JSON::XS::Typeを利用する
        $body = $c->escape_json($body);

        if ( ( $c->req->env->{'HTTP_USER_AGENT'} || '' ) =~ m/Safari/ ) {
            $body = "\xEF\xBB\xBF" . $body;
        }

        $c->res->status( 200 );
        $c->res->content_type('application/json; charset=UTF-8');
        $c->res->header( 'X-Content-Type-Options' => 'nosniff' ); # defense from XSS
        $c->res->body( $body );
        $c->res;
    };
}

1
