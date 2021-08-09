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


@app.route("/", methods=["GET"])
def get_index():
    return send_file("../public/index.html")


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


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=getenv("SERVER_APP_PORT", 3000), threaded=True)
