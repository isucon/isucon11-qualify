package IsuCondition::Web;
use v5.34;
use warnings;
use utf8;

use experimental qw(isa try signatures);

use Kossy;

use DBIx::Sunny;
use File::Slurp qw(read_file);
use HTTP::Status qw(:constants);
use Log::Minimal;
use JSON::MaybeXS qw(encode_json decode_json);
use Cpanel::JSON::XS::Type;
use Crypt::JWT qw(decode_jwt);
use Furl;
use Time::Moment;
use Scalar::Util qw(looks_like_number);

use constant {
    CONDITION_LIMIT              => 20,
    FRONTEND_CONTENTS_PATH       => "../public",
    JIA_JWT_SIGNING_KEY_PATH     => "../ec256-public.pem",
    DEFAULT_ICON_FILE_PATH       => "../NoImage.jpg",
    DEFAULT_JIA_SERVICE_URL      => "http://localhost:5000",
    MYSQL_ERRNUM_DUPLICATE_ENTRY => 1062,
};

use constant {
    CONDITION_LEVEL_INFO     => 'info',
    CONDITION_LEVEL_WARNING  => 'warning',
    CONDITION_LEVEL_CRITICAL => 'critical',
};

use constant {
    SCORE_CONDITION_LEVEL_INFO     => 3,
    SCORE_CONDITION_LEVEL_WARNING  => 2,
    SCORE_CONDITION_LEVEL_CRITICAL => 1,
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
    data                 => json_type_null_or_anyof(GraphDataPoint),
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
    latest_isu_condition => json_type_null_or_anyof(GetIsuConditionResponse),
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

use constant PostIsuConditionRequest => {
    is_sitting => JSON_TYPE_BOOL,
    condition  => JSON_TYPE_STRING,
    message    => JSON_TYPE_STRING,
    timestamp  => JSON_TYPE_INT,
};

sub MYSQL_CONNECTION_ENV() {
    return {
        host     => $ENV{MYSQL_HOST}   || '127.0.0.1',
        port     => $ENV{MYSQL_PORT}   || '3306',
        user     => $ENV{MYSQL_USER}   || 'isucon',
        dbname   => $ENV{MYSQL_DBNAME} || 'isucondition',
        password => $ENV{MYSQL_PASS}   || 'isucon',
    };
};

sub JIA_JWT_SIGNING_KEY() {
    Crypt::PK::ECC->new(JIA_JWT_SIGNING_KEY_PATH);
}

sub POST_ISUCONDITION_TARGET_BASE_URL() {
    my $url = $ENV{POST_ISUCONDITION_TARGET_BASE_URL};
    if (!$url) {
        critf("missing: POST_ISUCONDITION_TARGET_BASE_URL");
    }
    return $url;
}

sub get_user_id_from_session($self, $c) {
    my $jia_user_id = $c->session->get('jia_user_id');
    if (!$jia_user_id) {
        $c->halt_text(HTTP_UNAUTHORIZED, "you are not signed in");
    }

    my $count = $self->dbh->select_one("SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = ?", $jia_user_id);
    if ($count == 0) {
        $c->halt_text(HTTP_UNAUTHORIZED, "you are not signed in");
    }

    return $jia_user_id;
}

sub get_jia_service_url($self) {
    my $config = $self->dbh->select_row("SELECT * FROM `isu_association_config` WHERE `name` = ?", "jia_service_url");
    if (!$config) {
        return DEFAULT_JIA_SERVICE_URL;
    }
    return $config->{url};
}


post("/initialize",                   [qw/allow_json_request/], \&post_initialize);

post("/api/auth",                     \&post_authentication);
post("/api/signout",                  \&post_signout);
get("/api/user/me",                   \&get_me);
get("/api/isu",                       \&get_isu_list);
post("/api/isu",                      \&post_isu);
get("/api/isu/{jia_isu_uuid}/icon",   \&get_isu_icon);
get("/api/isu/{jia_isu_uuid}/graph",  \&get_isu_graph);
get("/api/isu/{jia_isu_uuid}",        \&get_isu_id);
get("/api/condition/{jia_isu_uuid}",  \&get_isu_conditions);
get("/api/trend",                     \&get_trend);

post("/api/condition/{jia_isu_uuid}", [qw/allow_json_request/], \&post_isu_condition);

get("/",                              \&get_index);
get("/isu/{jia_isu_uuid}/condition",  \&get_index);
get("/isu/{jia_isu_uuid}/graph",      \&get_index);
get("/isu/{jia_isu_uuid}",            \&get_index);
get("/register",                      \&get_index);

# POST /initialize
# サービスを初期化
sub post_initialize($self, $c) {
    if (!$c->req->parameters->{'jia_service_url'}) {
        $c->halt_text(HTTP_BAD_REQUEST, "bad request body");
    }

    my $e = system("../sql/init.sh");
    if ($e) {
        warnf("exec init.sh error: %s", $e);
        $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
    }

    $self->dbh->query(
        "INSERT INTO `isu_association_config` (`name`, `url`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)",
        "jia_service_url",
        $c->req->parameters->{'jia_service_url'},
    );

    return $c->render_json({
        language => "perl",
    }, InitializeResponse);
}

# POST /api/auth
# サインアップ・サインイン
sub post_authentication($self, $c) {
    my $req_jwt = $c->req->header('Authorization') =~ s/^Bearer //r;

    try {
        my $payload = decode_jwt(token => $req_jwt, key => JIA_JWT_SIGNING_KEY, accepted_alg => 'ES256');

        my $jia_user_id = $payload->{'jia_user_id'};
        if (!$jia_user_id || ref($jia_user_id)) {
            $c->halt_text(HTTP_BAD_REQUEST, 'invalid JWT payload');
        }

        $self->dbh->query("INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)", $jia_user_id);

        $c->session->set(jia_user_id => $jia_user_id);
    }
    catch ($e) {
        if ($e isa Kossy::Exception) {
            die $e; # rethrow
        }
        elsif ($e =~ /DBD::mysql::st execute failed:/) {
            warnf("db error: %s", $e);
            $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
        }
        $c->halt_text(HTTP_FORBIDDEN, "forbidden");
    }
    $c->halt_no_content(HTTP_OK);
}

# POST /api/signout
# サインアウト
sub post_signout($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);
    $c->session->expire();

    $c->halt_no_content(HTTP_OK);
}

