from os import getenv
from subprocess import call
from dataclasses import dataclass
import json
from datetime import datetime, timedelta
import urllib.request
from random import random
from enum import Enum
from flask import Flask, request, session, send_file, jsonify, abort, make_response
from flask.json import JSONEncoder
from werkzeug.exceptions import BadRequest, Unauthorized, NotFound, InternalServerError, ServiceUnavailable
import mysql.connector
from sqlalchemy.pool import QueuePool
import jwt


CONDITION_LIMIT = 20
FRONTEND_CONTENTS_PATH = "../public"
JIA_JWT_SIGNING_KEY_PATH = "../ec256-public.pem"
DEFAULT_ICON_FILE_PATH = "../NoImage.jpg"
DEFAULT_JIA_SERVICE_URL = "http://localhost:5000"
MYSQL_ERR_NUM_DUPLICATE_ENTRY = 1062


class CONDITION_LEVEL(str, Enum):
    INFO = "info"
    WARNING = "warning"
    CRITICAL = "critical"


class SCORE_CONDITION_LEVEL(int, Enum):
    INFO = 3
    WARNING = 2
    CRITICAL = 1


@dataclass
class Isu:
    id: int
    jia_isu_uuid: int
    name: str
    image: bytes
    character: str
    jia_user_id: str
    created_at: datetime
    updated_at: datetime


@dataclass
class IsuCondition:
    id: int
    jia_isu_uuid: str
    timestamp: datetime
    is_sitting: bool
    condition: str
    message: str
    created_at: datetime


@dataclass
class ConditionsPercentage:
    sitting: int
    is_broken: int
    is_dirty: int
    is_overweight: int


@dataclass
class GraphDataPoint:
    score: int
    percentage: ConditionsPercentage


@dataclass
class GraphDataPointWithInfo:
    jia_isu_uuid: str
    start_at: datetime
    data: GraphDataPoint
    condition_timestamps: list


@dataclass
class GraphResponse:
    start_at: int
    end_at: int
    data: GraphDataPoint
    condition_timestamps: list


@dataclass
class GetIsuConditionResponse:
    jia_isu_uuid: str
    isu_name: str
    timestamp: int
    is_sitting: bool
    condition: str
    condition_level: str
    message: str


@dataclass
class TrendCondition:
    isu_id: int
    timestamp: int


@dataclass
class TrendResponse:
    character: str
    info: TrendCondition
    warning: TrendCondition
    critical: TrendCondition


@dataclass
class PostIsuConditionRequest:
    is_sitting: bool
    condition: str
    message: str
    timestamp: int


class CustomJSONEncoder(JSONEncoder):
    def default(self, obj):
        if isinstance(obj, datetime):
            return obj.timestamp()
        if isinstance(obj, Isu):
            cols = ["id", "jia_isu_uuid", "name", "character"]
            return {col: obj.__dict__[col] for col in cols}
        return JSONEncoder.default(self, obj)


app = Flask(__name__, static_folder=f"{FRONTEND_CONTENTS_PATH}/assets", static_url_path="/assets")
app.session_cookie_name = "isucondition"
app.secret_key = getenv("SESSION_KEY", "isucondition")
app.json_encoder = CustomJSONEncoder

mysql_connection_env = {
    "host": getenv("MYSQL_HOST", "127.0.0.1"),
    "port": getenv("MYSQL_PORT", 3306),
    "user": getenv("MYSQL_USER", "isucon"),
    "password": getenv("MYSQL_PASS", "isucon"),
    "database": getenv("MYSQL_DBNAME", "isucondition"),
}

cnxpool = QueuePool(lambda: mysql.connector.connect(**mysql_connection_env), pool_size=10)


def select_all(query, *args, dictionary=True):
    cnx = cnxpool.connect()
    try:
        cur = cnx.cursor(dictionary=dictionary)
        cur.execute(query, *args)
        return cur.fetchall()
    finally:
        cnx.close()


def select_row(*args, **kwargs):
    rows = select_all(*args, **kwargs)
    return rows[0] if len(rows) > 0 else None


with open(JIA_JWT_SIGNING_KEY_PATH, "rb") as f:
    jwt_public_key = f.read()

