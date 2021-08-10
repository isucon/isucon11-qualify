from os import getenv
from subprocess import call
from flask import Flask, request, send_file
from werkzeug.exceptions import BadRequest
import mysql.connector
from sqlalchemy.pool import QueuePool

app = Flask(__name__, static_folder="../public/assets", static_url_path="/assets")

mysql_connection_env = {
    "host": getenv("MYSQL_HOST", "127.0.0.1"),
    "port": getenv("MYSQL_PORT", 3306),
    "user": getenv("MYSQL_USER", "isucon"),
    "password": getenv("MYSQL_PASS", "isucon"),
    "database": getenv("MYSQL_DBNAME", "isucondition"),
}

cnxpool = QueuePool(lambda: mysql.connector.connect(**mysql_connection_env), pool_size=10)


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
    except Exception:
        raise  # TODO raise 500
    finally:
        cnx.close()

    return {"Language": "python"}


@app.route("/api/auth", methods=["POST"])
def post_auth():
    raise NotImplementedError


@app.route("/api/signout", methods=["POST"])
def post_signout():
    raise NotImplementedError


@app.route("/api/user/me", methods=["GET"])
def get_me():
    raise NotImplementedError


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