# GET /api/user/me
# サインインしている自分自身の情報を取得
sub get_me($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);

    return $c->render_json({
        jia_user_id => $jia_user_id,
    }, GetMeResponse);
}

# GET /api/isu
# ISUの一覧を取得
sub get_isu_list($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);

    my $isu_list = $self->dbh->select_all(
        "SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC",
        $jia_user_id
    );

    my $response_list = []; # GetIsuListResponse
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
                $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
            }

            # GetIsuConditionResponse
            $formatted_condition = {
                jia_isu_uuid    => $last_condition->{jia_isu_uuid},
                isu_name        => $isu->{name},
                timestamp       => unix_from_mysql_datetime($last_condition->{timestamp}),
                is_sitting      => $last_condition->{is_sitting},
                condition       => $last_condition->{condition},
                condition_level => $condition_level,
                message         => $last_condition->{message},
            }
        }

        # GetIsuListResponse
        my $res = {
            id                   => $isu->{id},
            jia_isu_uuid         => $isu->{jia_isu_uuid},
            name                 => $isu->{name},
            character            => $isu->{character},
            latest_isu_condition => $formatted_condition,
        };
        push $response_list->@* => $res;
    }

    return $c->render_json($response_list, json_type_arrayof(GetIsuListResponse));
}

# POST /api/isu
# ISUを登録
sub post_isu($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);

    my $use_default_image = !!0;

    my $jia_isu_uuid = $c->req->parameters->{'jia_isu_uuid'};
    my $isu_name = $c->req->parameters->{'isu_name'};

    my $file = $c->req->uploads->{'image'};
    if (!$file) {
        $use_default_image = !!1;
    }

    my $image;
    if ($use_default_image) {
        $image = read_file(DEFAULT_ICON_FILE_PATH, binmode => ':raw');
    }
    else {
        $image = read_file($file->path, binmode => ':raw');
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope;
    my $isu;
    try {
        $dbh->query("INSERT INTO `isu` (`jia_isu_uuid`, `name`, `image`, `jia_user_id`) VALUES (?, ?, ?, ?)",
        $jia_isu_uuid, $isu_name, $image, $jia_user_id);

        my $target_url = $self->get_jia_service_url() . "/api/activate";
        my $body = {
            target_base_url => POST_ISUCONDITION_TARGET_BASE_URL,
            isu_uuid => $jia_isu_uuid,
        };

        my $furl = Furl->new;
        my $res = $furl->post($target_url,
            [ "Content-Type" => "application/json" ],
            encode_json($body),
        );

        if ($res->status != HTTP_ACCEPTED) {
            warnf("JIAService returned error: status code %s, message: %s", $res->status, $res->message);
            $c->halt_text($res->status, "JIAService returned error");
        }

        my $isu_from_jia = decode_json($res->body);
        $self->dbh->query("UPDATE `isu` SET `character` = ? WHERE  `jia_isu_uuid` = ?", $isu_from_jia->{character}, $jia_isu_uuid);

        $isu = $self->dbh->select_row(
            "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
            $jia_user_id, $jia_isu_uuid);

        $txn->commit;
    }
    catch ($e) {
        $txn->rollback;
        if ($e isa Kossy::Exception) {
            die $e; # rethrow
        }
        elsif ($e =~ /DBD::mysql::st execute failed: Duplicate entry/) {
            $c->halt_text(HTTP_CONFLICT, "duplicated: isu");
        }
        warnf("db error: %s", $e);
        $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
    }

    delete $isu->{image};
    return $c->render_json($isu, Isu, HTTP_CREATED);
}