post_isu_condition_target_base_url = getenv("POST_ISUCONDITION_TARGET_BASE_URL")
if post_isu_condition_target_base_url is None:
    raise Exception("missing: POST_ISUCONDITION_TARGET_BASE_URL")


def get_user_id_from_session():
    jia_user_id = session.get("jia_user_id")

    if jia_user_id is None:
        raise Unauthorized("no session")

    query = "SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = %s"
    (count,) = select_row(query, (jia_user_id,), dictionary=False)

    if count == 0:
        raise Unauthorized("not found: user")

    return jia_user_id


def get_jia_service_url() -> str:
    query = "SELECT * FROM `isu_association_config` WHERE `name` = %s"
    config = select_row(query, ("jia_service_url",))
    return config["url"] if config is not None else DEFAULT_JIA_SERVICE_URL


@app.route("/initialize", methods=["POST"])
def post_initialize():
    """サービスを初期化"""
    if "jia_service_url" not in request.json:
        raise BadRequest("bad request body")

    call("../sql/init.sh")

    cnx = cnxpool.connect()
    try:
        cur = cnx.cursor()
        query = "INSERT INTO `isu_association_config` (`name`, `url`) VALUES (%s, %s) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)"
        cur.execute(query, ("jia_service_url", request.json["jia_service_url"]))
        cnx.commit()
    finally:
        cnx.close()

    return {"Language": "python"}


@app.route("/api/auth", methods=["POST"])
def post_auth():
    """サインアップ・サインイン"""
    req_jwt = request.headers["Authorization"].removeprefix("Bearer ")
    req_jwt_header = jwt.get_unverified_header(req_jwt)
    req_jwt_payload = jwt.decode(req_jwt, jwt_public_key, algorithms=[req_jwt_header["alg"]])
    jia_user_id = req_jwt_payload["jia_user_id"]

    cnx = cnxpool.connect()
    try:
        cur = cnx.cursor()
        query = "INSERT IGNORE INTO user (`jia_user_id`) VALUES (%s)"
        cur.execute(query, (jia_user_id,))
        cnx.commit()
    finally:
        cnx.close()

    session["jia_user_id"] = jia_user_id

    return ""


@app.route("/api/signout", methods=["POST"])
def post_signout():
    """サインアウト"""
    get_user_id_from_session()
    session.clear()
    return ""


@app.route("/api/user/me", methods=["GET"])
def get_me():
    """サインインしている自分自身の情報を取得"""
    jia_user_id = get_user_id_from_session()
    return {"jia_user_id": jia_user_id}


@app.route("/api/isu", methods=["GET"])
def get_isu_list():
    """ISUの一覧を取得"""
    jia_user_id = get_user_id_from_session()

    query = "SELECT * FROM `isu` WHERE `jia_user_id` = %s ORDER BY `id` DESC"
    isu_list = select_all(query, (jia_user_id,))

    response_list = []
    for isu in isu_list:
        found_last_condition = True
        query = "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = %s ORDER BY `timestamp` DESC LIMIT 1"
        last_condition = select_row(query, (isu["jia_isu_uuid"],))
        if last_condition is None:
            found_last_condition = False

        formatted_condition = None
        if found_last_condition:
            condition_level = calculate_condition_level(last_condition["condition"])
            formatted_condition = {
                "jia_isu_uuid": last_condition["jia_isu_uuid"],
                "isu_name": isu["name"],
                "timestamp": last_condition["timestamp"].timestamp(),
                "condition": last_condition["condition"],
                "condition_level": condition_level,
                "message": last_condition["message"],
            }

        res = {
            "id": isu["id"],
            "jia_isu_uuid": isu["jia_isu_uuid"],
            "name": isu["name"],
            "character": isu["character"],
            "latest_isu_condition": formatted_condition,
        }
        response_list.append(res)

    return jsonify(response_list)


