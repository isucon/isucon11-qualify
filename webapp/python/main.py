from os import getenv
from subprocess import call
from dataclasses import dataclass
import json
from datetime import datetime
import urllib.request
from enum import Enum
from flask import Flask, request, session, send_file, jsonify, abort, make_response
from flask.json import JSONEncoder
from werkzeug.exceptions import BadRequest, Unauthorized, NotFound, InternalServerError
import mysql.connector
from sqlalchemy.pool import QueuePool
import jwt


class CONDITION_LEVEL(str, Enum):
    INFO = "info"
    WARNING = "warning"
    CRITICAL = "critical"


JIA_JWT_SIGNING_KEY_PATH = "../ec256-public.pem"
DEFAULT_ICON_FILE_PATH = "../NoImage.jpg"
DEFAULT_JIA_SERVICE_URL = "http://localhost:5000"


class CustomJSONEncoder(JSONEncoder):
    def default(self, obj):
        if isinstance(obj, datetime):
            return obj.timestamp()
        if isinstance(obj, Isu):
            cols = ["id", "jia_isu_uuid", "name", "character"]
            return {col: obj.__dict__[col] for col in cols}
        return JSONEncoder.default(self, obj)


app = Flask(__name__, static_folder="../public/assets", static_url_path="/assets")
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
    req_jwt = request.headers["Authorization"][len("Bearer ") :]
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
            if e.errno == 1062:
                abort(409, "duplicated: isu")
            raise

        target_url = get_jia_service_url() + "/api/activate"
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
    raise NotImplementedError


@app.route("/api/condition/<jia_isu_uuid>", methods=["GET"])
def get_isu_confitions(jia_isu_uuid):
    """ISUのコンディションを取得"""
    raise NotImplementedError


@app.route("/api/trend", methods=["GET"])
def get_trend():
    """ISUの性格毎の最新のコンディション情報"""
    raise NotImplementedError


@app.route("/api/condition/<jia_isu_uuid>", methods=["POST"])
def post_isu_condition(jia_isu_uuid):
    """ISUからのコンディションを受け取る"""
    raise NotImplementedError


@app.route("/", methods=["GET"])
def get_index():
    return send_file("../public/index.html")


@app.route("/condition", methods=["GET"])
def get_condition():
    return send_file("../public/index.html")


@app.route("/isu/<jia_isu_uuid>", methods=["GET"])
def get_isu(jia_isu_uuid):
    return send_file("../public/index.html")


@app.route("/register", methods=["GET"])
def get_register():
    return send_file("../public/index.html")


@app.route("/login", methods=["GET"])
def get_login():
    return send_file("../public/index.html")


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


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=getenv("SERVER_APP_PORT", 3000), threaded=True)