# GET /api/isu/:jia_isu_uuid
# ISUの情報を取得
sub get_isu_id($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);

    my $jia_isu_uuid = $c->args->{'jia_isu_uuid'};

    my $isu = $self->dbh->select_row(
        "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
        $jia_user_id, $jia_isu_uuid);

    if (!$isu) {
        $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
    }

    delete $isu->{image};
    return $c->render_json($isu, Isu);
}

# GET /api/isu/:jia_isu_uuid/icon
# ISUのアイコンを取得
sub get_isu_icon($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);

    my $jia_isu_uuid = $c->args->{'jia_isu_uuid'};

    my $image = $self->dbh->select_one(
        "SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
        $jia_user_id, $jia_isu_uuid);

    if (!$image) {
        $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
    }

    $c->res->status(HTTP_OK);
    $c->res->body($image);
    return $c->res;
}

# GET /api/isu/:jia_isu_uuid/graph
# ISUのコンディショングラフ描画のための情報を取得
sub get_isu_graph($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);

    my $jia_isu_uuid = $c->args->{'jia_isu_uuid'};
    my $datetime = $c->req->parameters->{'datetime'};
    if (!$datetime) {
        $c->halt_text(HTTP_BAD_REQUEST, "missing: datetime");
    }
    if (!looks_like_number($datetime)) {
        $c->halt_text(HTTP_BAD_REQUEST, "bad format: datetime");
    }
    my $date = tm_from_unix($datetime)->strftime('%Y-%m-%d %H:00:00');

    my $count = $self->dbh->select_one(
        "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
        $jia_user_id, $jia_isu_uuid);

    if ($count == 0) {
        return $c->halt_text(HTTP_NOT_FOUND, "not found: isu")
    }

    my $res = $self->generate_isu_graph_response($jia_isu_uuid, $date);
    return $c->render_json($res, json_type_arrayof(GraphResponse));
}