@app.route("/api/isu", methods=["POST"])
def post_isu():
    """ISUを登録"""
    jia_user_id = get_user_id_from_session()

    use_default_image = False

    jia_isu_uuid = request.form.get("jia_isu_uuid")
    isu_name = request.form.get("isu_name")
    image = request.files.get("image")

    if image is None:
        use_default_image = True

    if use_default_image:
        with open(DEFAULT_ICON_FILE_PATH, "rb") as f:
            image = f.read()
    else:
        image = image.read()

    cnx = cnxpool.connect()
    try:
        cnx.start_transaction()
        cur = cnx.cursor(dictionary=True)

        try:
            query = """
                INSERT
                INTO `isu` (`jia_isu_uuid`, `name`, `image`, `jia_user_id`)
                VALUES (%s, %s, %s, %s)
                """
            cur.execute(query, (jia_isu_uuid, isu_name, image, jia_user_id))
        except mysql.connector.errors.IntegrityError as e:
            if e.errno == MYSQL_ERR_NUM_DUPLICATE_ENTRY:
                abort(409, "duplicated: isu")
            raise

        target_url = f"{get_jia_service_url()}/api/activate"
        body = {
            "target_base_url": post_isu_condition_target_base_url,
            "isu_uuid": jia_isu_uuid,
        }
        headers = {
            "Content-Type": "application/json",
        }
        req_jia = urllib.request.Request(target_url, json.dumps(body).encode(), headers)
        try:
            with urllib.request.urlopen(req_jia) as res:
                isu_from_jia = json.load(res)
        except urllib.error.HTTPError as e:
            app.logger.error(f"JIAService returned error: status code {e.code}, message: {e.reason}")
            abort(e.code, "JIAService returned error")
        except urllib.error.URLError as e:
            app.logger.error(f"failed to request to JIAService: {e.reason}")
            raise InternalServerError

        query = "UPDATE `isu` SET `character` = %s WHERE  `jia_isu_uuid` = %s"
        cur.execute(query, (isu_from_jia["character"], jia_isu_uuid))

        query = "SELECT * FROM `isu` WHERE `jia_user_id` = %s AND `jia_isu_uuid` = %s"
        cur.execute(query, (jia_user_id, jia_isu_uuid))
        isu = Isu(**cur.fetchone())

        cnx.commit()
    except:
        cnx.rollback()
        raise
    finally:
        cnx.close()

    return jsonify(isu), 201


@app.route("/api/isu/<jia_isu_uuid>", methods=["GET"])
def get_isu_id(jia_isu_uuid):
    """ISUの情報を取得"""
    jia_user_id = get_user_id_from_session()

    query = "SELECT * FROM `isu` WHERE `jia_user_id` = %s AND `jia_isu_uuid` = %s"
    res = select_row(query, (jia_user_id, jia_isu_uuid))
    if res is None:
        raise NotFound("not found: isu")

    return jsonify(Isu(**res))


@app.route("/api/isu/<jia_isu_uuid>/icon", methods=["GET"])
def get_isu_icon(jia_isu_uuid):
    """ISUのアイコンを取得"""
    jia_user_id = get_user_id_from_session()

    query = "SELECT `image` FROM `isu` WHERE `jia_user_id` = %s AND `jia_isu_uuid` = %s"
    res = select_row(query, (jia_user_id, jia_isu_uuid))
    if res is None:
        raise NotFound("not found: isu")

    return make_response(res["image"], 200, {"Content-Type": "image/jpeg"})


@app.route("/api/isu/<jia_isu_uuid>/graph", methods=["GET"])
def get_isu_graph(jia_isu_uuid):
    """ISUのコンディショングラフ描画のための情報を取得"""
    jia_user_id = get_user_id_from_session()

    dt = request.args.get("datetime")
    if dt is None:
        raise BadRequest("missing: datetime")
    try:
        dt = datetime.fromtimestamp(int(dt))
    except:
        raise BadRequest("bad format: datetime")
    dt = truncate_datetime(dt)

    query = "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = %s AND `jia_isu_uuid` = %s"
    (count,) = select_row(query, (jia_user_id, jia_isu_uuid))
    if count == 0:
        raise NotFound("not found: isu")

    res = generate_isu_graph_response(jia_isu_uuid, dt)
    return jsonify(res)


