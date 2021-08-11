from os import getenv
from subprocess import call
from flask import Flask, request, session, send_file
from werkzeug.exceptions import BadRequest, Unauthorized
import mysql.connector
from sqlalchemy.pool import QueuePool
import jwt

app = Flask(__name__, static_folder="../public/assets", static_url_path="/assets")
app.secret_key = getenv("SESSION_KEY", "isucondition")

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


with open("../ec256-public.pem", "rb") as f:
    jwt_public_key = f.read()


def get_user_id_from_session():
    jia_user_id = session.get("jia_user_id")

    if jia_user_id is None:
        raise Unauthorized("no session")

    query = "SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = %s"
    (count,) = select_row(query, (jia_user_id,), dictionary=False)

    if count == 0:
        raise Unauthorized("not found: user")

    return jia_user_id


@app.route("/initialize", methods=["POST"])
def post_initialize():
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
    get_user_id_from_session()
    session.clear()
    return ""


@app.route("/api/user/me", methods=["GET"])
def get_me():
    jia_user_id = get_user_id_from_session()
    return {"jia_user_id": jia_user_id}


@app.route("/api/isu", methods=["GET"])
def get_isu_list():
    raise NotImplementedError


@app.route("/api/isu", methods=["POST"])
def post_isu():
    raise NotImplementedError


@app.route("/api/isu/:jia_isu_uuid", methods=["GET"])
def get_isu_id():
    raise NotImplementedError


@app.route("/api/isu/:jia_isu_uuid/icon", methods=["GET"])
def get_isu_icon():
    raise NotImplementedError


@app.route("/api/isu/:jia_isu_uuid/graph", methods=["GET"])
def get_isu_graph():
    raise NotImplementedError


@app.route("/api/condition/:jia_isu_uuid", methods=["GET"])
def get_isu_confitions():
    raise NotImplementedError


@app.route("/api/trend", methods=["GET"])
def get_trend():
    raise NotImplementedError


@app.route("/api/condition/:jia_isu_uuid", methods=["POST"])
def post_isu_condition():
    raise NotImplementedError


@app.route("/", methods=["GET"])
def get_index():
    return send_file("../public/index.html")


@app.route("/condition", methods=["GET"])
def get_condition():
    return send_file("../public/index.html")


@app.route("/isu/:jia_isu_uuid", methods=["GET"])
def get_isu():
    return send_file("../public/index.html")


@app.route("/register", methods=["GET"])
def get_register():
    return send_file("../public/index.html")


@app.route("/login", methods=["GET"])
def get_login():
    return send_file("../public/index.html")


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=getenv("SERVER_APP_PORT", 3000), threaded=True)