# グラフのデータ点を一日分生成
sub generate_isu_graph_response($self, $jia_isu_uuid, $graph_date) {
    my $data_points = [];
    my $conditions_in_this_hour = [];
    my $timestamps_in_this_hour = [];
    my $start_time_in_this_hour = '';
    my $condition = {};

    my $rows = $self->dbh->select_all(
        "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC", $jia_isu_uuid);

    for my $condition ($rows->@*) {
        my $truncated_condition_time = tm_from_mysql_datetime($condition->{timestamp})->strftime('%Y-%m-%d %H:00:00');

        if ($truncated_condition_time ne $start_time_in_this_hour) {
            if ($conditions_in_this_hour->@* > 0) {
                my $data = calculate_graph_data_point($conditions_in_this_hour);

                push $data_points->@* => {
                    jia_isu_uuid         => $jia_isu_uuid,
                    start_at             => $start_time_in_this_hour,
                    data                 => $data,
                    condition_timestamps => $timestamps_in_this_hour,
                };
            }

            $start_time_in_this_hour = $truncated_condition_time;
            $conditions_in_this_hour = [];
            $timestamps_in_this_hour = [];
        }
        push $conditions_in_this_hour->@*, $condition;
        push $timestamps_in_this_hour->@*, unix_from_mysql_datetime($condition->{timestamp});
    }

    if ($conditions_in_this_hour->@* > 0) {
        my $data = calculate_graph_data_point($conditions_in_this_hour);
        push $data_points->@* => {
            jia_isu_uuid         => $jia_isu_uuid,
            start_at             => $start_time_in_this_hour,
            data                 => $data,
            condition_timestamps => $timestamps_in_this_hour,
        };
    }

    my $tm_graph_date  = tm_from_mysql_datetime($graph_date);
    my $tm_end_time    = $tm_graph_date->plus_hours(24);
    my $start_index    = $data_points->@*;
    my $end_next_index = $data_points->@*;

    for (my $i = 0; $i < $data_points->@*; $i++) {
        my $graph       = $data_points->[$i];
        my $tm_start_at = tm_from_mysql_datetime($graph->{start_at});

        if ($start_index == $data_points->@* && !($tm_start_at < $tm_graph_date)) {
            $start_index = $i;
        }
        if ($end_next_index == $data_points->@* and $tm_start_at > $tm_end_time) {
            $end_next_index = $i;
        }
    }

    my $filtered_data_points = [];
    if ($start_index < $end_next_index) {
        $filtered_data_points = [ $data_points->@[$start_index .. $end_next_index - 1] ];
    }

    my $response_list = []; # GraphResponse
    my $index = 0;
    my $tm_this_time = $tm_graph_date;

    while ($tm_this_time < $tm_graph_date->plus_hours(24)) {
        my $data;
        my $timestamps = [];
        my $this_time = mysql_datetime($tm_this_time);

        if ($index < $filtered_data_points->@*) {
            my $data_with_info = $filtered_data_points->[$index];
            if ($data_with_info->{start_at} eq $this_time) {
                $data       = $data_with_info->{data};
                $timestamps = $data_with_info->{condition_timestamps};
                $index++;
            }
        }

        push $response_list->@* => {
            start_at             => $tm_this_time->epoch,
            end_at               => $tm_this_time->plus_hours(1)->epoch,
            data                 => $data,
            condition_timestamps => $timestamps,
        };

        $tm_this_time = $tm_this_time->plus_hours(1);
    }

    return $response_list;
}