def truncate_datetime(dt: datetime) -> datetime:
    """datetime 値の時刻を 00:00:00 にする"""
    return datetime(dt.year, dt.month, dt.day)


def generate_isu_graph_response(jia_isu_uuid: str, graph_date: datetime) -> list[GraphResponse]:
    """グラフのデータ点を一日分生成"""
    data_points = []
    conditions_in_this_hour = []
    timestamps_in_this_hour = []
    start_time_in_this_hour = None

    query = "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = %s ORDER BY `timestamp` ASC"
    rows = select_all(query, (jia_isu_uuid,))
    for row in rows:
        condition = IsuCondition(**row)
        truncated_condition_time = truncate_datetime(condition.timestamp)
        if truncate_datetime != start_time_in_this_hour:
            if len(conditions_in_this_hour) > 0:
                data_points.append(
                    GraphDataPointWithInfo(
                        jia_isu_uuid=jia_isu_uuid,
                        start_at=start_time_in_this_hour,
                        data=calculate_graph_data_point(conditions_in_this_hour),
                        condition_timestamps=timestamps_in_this_hour,
                    )
                )
            start_time_in_this_hour = truncated_condition_time
            conditions_in_this_hour = []
            timestamps_in_this_hour = []
        conditions_in_this_hour.append(condition)
        timestamps_in_this_hour.append(condition.timestamp.timestamp())

    if len(conditions_in_this_hour) > 0:
        data_points.append(
            GraphDataPointWithInfo(
                jia_isu_uuid=jia_isu_uuid,
                start_at=start_time_in_this_hour,
                data=calculate_graph_data_point(conditions_in_this_hour),
                condition_timestamps=timestamps_in_this_hour,
            )
        )

    end_time = graph_date + timedelta(days=1)
    start_index = 0
    end_next_index = len(data_points)
    for i, graph in enumerate(data_points):
        if start_index == 0 and graph.start_at >= graph_date:
            start_index = i
        if end_next_index == len(data_points) and graph.start_at > end_time:
            end_next_index = 1

    filtered_data_points = []
    if start_index < end_next_index:
        filtered_data_points = data_points[start_index:end_next_index]

    response_list = []
    index = 0
    this_time = graph_date

    while this_time < graph_date + timedelta(days=1):
        data = None
        timestamps = []

        if index < len(filtered_data_points):
            data_with_info = filtered_data_points[index]

            if data_with_info.start_at == this_time:
                data = data_with_info.data
                timestamps = data_with_info.condition_timestamps
                index += 1

        response_list.append(
            GraphResponse(
                start_at=this_time.timestamp(),
                end_at=(this_time + timedelta(hours=1)).timestamp(),
                data=data,
                condition_timestamps=timestamps,
            )
        )

        this_time += timedelta(hours=1)

    return response_list


def calculate_graph_data_point(isu_conditions: list[IsuCondition]) -> GraphDataPoint:
    """複数のISUのコンディションからグラフの一つのデータ点を計算"""
    conditions_count = {"is_broken": 0, "is_dirty": 0, "is_overweight": 0}
    raw_score = 0
    for condition in isu_conditions:
        bad_conditions_count = 0

        if not is_valid_condition_format(condition.condition):
            raise Exception("invalid condition format")

        for cond_str in condition.condition.split(","):
            key_value = cond_str.split("=")

            condition_name = key_value[0]
            if key_value[1] == "true":
                conditions_count[condition_name] += 1
                bad_conditions_count += 1

        if bad_conditions_count >= 3:
            raw_score += SCORE_CONDITION_LEVEL.CRITICAL
        elif bad_conditions_count >= 1:
            raw_score += SCORE_CONDITION_LEVEL.WARNING
        else:
            raw_score += SCORE_CONDITION_LEVEL.INFO

    sitting_count = 0
    for condition in isu_conditions:
        if condition.is_sitting:
            sitting_count += 1

    isu_conditions_length = len(isu_conditions)

    score = raw_score / isu_conditions_length

    sitting_percentage = sitting_count * 100 / isu_conditions_length
    is_broken_percentage = conditions_count["is_broken"] * 100 / isu_conditions_length
    is_overweight_percentage = conditions_count["is_overweight"] * 100 / isu_conditions_length
    is_dirty_percentage = conditions_count["is_dirty"] * 100 / isu_conditions_length

    data_point = GraphDataPoint(
        score=score,
        percentage=ConditionsPercentage(
            sitting=int(sitting_percentage),
            is_broken=int(is_broken_percentage),
            is_overweight=int(is_overweight_percentage),
            is_dirty=int(is_dirty_percentage),
        ),
    )
    return data_point


@app.route("/api/condition/<jia_isu_uuid>", methods=["GET"])
def get_isu_confitions(jia_isu_uuid):
    """ISUのコンディションを取得"""
    jia_user_id = get_user_id_from_session()

    try:
        end_time = datetime.fromtimestamp(int(request.args.get("end_time")))
    except:
        raise BadRequest("bad format: end_time")

    condition_level_csv = request.args.get("condition_level")
    if condition_level_csv is None:
        raise BadRequest("missing: condition_level")
    condition_level = {level: None for level in condition_level_csv.split(",")}

    start_time_str = request.args.get("start_time")
    start_time = None
    if start_time_str is not None:
        try:
            start_time = datetime.fromtimestamp(int(start_time_str))
        except:
            raise BadRequest("bad format: start_time")

    query = "SELECT name FROM `isu` WHERE `jia_isu_uuid` = %s AND `jia_user_id` = %s"
    row = select_row(query, (jia_isu_uuid, jia_user_id))
    if row is None:
        raise NotFound("not found: isu")
    isu_name = row["name"]

    condition_response = get_isu_condition_from_db(
        jia_isu_uuid,
        end_time,
        condition_level,
        start_time,
        CONDITION_LIMIT,
        isu_name,
    )

    return jsonify(condition_response)


def get_isu_condition_from_db(
    jia_isu_uuid: str,
    end_time: datetime,
    condition_level: dict,
    start_time: datetime,
    limit: int,
    isu_name: str,
) -> list[GetIsuConditionResponse]:
    """ISUのコンディションをDBから取得"""
    if start_time is None:
        query = """
            SELECT *
            FROM `isu_condition`
            WHERE `jia_isu_uuid` = %s AND `timestamp` < %s
            ORDER BY `timestamp` DESC
            """
        conditions = [IsuCondition(**row) for row in select_all(query, (jia_isu_uuid, end_time))]
    else:
        query = """
            SELECT *
            FROM `isu_condition`
            WHERE `jia_isu_uuid` = %s AND `timestamp` < %s AND %s <= `timestamp`
            ORDER BY `timestamp` DESC
            """
        conditions = [IsuCondition(**row) for row in select_all(query, (jia_isu_uuid, end_time, start_time))]

    condition_response = []
    for c in conditions:
        try:
            c_level = calculate_condition_level(c.condition)
        except:
            continue

        if condition_level[c_level]:
            condition_response.append(
                GetIsuConditionResponse(
                    jia_isu_uuid=jia_isu_uuid,
                    isu_name=isu_name,
                    timestamp=c.timestamp.timestamp(),
                    is_sitting=c.is_sitting,
                    condition=c.condition,
                    condition_level=c_level,
                    message=c.message,
                )
            )

    if len(condition_response) > limit:
        condition_response = condition_response[:limit]

    return condition_response


@app.route("/api/trend", methods=["GET"])
def get_trend():
    """ISUの性格毎の最新のコンディション情報"""
    query = "SELECT `character` FROM `isu` GROUP BY `character`"
    character_list = [row["character"] for row in select_all(query)]

    res = []

    for character in character_list:
        query = "SELECT * FROM `isu` WHERE `character` = %s"
        isu_list = [Isu(**row) for row in select_all(query, (character,))]

        character_info_isu_conditions = []
        character_warning_isu_conditions = []
        character_critical_isu_conditions = []
        for isu in isu_list:
            query = "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = %s ORDER BY timestamp DESC"
            conditions = [IsuCondition(**row) for row in select_all(query, (isu.jia_isu_uuid,))]

            if len(conditions) > 0:
                isu_last_condition = conditions[0]
                condition_level = calculate_condition_level(isu_last_condition.condition)

                trend_condition = TrendCondition(isu_id=isu.id, timestamp=isu_last_condition.timestamp.timestamp())

                if condition_level == "info":
                    character_info_isu_conditions.append(trend_condition)
                elif condition_level == "warning":
                    character_warning_isu_conditions.append(trend_condition)
                elif condition_level == "critical":
                    character_critical_isu_conditions.append(trend_condition)

        character_info_isu_conditions.sort(key=lambda c: c.timestamp)
        character_warning_isu_conditions.sort(key=lambda c: c.timestamp)
        character_critical_isu_conditions.sort(key=lambda c: c.timestamp)

        res.append(
            TrendResponse(
                character=character,
                info=character_info_isu_conditions,
                warning=character_warning_isu_conditions,
                critical=character_critical_isu_conditions,
            )
        )

    return jsonify(res)