# 複数のISUのコンディションからグラフの一つのデータ点を計算
sub calculate_graph_data_point($isu_conditions) {
    my $conditions_count = { is_broken => 0, is_dirty => 0, is_overweight => 0 };

    my $raw_score = 0;
    for my $condition ($isu_conditions->@*) {
        my $bad_conditions_count = 0;

        if (!is_valid_condition_format($condition->{condition})) {
            critf("invalid condition format");
        }

        for my $cond_str (split /,/, $condition->{condition}) {
            my ($condition_name, $value) = split /=/, $cond_str;
            if ($value eq "true") {
                $conditions_count->{$condition_name} += 1;
                $bad_conditions_count++;
            }
        }

        if ($bad_conditions_count >= 3) {
            $raw_score += SCORE_CONDITION_LEVEL_CRITICAL;
        }
        elsif ($bad_conditions_count >= 1) {
            $raw_score += SCORE_CONDITION_LEVEL_WARNING;
        }
        else {
            $raw_score += SCORE_CONDITION_LEVEL_INFO;
        }
    }

    my $sitting_count = 0;
    for my $condition ($isu_conditions->@*) {
        if ($condition->{is_sitting}) {
            $sitting_count++;
        }
    }

    my $isu_conditions_length = $isu_conditions->@*;

    my $score = $raw_score * 100 / 3 / $isu_conditions_length;

    my $sitting_percentage       = $sitting_count * 100 / $isu_conditions_length;
    my $is_broken_percentage     = $conditions_count->{"is_broken"} * 100 / $isu_conditions_length;
    my $is_overweight_percentage = $conditions_count->{"is_overweight"} * 100 / $isu_conditions_length;
    my $is_dirty_percentage      = $conditions_count->{"is_dirty"} * 100 / $isu_conditions_length;

    my $data_point = { # GraphDataPoint
        score => $score,
        percentage => {
            sitting       => $sitting_percentage,
            is_broken     => $is_broken_percentage,
            is_overweight => $is_overweight_percentage,
            is_dirty      => $is_dirty_percentage,
        },
    };
    return $data_point;
}

# GET /api/condition/:jia_isu_uuid
# ISUのコンディションを取得
sub get_isu_conditions($self, $c) {
    my $jia_user_id = $self->get_user_id_from_session($c);
    my $jia_isu_uuid = $c->args->{'jia_isu_uuid'};
    if (!$jia_isu_uuid) {
        $c->halt_text(HTTP_BAD_REQUEST, "missing: jia_isu_uuid");
    }

    my $end_time = $c->req->parameters->{'end_time'};
    if (!looks_like_number($end_time)) {
        $c->halt_text(HTTP_BAD_REQUEST, "bad format: end_time");
    }
    $end_time = mysql_datetime_from_unix($end_time);

    my $condition_level_csv = $c->req->parameters->{'condition_level'};
    if (!$condition_level_csv) {
        $c->halt_text(HTTP_BAD_REQUEST, "missing: condition_level");
    }
    my $condition_level = {};
    for my $level (split /,/, $condition_level_csv) {
        $condition_level->{$level} = {};
    }

    my $start_time = $c->req->parameters->{'start_time'};
    if ($start_time) {
        if (!looks_like_number($start_time)) {
            $c->halt_text(HTTP_BAD_REQUEST, "bad format: start_time");
        }
        $start_time = mysql_datetime_from_unix($start_time);
    }

    my $isu_name = $self->dbh->select_one(
        "SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ?",
        $jia_isu_uuid, $jia_user_id);

    if (!$isu_name) {
        $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
    }

    my $conditions_response = $self->get_isu_conditions_from_db(
        $jia_isu_uuid,
        $end_time,
        $condition_level,
        $start_time,
        CONDITION_LIMIT,
        $isu_name,
    );

    return $c->render_json($conditions_response, json_type_arrayof(GetIsuConditionResponse));
}

# ISUのコンディションをDBから取得
sub get_isu_conditions_from_db($self, $jia_isu_uuid, $end_time, $condition_level, $start_time, $limit, $isu_name) {
    my $conditions;
    if (!$start_time) {
        $conditions = $self->dbh->select_all(
            "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?".
                "    AND `timestamp` < ?".
                "    ORDER BY `timestamp` DESC",
            $jia_isu_uuid, $end_time,
        )
    }
    else {
        $conditions = $self->dbh->select_all(
            "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?".
                "    AND `timestamp` < ?".
                "    AND ? <= `timestamp`".
                "    ORDER BY `timestamp` DESC",
            $jia_isu_uuid, $end_time, $start_time,
        )
    }

    my $conditions_response = [];
    for my $c ($conditions->@*) {
        my ($c_level, $e) = calculate_condition_level($c->{condition});
        if ($e) {
            next;
        }

        if ($condition_level->{$c_level}) {

            # GetIsuConditionResponse
            push $conditions_response->@*, {
                jia_isu_uuid    => $c->{jia_isu_uuid},
                isu_name        => $isu_name,
                timestamp       => unix_from_mysql_datetime($c->{timestamp}),
                is_sitting      => $c->{is_sitting},
                condition       => $c->{condition},
                condition_level => $c_level,
                message         => $c->{message},
            };
        }
    }

    if ($conditions_response->@* > $limit) {
        $conditions_response = [ splice $conditions_response->@*, 0, $limit ];
    }
    return $conditions_response;
}