@app.route("/api/condition/<jia_isu_uuid>", methods=["POST"])
def post_isu_condition(jia_isu_uuid):
    """ISUからのコンディションを受け取る"""
    # TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき
    drop_probability = 0.9
    if random() <= drop_probability:
        app.logger.warning("drop post isu condition request")
        raise ServiceUnavailable

    try:
        req = [PostIsuConditionRequest(**row) for row in request.json]
    except:
        raise BadRequest("bad request body")

    cnx = cnxpool.connect()
    try:
        cnx.start_transaction()
        cur = cnx.cursor(dictionary=True)

        query = "SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = %s"
        cur.execute(query, (jia_isu_uuid,))
        (count,) = cur.fetchone()
        if count == 0:
            raise NotFound("not found: isu")

        for cond in req:
            if not is_valid_condition_format(cond.condition):
                raise BadRequest("bad request body")

            query = """
                INSERT
                INTO `isu_condition` (`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)
                VALUES (%s, %s, %s, %s, %s)
                """
            cur.execute(
                query,
                (
                    jia_isu_uuid,
                    datetime.fromtimestamp(cond.timestamp),
                    cond.is_sitting,
                    cond.condition,
                    cond.message,
                ),
            )

        cnx.commit()
    except:
        cnx.rollback()
        raise
    finally:
        cnx.close()

    return "", 201


def get_index(**kwargs):
    return send_file(f"{FRONTEND_CONTENTS_PATH}/index.html")


app.add_url_rule("/", view_func=get_index)
app.add_url_rule("/isu/<jia_isu_uuid>", view_func=get_index)
app.add_url_rule("/isu/<jia_isu_uuid>/condition", view_func=get_index)
app.add_url_rule("/isu/<jia_isu_uuid>/graph", view_func=get_index)
app.add_url_rule("/register", view_func=get_index)


def calculate_condition_level(condition: str) -> CONDITION_LEVEL:
    """ISUのコンディションの文字列からコンディションレベルを計算"""
    warn_count = condition.count("=true")

    if warn_count == 0:
        condition_level = CONDITION_LEVEL.INFO
    elif warn_count in (1, 2):
        condition_level = CONDITION_LEVEL.WARNING
    elif warn_count == 3:
        condition_level = CONDITION_LEVEL.CRITICAL
    else:
        raise Exception("unexpected warn count")

    return condition_level


def is_valid_condition_format(condition_str: str) -> bool:
    """ISUのコンディションの文字列がcsv形式になっているか検証"""
    keys = ["is_dirty=", "is_overweight=", "is_broken="]
    value_true = "true"
    value_false = "false"

    idx_cond_str = 0
    for idx_keys, key in enumerate(keys):
        if not condition_str[idx_cond_str:].startswith(key):
            return False
        idx_cond_str += len(key)

        if condition_str[idx_cond_str:].startswith(value_true):
            idx_cond_str += len(value_true)
        elif condition_str[idx_cond_str:].startswith(value_false):
            idx_cond_str += len(value_false)
        else:
            return False

        if idx_keys < (len(keys) - 1):
            if condition_str[idx_cond_str] != ",":
                return False
            idx_cond_str += 1

    return idx_cond_str == len(condition_str)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=getenv("SERVER_APP_PORT", 3000), threaded=True)