# ISUのコンディションの文字列からコンディションレベルを計算
sub calculate_condition_level($condition) {
    my $warn_count = () = $condition =~ m!=true!g;

    my $condition_level;
    if ($warn_count == 0) {
        $condition_level = CONDITION_LEVEL_INFO;
    }
    elsif($warn_count == 1 || $warn_count == 2) {
        $condition_level = CONDITION_LEVEL_WARNING;
    }
    elsif ($warn_count == 3) {
        $condition_level = CONDITION_LEVEL_CRITICAL;
    }
    else {
        return (undef, "unexpected warn_count");
    }
    return ($condition_level, undef);
}

# GET /api/trend
# ISUの性格毎の最新のコンディション情報
sub get_trend($self, $c) {
    my $character_list = $self->dbh->select_all(
        "SELECT `character` FROM `isu` GROUP BY `character`");

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
            );

            if ($conditions->@* > 0) {
                my $isu_last_condition = $conditions->[0];
                my ($condition_level, $e) = calculate_condition_level($isu_last_condition->{condition});
                if ($e) {
                    warnf($e);
                    $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
                }
                my $trend_condition = {
                    isu_id    => $isu->{id},
                    timestamp => unix_from_mysql_datetime($isu_last_condition->{timestamp}),
                };
                push $character_isu_conditions->{$condition_level}->@* => $trend_condition;
            }
        }

        for my $level (keys $character_isu_conditions->%*) {
            my $conditions = $character_isu_conditions->{$level};
            my @sorted_conditions = sort { $b->{timestamp} <=> $a->{timestamp} } $conditions->@*;
            $character_isu_conditions->{$level} = \@sorted_conditions;
        }

        push $trend_response->@* => {
            character => $character->{character},
            info      => $character_isu_conditions->{info}     // [],
            warning   => $character_isu_conditions->{warning}  // [],
            critical  => $character_isu_conditions->{critical} // [],
        };
    }

    return $c->render_json($trend_response, json_type_arrayof(TrendResponse));
}

# POST /api/condition/:jia_isu_uuid
# ISUからのコンディションを受け取る
sub post_isu_condition($self, $c) {
    # TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき
    my $drop_probability = 0.9;
    if (rand() <= $drop_probability) {
        warnf("drop post isu condition request");
        $c->halt_no_content(HTTP_ACCEPTED);
    }

    my $jia_isu_uuid = $c->args->{'jia_isu_uuid'};
    if (!$jia_isu_uuid) {
        $c->halt_text(HTTP_BAD_REQUEST, "missing: jia_isu_uuid");
    }

    my $req = decode_json($c->req->content, json_type_arrayof(PostIsuConditionRequest));

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope;
    try {
        my $count = $self->dbh->select_one("SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?", $jia_isu_uuid);
        if ($count == 0) {
            $c->halt_text(HTTP_NOT_FOUND, "not found: isu");
        }

        for my $cond ($req->@*) {
            my $timestamp = mysql_datetime_from_unix($cond->{timestamp});

            if (!is_valid_condition_format($cond->{condition})) {
                $c->halt_text(HTTP_BAD_REQUEST, "bad request body");
            }

            $dbh->query(
                "INSERT INTO `isu_condition`" .
                "    (`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)".
                "    VALUES (?, ?, ?, ?, ?)",
                $jia_isu_uuid, $timestamp, $cond->{is_sitting}, $cond->{condition}, $cond->{message}
            );
        };

        $txn->commit;
    }
    catch ($e) {
        $txn->rollback;
        if ($e isa Kossy::Exception) {
            die $e; # rethrow
        }
        warnf("db error: %s", $e);
        $c->halt_no_content(HTTP_INTERNAL_SERVER_ERROR);
    }

    return $c->halt_no_content(HTTP_ACCEPTED);
}

# ISUのコンディションの文字列がcsv形式になっているか検証
sub is_valid_condition_format($condition_str) {

    my $keys = ["is_dirty=", "is_overweight=", "is_broken="];
    my $value_true = "true";
    my $value_false = "false";

    my $idx_cond_str = 0;

    for (my $idx_keys = 0; $idx_keys < $keys->@*; $idx_keys++) {
        my $key = $keys->[$idx_keys];

        if (index($condition_str, $key, $idx_cond_str) != $idx_cond_str) {
            return !!0;
        }
        $idx_cond_str += length $key;

        if (index($condition_str, $value_true, $idx_cond_str) == $idx_cond_str) {
            $idx_cond_str += length $value_true;
        }
        elsif (index($condition_str, $value_false, $idx_cond_str) == $idx_cond_str) {
            $idx_cond_str += length $value_false;
        }
        else {
            return !!0;
        }

        if ($idx_keys < $keys->@* - 1) {
            if (index($condition_str, ",", $idx_cond_str) != $idx_cond_str) {
                return !!0;
            }
            $idx_cond_str++;
        }
    }

    return $idx_cond_str == length $condition_str;
}

sub get_index($self, $c) {
    my $file = FRONTEND_CONTENTS_PATH . "/index.html";
    my $html = read_file($file);
    $c->res->status(HTTP_OK);
    $c->res->content_type('text/html; charset=UTF-8');
    $c->res->body($html);
    $c->res;
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
        my ($host, $port, $user, $dbname, $password) = MYSQL_CONNECTION_ENV->@{qw/host port user dbname password/};
        my $dsn = "dbi:mysql:database=$dbname;host=$host;port=$port";
        DBIx::Sunny->connect($dsn, $user, $password, {
            mysql_enable_utf8mb4 => 1,
            mysql_auto_reconnect => 1,
            Callbacks => {
                connected => sub {
                    my $dbh = shift;
                    return;
                },
                'connect_cached.connected' => sub {
                    shift->do('SET timezone = "Asia/Tokyo"');
                }
            },
        });
    };
}

sub unix_from_mysql_datetime {
    my $str = shift;
    my $tm = tm_from_mysql_datetime($str);
    return $tm->epoch;
}

sub mysql_datetime_from_unix {
    my $epoch = shift;
    my $tm = tm_from_unix($epoch);
    return mysql_datetime($tm)
}

sub mysql_datetime {
    my $tm = shift;
    return $tm->strftime("%Y-%m-%d %H:%M:%S");
}

sub tm_from_mysql_datetime {
    my $str = shift;
    return Time::Moment->from_string($str.'+9', lenient => 1);
}

sub tm_from_unix {
    my $epoch = shift;
    return Time::Moment->from_epoch($epoch)->with_offset_same_instant(9*60);
}

# XXX hack Kossy
{
    use Plack::Session;

    no warnings qw(redefine);
    my $orig = \&Kossy::Exception::response;
    *Kossy::Exception::response = sub {
        my $self = $_[0];
        if ($self->{my_response} isa Plack::Response) {
            return $self->{my_response}->finalize;
        }
        goto $orig;
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

    *Kossy::Connection::session = sub {
        my $c = shift;
        Plack::Session->new($c->env);
    };

    # override
    my $_JSON = JSON::MaybeXS->new()->allow_blessed(1)->convert_blessed(1)->ascii(0);
    *Kossy::Connection::render_json = sub {
        my ($c, $obj, $json_spec, $status) = @_;

        my $body = $_JSON->encode($obj, $json_spec); # Cpanel::JSON::XS::Typeを利用する
        $body = $c->escape_json($body);

        $status //= 200;

        $c->res->status( $status );
        $c->res->content_type('application/json; charset=UTF-8');
        $c->res->header( 'X-Content-Type-Options' => 'nosniff' ); # defense from XSS
        $c->res->body( $body );
        $c->res;
    };
}

1
