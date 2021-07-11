use actix_web::{web, HttpResponse};
use chrono::Offset as _;
use chrono::TimeZone as _;
use chrono::{DateTime, NaiveDateTime};
use futures::StreamExt as _;
use futures::TryStreamExt as _;
use std::collections::{HashMap, HashSet};

const CONDITION_LIMIT: i64 = 20;
const FRONTEND_CONTENTS_PATH: &str = "../public";
const JWT_VERIFICATION_KEY_PATH: &str = "../ec256-public.pem";
const DEFAULT_ICON_FILE_PATH: &str = "../NoImage.jpg";
const DEFAULT_JIA_SERVICE_URL: &str = "http://localhost:5000";
const MYSQL_ERR_NUM_DUPLICATE_ENTRY: u16 = 1062;
const CONDITION_LEVEL_INFO: &str = "info";
const CONDITION_LEVEL_WARNING: &str = "warning";
const CONDITION_LEVEL_CRITICAL: &str = "critical";
const SCORE_CONDITION_LEVEL_INFO: i64 = 3;
const SCORE_CONDITION_LEVEL_WARNING: i64 = 2;
const SCORE_CONDITION_LEVEL_CRITICAL: i64 = 1;

lazy_static::lazy_static! {
    static ref TEMPLATES: Vec<u8> = std::fs::read(std::path::Path::new(FRONTEND_CONTENTS_PATH).join("index.html")).expect("Failed to load index.html");

    static ref JWT_VERIFICATION_KEY_PEM: Vec<u8> = std::fs::read(JWT_VERIFICATION_KEY_PATH).expect("failed to read JWT verification key file");
    static ref JWT_VERIFICATION_KEY: jsonwebtoken::DecodingKey<'static> = jsonwebtoken::DecodingKey::from_ec_pem(&JWT_VERIFICATION_KEY_PEM).expect("failed to load JWT verification key");

    static ref JST_OFFSET: chrono::FixedOffset = chrono::FixedOffset::east(9 * 60 * 60);
}

#[derive(Debug, sqlx::FromRow)]
struct Config {
    name: String,
    url: String,
}

#[derive(Debug, serde::Serialize)]
struct Isu {
    id: i64,
    jia_isu_uuid: String,
    name: String,
    #[serde(skip)]
    image: Vec<u8>,
    character: String,
    #[serde(skip)]
    jia_user_id: String,
    #[serde(skip)]
    created_at: DateTime<chrono::FixedOffset>,
    #[serde(skip)]
    updated_at: DateTime<chrono::FixedOffset>,
}
impl sqlx::FromRow<'_, sqlx::mysql::MySqlRow> for Isu {
    fn from_row(row: &sqlx::mysql::MySqlRow) -> sqlx::Result<Self> {
        use sqlx::Row as _;

        let created_at: NaiveDateTime = row.try_get("created_at")?;
        let updated_at: NaiveDateTime = row.try_get("updated_at")?;
        // DB の datetime 型は JST として解釈する
        let created_at = JST_OFFSET.from_local_datetime(&created_at).unwrap();
        let updated_at = JST_OFFSET.from_local_datetime(&updated_at).unwrap();
        Ok(Self {
            id: row.try_get("id")?,
            jia_isu_uuid: row.try_get("jia_isu_uuid")?,
            name: row.try_get("name")?,
            image: row.try_get("image")?,
            character: row.try_get("character")?,
            jia_user_id: row.try_get("jia_user_id")?,
            created_at,
            updated_at,
        })
    }
}

#[derive(Debug, serde::Serialize)]
struct GetIsuListResponse {
    id: i64,
    jia_isu_uuid: String,
    name: String,
    character: String,
    latest_isu_condition: Option<GetIsuConditionResponse>,
}

#[derive(Debug, serde::Deserialize)]
struct IsuFromJIA {
    character: String,
}

#[derive(Debug)]
struct IsuCondition {
    id: i64,
    jia_isu_uuid: String,
    timestamp: DateTime<chrono::FixedOffset>,
    is_sitting: bool,
    condition: String,
    message: String,
    created_at: DateTime<chrono::FixedOffset>,
}
impl sqlx::FromRow<'_, sqlx::mysql::MySqlRow> for IsuCondition {
    fn from_row(row: &sqlx::mysql::MySqlRow) -> sqlx::Result<Self> {
        use sqlx::Row as _;

        let timestamp: NaiveDateTime = row.try_get("timestamp")?;
        let created_at: NaiveDateTime = row.try_get("created_at")?;
        // DB の datetime 型は JST として解釈する
        let timestamp = JST_OFFSET.from_local_datetime(&timestamp).unwrap();
        let created_at = JST_OFFSET.from_local_datetime(&created_at).unwrap();
        Ok(Self {
            id: row.try_get("id")?,
            jia_isu_uuid: row.try_get("jia_isu_uuid")?,
            timestamp,
            is_sitting: row.try_get("is_sitting")?,
            condition: row.try_get("condition")?,
            message: row.try_get("message")?,
            created_at,
        })
    }
}

#[derive(Debug)]
struct MySQLConnectionEnv {
    host: String,
    port: u16,
    user: String,
    db_name: String,
    password: String,
}
impl Default for MySQLConnectionEnv {
    fn default() -> Self {
        let port = if let Ok(port) = std::env::var("MYSQL_PORT") {
            port.parse().unwrap_or(3306)
        } else {
            3306
        };
        Self {
            host: std::env::var("MYSQL_HOST").unwrap_or_else(|_| "127.0.0.1".to_owned()),
            port,
            user: std::env::var("MYSQL_USER").unwrap_or_else(|_| "isucon".to_owned()),
            db_name: std::env::var("MYSQL_DBNAME").unwrap_or_else(|_| "isucondition".to_owned()),
            password: std::env::var("MYSQL_PASS").unwrap_or_else(|_| "isucon".to_owned()),
        }
    }
}

#[derive(Debug, serde::Deserialize)]
struct InitializeRequest {
    jia_service_url: String,
}

#[derive(Debug, serde::Serialize)]
struct InitializeResponse {
    language: String,
}

#[derive(Debug, serde::Serialize)]
struct GetMeResponse {
    jia_user_id: String,
}

#[derive(Debug, serde::Deserialize)]
struct PutIsuRequest {
    name: String,
}

#[derive(Debug, serde::Serialize)]
struct GraphResponse {
    start_at: i64,
    end_at: i64,
    data: Option<GraphDataPoint>,
    condition_timestamps: Vec<i64>,
}

#[derive(Debug, serde::Serialize)]
struct GraphDataPoint {
    score: i64,
    percentage: ConditionsPercentage,
}

#[derive(Debug, serde::Serialize)]
struct ConditionsPercentage {
    sitting: i64,
    is_broken: i64,
    is_dirty: i64,
    is_overweight: i64,
}

#[derive(Debug)]
struct GraphDataPointWithInfo {
    jia_isu_uuid: String,
    start_at: DateTime<chrono::FixedOffset>,
    data: GraphDataPoint,
    condition_timestamps: Vec<i64>,
}

#[derive(Debug, serde::Serialize)]
struct GetIsuConditionResponse {
    jia_isu_uuid: String,
    isu_name: String,
    timestamp: i64,
    is_sitting: bool,
    condition: String,
    condition_level: &'static str,
    message: String,
}

#[derive(Debug, serde::Serialize)]
struct TrendResponse {
    character: String,
    conditions: Vec<TrendCondition>,
}

#[derive(Debug, serde::Serialize)]
struct TrendCondition {
    #[serde(rename = "isu_id")]
    id: i64,
    timestamp: i64,
    condition_level: &'static str,
}

#[derive(Debug, serde::Deserialize)]
struct PostIsuConditionRequest {
    is_sitting: bool,
    condition: String,
    message: String,
    timestamp: i64,
}

#[derive(Debug, serde::Serialize)]
struct JIAServiceRequest {
    target_ip: String,
    target_port: u16,
    isu_uuid: String,
}

struct IsuConditionAddress {
    public_address: String,
    public_port: u16,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or("info,sqlx=warn"))
        .init();
    let mysql_connection_env = MySQLConnectionEnv::default();

    let pool = sqlx::mysql::MySqlPoolOptions::new()
        .max_connections(10)
        .after_connect(|conn| {
            Box::pin(async move {
                use sqlx::Executor as _;
                // DB のタイムゾーンを JST に強制する
                conn.execute("set time_zone = '+09:00'").await?;
                Ok(())
            })
        })
        .connect_with(
            sqlx::mysql::MySqlConnectOptions::new()
                .host(&mysql_connection_env.host)
                .port(mysql_connection_env.port)
                .database(&mysql_connection_env.db_name)
                .username(&mysql_connection_env.user)
                .password(&mysql_connection_env.password),
        )
        .await
        .expect("DB connection failed");

    let isu_condition_public_address =
        std::env::var("SERVER_PUBLIC_ADDRESS").expect("env var SERVER_PUBLIC_ADDRESS is missing");
    let isu_condition_public_port = match std::env::var("SERVER_PUBLIC_PORT") {
        Ok(port_str) => port_str
            .parse()
            .expect("env var SERVER_PUBLIC_PORT is invalid"),
        Err(_) => 80,
    };

    let mut session_key = std::env::var("SESSION_KEY")
        .map(|k| k.into_bytes())
        .unwrap_or_else(|_| b"isucondition".to_vec());
    if session_key.len() < 32 {
        session_key.resize(32, 0);
    }

    let server = actix_web::HttpServer::new(move || {
        actix_web::App::new()
            .app_data(web::JsonConfig::default().error_handler(|err, _| {
                if matches!(err, actix_web::error::JsonPayloadError::Deserialize(_)) {
                    actix_web::error::ErrorBadRequest("bad request body")
                } else {
                    actix_web::error::ErrorBadRequest(err)
                }
            }))
            .app_data(web::Data::new(pool.clone()))
            .app_data(web::Data::new(IsuConditionAddress {
                public_address: isu_condition_public_address.clone(),
                public_port: isu_condition_public_port,
            }))
            .wrap(actix_web::middleware::Logger::default())
            .wrap(
                actix_session::CookieSession::signed(&session_key)
                    .secure(false)
                    .name("isucondition")
                    .max_age(2592000),
            )
            .service(post_initialize)
            .service(post_authentication)
            .service(post_signout)
            .service(get_me)
            .service(get_isu_list)
            .service(post_isu)
            .service(get_isu)
            .service(get_isu_icon)
            .service(get_isu_graph)
            .service(get_isu_conditions)
            .service(get_trend)
            .service(post_isu_condition)
            .route("/", web::get().to(get_index))
            .route("/condition", web::get().to(get_index))
            .route("/isu/{jia_isu_uuid:.*}", web::get().to(get_index))
            .route("/register", web::get().to(get_index))
            .route("/login", web::get().to(get_index))
            .service(actix_files::Files::new(
                "/assets",
                std::path::Path::new(FRONTEND_CONTENTS_PATH).join("assets"),
            ))
    });
    let server = if let Some(l) = listenfd::ListenFd::from_env().take_tcp_listener(0)? {
        server.listen(l)?
    } else {
        server.bind((
            "0.0.0.0",
            std::env::var("SERVER_APP_PORT")
                .map(|port_str| port_str.parse().expect("Failed to parse SERVER_APP_PORT"))
                .unwrap_or(3000),
        ))?
    };
    server.run().await
}

#[derive(Debug)]
struct SqlxError(sqlx::Error);
impl std::fmt::Display for SqlxError {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        self.0.fmt(f)
    }
}
impl actix_web::ResponseError for SqlxError {
    fn error_response(&self) -> HttpResponse {
        HttpResponse::InternalServerError()
            .content_type(mime::TEXT_PLAIN)
            .body(format!("SQLx error: {:?}", self.0))
    }
}

#[derive(Debug)]
struct ReqwestError(reqwest::Error);
impl std::fmt::Display for ReqwestError {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        self.0.fmt(f)
    }
}
impl actix_web::ResponseError for ReqwestError {
    fn error_response(&self) -> HttpResponse {
        HttpResponse::InternalServerError()
            .content_type(mime::TEXT_PLAIN)
            .body(format!("reqwest error: {:?}", self.0))
    }
}

async fn require_signed_in<'e, 'c, E>(
    executor: E,
    session: actix_session::Session,
) -> actix_web::Result<String>
where
    'c: 'e,
    E: 'e + sqlx::Executor<'c, Database = sqlx::MySql>,
{
    if let Some(jia_user_id) = session.get("jia_user_id")? {
        let count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = ?")
            .bind(&jia_user_id)
            .fetch_one(executor)
            .await
            .map_err(SqlxError)?;
        if count == 0 {
            todo!()
        } else {
            Ok(jia_user_id)
        }
    } else {
        let resp = actix_web::HttpResponse::Unauthorized()
            .content_type(mime::TEXT_PLAIN)
            .body("you are not signed in");
        Err(actix_web::error::InternalError::from_response("", resp).into())
    }
}

async fn get_jia_service_url<'e, 'c, E>(executor: E) -> sqlx::Result<String>
where
    'c: 'e,
    E: 'e + sqlx::Executor<'c, Database = sqlx::MySql>,
{
    let config: Option<Config> =
        sqlx::query_as("SELECT * FROM `isu_association_config` WHERE `name` = ?")
            .bind("jia_service_url")
            .fetch_optional(executor)
            .await?;
    Ok(config
        .map(|c| c.url)
        .unwrap_or_else(|| DEFAULT_JIA_SERVICE_URL.to_owned()))
}

async fn get_index() -> actix_web::Result<HttpResponse> {
    Ok(HttpResponse::Ok()
        .content_type(mime::TEXT_PLAIN)
        .body(TEMPLATES.clone()))
}

#[actix_web::post("/initialize")]
async fn post_initialize(
    pool: web::Data<sqlx::MySqlPool>,
    request: web::Json<InitializeRequest>,
) -> actix_web::Result<HttpResponse> {
    let status = tokio::process::Command::new("../sql/init.sh")
        .status()
        .await
        .map_err(|e| {
            log::error!("exec init.sh error: {}", e);
            e
        })?;
    if !status.success() {
        log::error!("exec init.sh failed with exit code {:?}", status.code());
        return Ok(HttpResponse::InternalServerError()
            .content_type(mime::TEXT_PLAIN)
            .body(""));
    }

    sqlx::query(
        "INSERT INTO `isu_association_config` (`name`, `url`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)",
    )
    .bind("jia_service_url")
    .bind(&request.jia_service_url)
    .execute(pool.as_ref())
    .await
    .map_err(|e| {
        log::error!("db error : {}", e);
        SqlxError(e)
    })?;
    Ok(HttpResponse::Ok().json(InitializeResponse {
        language: "rust".to_owned(),
    }))
}

#[derive(Debug, serde::Deserialize)]
struct Claims {
    jia_user_id: String,
}

#[actix_web::post("/api/auth")]
async fn post_authentication(
    pool: web::Data<sqlx::MySqlPool>,
    request: actix_web::HttpRequest,
    session: actix_session::Session,
) -> actix_web::Result<HttpResponse> {
    let req_jwt = request
        .headers()
        .get("Authorization")
        .map(|value| value.to_str().unwrap_or_default())
        .unwrap_or_default()
        .trim_start_matches("Bearer ");

    let validation = jsonwebtoken::Validation::new(jsonwebtoken::Algorithm::ES256);
    let token = match jsonwebtoken::decode(req_jwt, &JWT_VERIFICATION_KEY, &validation) {
        Ok(token) => token,
        Err(e) => {
            if matches!(e.kind(), jsonwebtoken::errors::ErrorKind::Json(_)) {
                return Ok(HttpResponse::BadRequest()
                    .content_type(mime::TEXT_PLAIN)
                    .body("invalid JWT payload"));
            } else {
                return Ok(HttpResponse::Forbidden()
                    .content_type(mime::TEXT_PLAIN)
                    .body("forbidden"));
            }
        }
    };

    let claims: Claims = token.claims;
    let jia_user_id = claims.jia_user_id;

    sqlx::query("INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)")
        .bind(&jia_user_id)
        .execute(pool.as_ref())
        .await
        .map_err(|e| {
            log::error!("db error: {}", e);
            SqlxError(e)
        })?;

    session.insert("jia_user_id", jia_user_id).map_err(|e| {
        log::error!("failed to set cookie: {}", e);
        e
    })?;

    Ok(HttpResponse::Ok().finish())
}

#[actix_web::post("/api/signout")]
async fn post_signout(session: actix_session::Session) -> actix_web::Result<HttpResponse> {
    if session.remove("jia_user_id").is_some() {
        Ok(HttpResponse::Ok().finish())
    } else {
        Ok(HttpResponse::Unauthorized()
            .content_type(mime::TEXT_PLAIN)
            .body("you are not signed in"))
    }
}

#[actix_web::get("/api/user/me")]
async fn get_me(
    pool: web::Data<sqlx::MySqlPool>,
    session: actix_session::Session,
) -> actix_web::Result<HttpResponse> {
    let jia_user_id = require_signed_in(pool.as_ref(), session).await?;
    Ok(HttpResponse::Ok().json(GetMeResponse { jia_user_id }))
}

// 自分のISUの一覧を取得
#[actix_web::get("/api/isu")]
async fn get_isu_list(
    pool: web::Data<sqlx::MySqlPool>,
    session: actix_session::Session,
) -> actix_web::Result<HttpResponse> {
    let jia_user_id = require_signed_in(pool.as_ref(), session).await?;

    let mut tx = pool.begin().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    let isu_list: Vec<Isu> =
        sqlx::query_as("SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC")
            .bind(&jia_user_id)
            .fetch_all(&mut tx)
            .await
            .map_err(|e| {
                log::error!("db error: {}", e);
                SqlxError(e)
            })?;

    let mut response_list = Vec::new();
    for isu in isu_list {
        let last_condition: Option<IsuCondition> = sqlx::query_as(
            "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` DESC LIMIT 1"
        ).bind(&isu.jia_isu_uuid).fetch_optional(&mut tx).await.map_err(|e| {
            log::error!("db error: {}", e);
            SqlxError(e)
        })?;

        let res = if let Some(last_condition) = last_condition {
            let condition_level = calculate_condition_level(&last_condition.condition);
            if condition_level.is_none() {
                log::error!("failed to get condition level: unexpected warn count");
                return Ok(HttpResponse::InternalServerError()
                    .content_type(mime::TEXT_PLAIN)
                    .body(""));
            }
            let condition_level = condition_level.unwrap();
            GetIsuListResponse {
                id: isu.id,
                jia_isu_uuid: isu.jia_isu_uuid,
                name: isu.name.clone(),
                character: isu.character,
                latest_isu_condition: Some(GetIsuConditionResponse {
                    jia_isu_uuid: last_condition.jia_isu_uuid,
                    isu_name: isu.name,
                    timestamp: last_condition.timestamp.timestamp(),
                    is_sitting: last_condition.is_sitting,
                    condition: last_condition.condition,
                    condition_level,
                    message: last_condition.message,
                }),
            }
        } else {
            GetIsuListResponse {
                id: isu.id,
                jia_isu_uuid: isu.jia_isu_uuid,
                name: isu.name,
                character: isu.character,
                latest_isu_condition: None,
            }
        };
        response_list.push(res);
    }

    tx.commit().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    Ok(HttpResponse::Ok().json(response_list))
}

//  POST /api/isu
// 自分のISUの登録
#[actix_web::post("/api/isu")]
async fn post_isu(
    pool: web::Data<sqlx::MySqlPool>,
    session: actix_session::Session,
    mut payload: actix_multipart::Multipart,
    isu_condition_address: web::Data<IsuConditionAddress>,
) -> actix_web::Result<HttpResponse> {
    let jia_user_id = require_signed_in(pool.as_ref(), session).await?;

    let mut jia_isu_uuid = None;
    let mut isu_name = None;
    let mut image = None;
    while let Some(field) = payload.next().await {
        let field = field?;
        let content_disposition = field.content_disposition().unwrap();
        let content = field
            .map_ok(|chunk| bytes::BytesMut::from(&chunk[..]))
            .try_concat()
            .await
            .map_err(|e| {
                log::error!("failed to field: {}", e);
                e
            })?
            .freeze();
        match content_disposition.get_name().unwrap() {
            "jia_isu_uuid" => {
                jia_isu_uuid = Some(String::from_utf8_lossy(&content).into_owned());
            }
            "isu_name" => {
                isu_name = Some(String::from_utf8_lossy(&content).into_owned());
            }
            "image" => {
                image = Some(content);
            }
            _ => {}
        }
    }
    let jia_isu_uuid: String = jia_isu_uuid.unwrap_or_default();
    let isu_name: String = isu_name.unwrap_or_default();
    let image = match image {
        Some(image) => image,
        None => {
            let content = tokio::fs::read(DEFAULT_ICON_FILE_PATH).await.map_err(|e| {
                log::error!("failed to read default icon: {}", e);
                e
            })?;
            bytes::Bytes::from(content)
        }
    };

    let mut tx = pool.begin().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    let result = sqlx::query(
        "INSERT INTO `isu` (`jia_isu_uuid`, `name`, `image`, `jia_user_id`) VALUES (?, ?, ?, ?)",
    )
    .bind(&jia_isu_uuid)
    .bind(&isu_name)
    .bind(image.as_ref())
    .bind(&jia_user_id)
    .execute(&mut tx)
    .await;
    if let Err(sqlx::Error::Database(ref db_error)) = result {
        if let Some(mysql_error) = db_error.try_downcast_ref::<sqlx::mysql::MySqlDatabaseError>() {
            if mysql_error.number() == MYSQL_ERR_NUM_DUPLICATE_ENTRY {
                return Ok(HttpResponse::Conflict()
                    .content_type(mime::TEXT_PLAIN)
                    .body("duplicated: isu"));
            }
        }
    }
    result.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    let target_url = format!(
        "{}/api/activate",
        get_jia_service_url(&mut tx).await.map_err(SqlxError)?
    );
    let body = JIAServiceRequest {
        target_ip: isu_condition_address.public_address.clone(),
        target_port: isu_condition_address.public_port,
        isu_uuid: jia_isu_uuid.clone(),
    };

    let resp = reqwest::Client::new()
        .post(&target_url)
        .json(&body)
        .send()
        .await
        .map_err(|e| {
            log::error!("failed to request to JIAService: {}", e);
            ReqwestError(e)
        })?;

    if resp.status() != reqwest::StatusCode::ACCEPTED {
        log::error!("JIAService returned error: status code {}", resp.status());
        return Ok(HttpResponse::build(resp.status())
            .content_type(mime::TEXT_PLAIN)
            .body("JIAService returned error"));
    }

    let isu_from_jia: IsuFromJIA = resp.json().await.map_err(|e| {
        log::error!("error occured while reading JIA response: {}", e);
        ReqwestError(e)
    })?;

    sqlx::query("UPDATE `isu` SET `character` = ? WHERE  `jia_isu_uuid` = ?")
        .bind(&isu_from_jia.character)
        .bind(&jia_isu_uuid)
        .execute(&mut tx)
        .await
        .map_err(|e| {
            log::error!("db error: {}", e);
            SqlxError(e)
        })?;

    let isu: Isu =
        sqlx::query_as("SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?")
            .bind(&jia_user_id)
            .bind(&jia_isu_uuid)
            .fetch_one(&mut tx)
            .await
            .map_err(|e| {
                log::error!("db error: {}", e);
                SqlxError(e)
            })?;

    tx.commit().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    Ok(HttpResponse::Created().json(isu))
}

#[derive(Debug, serde::Deserialize)]
struct GetIsuSearchQuery {
    name: Option<String>,
    character: Option<String>,
    catalog_name: Option<String>,
    min_limit_weight: Option<i64>,
    max_limit_weight: Option<i64>,
    catalog_tags: Option<String>,
    page: Option<i64>,
}

// 椅子の情報を取得する
#[actix_web::get("/api/isu/{jia_isu_uuid}")]
async fn get_isu(
    pool: web::Data<sqlx::MySqlPool>,
    session: actix_session::Session,
    jia_isu_uuid: web::Path<String>,
) -> actix_web::Result<HttpResponse> {
    let jia_user_id = require_signed_in(pool.as_ref(), session).await?;

    let isu: Option<Isu> =
        sqlx::query_as("SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?")
            .bind(&jia_user_id)
            .bind(jia_isu_uuid.as_ref())
            .fetch_optional(pool.as_ref())
            .await
            .map_err(|e| {
                log::error!("db error: {}", e);
                SqlxError(e)
            })?;
    if isu.is_none() {
        return Ok(HttpResponse::NotFound()
            .content_type(mime::TEXT_PLAIN)
            .body("not found: isu"));
    }
    let isu = isu.unwrap();

    Ok(HttpResponse::Ok().json(isu))
}

// ISUのアイコンを取得する
#[actix_web::get("/api/isu/{jia_isu_uuid}/icon")]
async fn get_isu_icon(
    pool: web::Data<sqlx::MySqlPool>,
    session: actix_session::Session,
    jia_isu_uuid: web::Path<String>,
) -> actix_web::Result<HttpResponse> {
    let jia_user_id = require_signed_in(pool.as_ref(), session).await?;

    let image: Option<Vec<u8>> = sqlx::query_scalar(
        "SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
    )
    .bind(&jia_user_id)
    .bind(jia_isu_uuid.as_ref())
    .fetch_optional(pool.as_ref())
    .await
    .map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    if let Some(image) = image {
        Ok(HttpResponse::Ok().body(image))
    } else {
        Ok(HttpResponse::NotFound()
            .content_type(mime::TEXT_PLAIN)
            .body("not found: isu"))
    }
}

#[derive(Debug, serde::Deserialize)]
struct GetIsuGraphQuery {
    datetime: Option<String>,
}
// グラフ描画のための情報を計算して返却する
// ユーザーがISUの機嫌を知りたい
// この時間帯とか、この日とかの機嫌を知りたい
// 日毎時間単位グラフ
// conditionを何件か集めて、ISUにとっての快適度数みたいな値を算出する
#[actix_web::get("/api/isu/{jia_isu_uuid}/graph")]
async fn get_isu_graph(
    pool: web::Data<sqlx::MySqlPool>,
    session: actix_session::Session,
    jia_isu_uuid: web::Path<String>,
    query: web::Query<GetIsuGraphQuery>,
) -> actix_web::Result<HttpResponse> {
    let jia_user_id = require_signed_in(pool.as_ref(), session).await?;

    let date = match &query.datetime {
        Some(datetime_str) => match datetime_str.parse() {
            Ok(datetime) => truncate_after_hours(DateTime::from_utc(
                NaiveDateTime::from_timestamp(datetime, 0),
                JST_OFFSET.fix(),
            )),
            Err(_) => {
                return Ok(HttpResponse::BadRequest()
                    .content_type(mime::TEXT_PLAIN)
                    .body("bad format: datetime"));
            }
        },
        None => {
            return Ok(HttpResponse::BadRequest()
                .content_type(mime::TEXT_PLAIN)
                .body("missing: datetime"));
        }
    };

    let mut tx = pool.begin().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    let count: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
    )
    .bind(&jia_user_id)
    .bind(jia_isu_uuid.as_ref())
    .fetch_one(&mut tx)
    .await
    .map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;
    if count == 0 {
        return Ok(HttpResponse::NotFound()
            .content_type(mime::TEXT_PLAIN)
            .body("not found: isu"));
    }

    let res = generate_isu_graph_response(&mut tx, &jia_isu_uuid, date)
        .await
        .map_err(|e| {
            log::error!("failed to generate isu graph: {}", e);
            SqlxError(e)
        })?;

    tx.commit().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    Ok(HttpResponse::Ok().json(res))
}

// GET /api/isu/{jia_isu_uuid}/graph のレスポンス作成のため，
// グラフのデータ点を一日分生成
async fn generate_isu_graph_response(
    tx: &mut sqlx::Transaction<'_, sqlx::MySql>,
    jia_isu_uuid: &str,
    graph_date: DateTime<chrono::FixedOffset>,
) -> sqlx::Result<Vec<GraphResponse>> {
    let mut data_points = Vec::new();
    let mut conditions_in_this_hour = Vec::new();
    let mut timestamps_in_this_hour = Vec::new();
    let mut start_time_in_this_hour =
        DateTime::from_utc(NaiveDateTime::from_timestamp(0, 0), JST_OFFSET.fix());

    let mut rows = sqlx::query_as(
        "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC",
    )
    .bind(jia_isu_uuid)
    .fetch(tx);

    while let Some(row) = rows.next().await {
        let condition: IsuCondition = row?;

        let truncated_condition_time = truncate_after_hours(condition.timestamp);
        if truncated_condition_time != start_time_in_this_hour {
            if !conditions_in_this_hour.is_empty() {
                let data = calculate_graph_data_point(&conditions_in_this_hour);
                data_points.push(GraphDataPointWithInfo {
                    jia_isu_uuid: jia_isu_uuid.to_owned(),
                    start_at: start_time_in_this_hour,
                    data,
                    condition_timestamps: timestamps_in_this_hour.clone(),
                });
            }

            start_time_in_this_hour = truncated_condition_time;
            conditions_in_this_hour = Vec::new();
            timestamps_in_this_hour = Vec::new();
        }
        timestamps_in_this_hour.push(condition.timestamp.timestamp());
        conditions_in_this_hour.push(condition);
    }

    if !conditions_in_this_hour.is_empty() {
        let data = calculate_graph_data_point(&conditions_in_this_hour);
        data_points.push(GraphDataPointWithInfo {
            jia_isu_uuid: jia_isu_uuid.to_owned(),
            start_at: start_time_in_this_hour,
            data,
            condition_timestamps: timestamps_in_this_hour,
        });
    }

    let end_time = graph_date + chrono::Duration::hours(24);
    let mut filtered_data_points = data_points
        .into_iter()
        .skip_while(|graph| graph.start_at < graph_date)
        .take_while(|graph| graph.start_at < end_time)
        .peekable();

    let mut response_list = Vec::new();
    let mut this_time = graph_date;

    while this_time < end_time {
        let (data, timestamps) = filtered_data_points
            .next_if(|data_with_info| data_with_info.start_at == this_time)
            .map(|data_with_info| {
                (
                    Some(data_with_info.data),
                    data_with_info.condition_timestamps,
                )
            })
            .unwrap_or_else(|| (None, Vec::new()));

        let resp = GraphResponse {
            start_at: this_time.timestamp(),
            end_at: (this_time + chrono::Duration::hours(1)).timestamp(),
            data,
            condition_timestamps: timestamps,
        };
        response_list.push(resp);

        this_time = this_time + chrono::Duration::hours(1);
    }

    Ok(response_list)
}

// 複数のISU conditionからグラフの一つのデータ点を計算
fn calculate_graph_data_point(isu_conditions: &[IsuCondition]) -> GraphDataPoint {
    use std::iter::FromIterator as _;

    let mut conditions_count: HashMap<&str, i64> =
        HashMap::from_iter([("is_broken", 0), ("is_dirty", 0), ("is_overweight", 0)]);
    let mut raw_score = 0;
    for condition in isu_conditions {
        let conditions = condition
            .condition
            .split(',')
            .map(|cond_str| {
                let mut key_value = cond_str.split('=');
                (key_value.next().unwrap(), key_value.next().unwrap())
            })
            .filter(|(_, value)| *value == "true");
        let mut bad_conditions_count = 0;
        for (condition_name, _) in conditions {
            bad_conditions_count += 1;
            *conditions_count.get_mut(&condition_name).unwrap() += 1;
        }

        if bad_conditions_count >= 3 {
            raw_score += SCORE_CONDITION_LEVEL_CRITICAL;
        } else if bad_conditions_count >= 1 {
            raw_score += SCORE_CONDITION_LEVEL_WARNING;
        } else {
            raw_score += SCORE_CONDITION_LEVEL_INFO;
        }
    }

    let sitting_count = isu_conditions
        .iter()
        .filter(|condition| condition.is_sitting)
        .count() as i64;

    let isu_conditions_length = isu_conditions.len() as i64;
    let score = raw_score / isu_conditions_length;

    let sitting_percentage = sitting_count * 100 / isu_conditions_length;
    let is_broken_percentage =
        conditions_count.get("is_broken").unwrap() * 100 / isu_conditions_length;
    let is_overweight_percentage =
        conditions_count.get("is_overweight").unwrap() * 100 / isu_conditions_length;
    let is_dirty_percentage =
        conditions_count.get("is_dirty").unwrap() * 100 / isu_conditions_length;

    GraphDataPoint {
        score,
        percentage: ConditionsPercentage {
            sitting: sitting_percentage,
            is_broken: is_broken_percentage,
            is_overweight: is_overweight_percentage,
            is_dirty: is_dirty_percentage,
        },
    }
}

#[derive(Debug, serde::Deserialize)]
struct GetIsuConditionsQuery {
    end_time: Option<String>,
    condition_level: Option<String>,
    start_time: Option<String>,
    limit: Option<String>,
}

// 自分の所持椅子のうち、指定した椅子の通知を取得する
#[actix_web::get("/api/condition/{jia_isu_uuid}")]
async fn get_isu_conditions(
    pool: web::Data<sqlx::MySqlPool>,
    session: actix_session::Session,
    jia_isu_uuid: web::Path<String>,
    query: web::Query<GetIsuConditionsQuery>,
) -> actix_web::Result<HttpResponse> {
    let jia_user_id = require_signed_in(pool.as_ref(), session).await?;

    let end_time = match &query.end_time {
        Some(end_time_str) => match end_time_str.parse() {
            Ok(end_time) => {
                DateTime::from_utc(NaiveDateTime::from_timestamp(end_time, 0), JST_OFFSET.fix())
            }
            Err(_) => {
                return Ok(HttpResponse::BadRequest()
                    .content_type(mime::TEXT_PLAIN)
                    .body("bad format: end_time"));
            }
        },
        None => {
            return Ok(HttpResponse::BadRequest()
                .content_type(mime::TEXT_PLAIN)
                .body("bad format: end_time"));
        }
    };
    if query.condition_level.is_none() {
        return Ok(HttpResponse::BadRequest()
            .content_type(mime::TEXT_PLAIN)
            .body("missing: condition_level"));
    }
    let mut condition_level = HashSet::new();
    for level in query.condition_level.as_ref().unwrap().split(',') {
        condition_level.insert(level);
    }

    let start_time = match &query.start_time {
        Some(start_time_str) => match start_time_str.parse() {
            Ok(start_time) => Some(DateTime::from_utc(
                NaiveDateTime::from_timestamp(start_time, 0),
                JST_OFFSET.fix(),
            )),
            Err(_) => {
                return Ok(HttpResponse::BadRequest()
                    .content_type(mime::TEXT_PLAIN)
                    .body("bad format: start_time"));
            }
        },
        None => None,
    };
    let limit = match &query.limit {
        Some(limit_str) => match limit_str.parse() {
            Ok(limit) => {
                if limit <= 0 {
                    return Ok(HttpResponse::BadRequest()
                        .content_type(mime::TEXT_PLAIN)
                        .body("bad format: limit"));
                } else {
                    limit
                }
            }
            Err(_) => {
                return Ok(HttpResponse::BadRequest()
                    .content_type(mime::TEXT_PLAIN)
                    .body("bad format: limit"));
            }
        },
        None => CONDITION_LIMIT,
    };

    let isu_name: Option<String> =
        sqlx::query_scalar("SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ?")
            .bind(jia_isu_uuid.as_ref())
            .bind(&jia_user_id)
            .fetch_optional(pool.as_ref())
            .await
            .map_err(|e| {
                log::error!("db error: {}", e);
                SqlxError(e)
            })?;
    if isu_name.is_none() {
        log::error!("isu not found");
        return Ok(HttpResponse::NotFound()
            .content_type(mime::TEXT_PLAIN)
            .body("not found: isu"));
    }
    let isu_name = isu_name.unwrap();

    let conditions_response = get_isu_conditions_from_db(
        &pool,
        &jia_isu_uuid,
        end_time,
        &condition_level,
        &start_time,
        limit as usize,
        &isu_name,
    )
    .await
    .map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    Ok(HttpResponse::Ok().json(conditions_response))
}

async fn get_isu_conditions_from_db(
    pool: &sqlx::MySqlPool,
    jia_isu_uuid: &str,
    end_time: DateTime<chrono::FixedOffset>,
    condition_level: &HashSet<&str>,
    start_time: &Option<DateTime<chrono::FixedOffset>>,
    limit: usize,
    isu_name: &str,
) -> sqlx::Result<Vec<GetIsuConditionResponse>> {
    let conditions: Vec<IsuCondition> = if let Some(ref start_time) = start_time {
        sqlx::query_as(
            "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? AND `timestamp` < ?	AND ? <= `timestamp` ORDER BY `timestamp` DESC",
        )
            .bind(jia_isu_uuid)
            .bind(end_time.naive_local())
            .bind(start_time.naive_local())
            .fetch_all(pool)
    } else {
        sqlx::query_as(
            "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? AND `timestamp` < ? ORDER BY `timestamp` DESC",
        )
        .bind(jia_isu_uuid)
        .bind(end_time.naive_local())
        .fetch_all(pool)
    }.await?;

    let mut conditions_response = Vec::new();
    for c in conditions {
        if let Some(c_level) = calculate_condition_level(&c.condition) {
            if condition_level.contains(c_level) {
                conditions_response.push(GetIsuConditionResponse {
                    jia_isu_uuid: c.jia_isu_uuid,
                    isu_name: isu_name.to_owned(),
                    timestamp: c.timestamp.timestamp(),
                    is_sitting: c.is_sitting,
                    condition: c.condition,
                    condition_level: c_level,
                    message: c.message,
                });
            }
        }
    }

    if conditions_response.len() > limit {
        conditions_response.truncate(limit);
    }

    Ok(conditions_response)
}

// conditionのcsvからcondition levelを計算
fn calculate_condition_level(condition: &str) -> Option<&'static str> {
    let warn_count = condition.matches("=true").count();
    match warn_count {
        0 => Some(CONDITION_LEVEL_INFO),
        1 | 2 => Some(CONDITION_LEVEL_WARNING),
        3 => Some(CONDITION_LEVEL_CRITICAL),
        _ => None,
    }
}

// ISUからのセンサデータを受け取る
#[actix_web::post("/api/condition/{jia_isu_uuid}")]
async fn post_isu_condition(
    pool: web::Data<sqlx::MySqlPool>,
    jia_isu_uuid: web::Path<String>,
    req: web::Json<Vec<PostIsuConditionRequest>>,
) -> actix_web::Result<HttpResponse> {
    // TODO: これ良くないので後でなんとかする
    const DROP_PROBABILITY: f64 = 0.1;
    if rand::random::<f64>() <= DROP_PROBABILITY {
        log::warn!("drop post isu condition request");
        return Ok(HttpResponse::ServiceUnavailable()
            .content_type(mime::TEXT_PLAIN)
            .body(""));
    }

    if req.is_empty() {
        log::error!("bad request body: array length is 0");
        return Ok(HttpResponse::BadRequest()
            .content_type(mime::TEXT_PLAIN)
            .body("bad request body"));
    }

    let mut tx = pool.begin().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    let count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?")
        .bind(jia_isu_uuid.as_ref())
        .fetch_one(&mut tx)
        .await
        .map_err(|e| {
            log::error!("db error: {}", e);
            SqlxError(e)
        })?;
    if count == 0 {
        log::error!("isu not found");
        return Ok(HttpResponse::NotFound()
            .content_type(mime::TEXT_PLAIN)
            .body("not found: isu"));
    }

    for cond in req.iter() {
        let timestamp: DateTime<chrono::FixedOffset> = DateTime::from_utc(
            NaiveDateTime::from_timestamp(cond.timestamp, 0),
            JST_OFFSET.fix(),
        );

        if !is_valid_condition_format(&cond.condition) {
            return Ok(HttpResponse::BadRequest()
                .content_type(mime::TEXT_PLAIN)
                .body("bad request body"));
        }

        sqlx::query(
            "INSERT INTO `isu_condition` (`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`) VALUES (?, ?, ?, ?, ?)",
        )
            .bind(jia_isu_uuid.as_ref())
            .bind(&timestamp.naive_local())
            .bind(&cond.is_sitting)
            .bind(&cond.condition)
            .bind(&cond.message)
            .execute(&mut tx)
            .await.map_err(|e| {
            log::error!("db error: {}", e);
            SqlxError(e)
        })?;
    }

    tx.commit().await.map_err(|e| {
        log::error!("db error: {}", e);
        SqlxError(e)
    })?;

    Ok(HttpResponse::Created().finish())
}

// conditionの文字列がcsv形式になっているか検証
fn is_valid_condition_format(condition_str: &str) -> bool {
    let keys = ["is_dirty=", "is_overweight=", "is_broken="];
    const VALUE_TRUE: &str = "true";
    const VALUE_FALSE: &str = "false";

    let mut idx_cond_str = 0;

    for (idx_keys, key) in keys.iter().enumerate() {
        if !condition_str[idx_cond_str..].starts_with(key) {
            return false;
        }
        idx_cond_str += key.len();

        if condition_str[idx_cond_str..].starts_with(VALUE_TRUE) {
            idx_cond_str += VALUE_TRUE.len();
        } else if condition_str[idx_cond_str..].starts_with(VALUE_FALSE) {
            idx_cond_str += VALUE_FALSE.len();
        } else {
            return false;
        }

        if idx_keys < keys.len() - 1 {
            if !condition_str[idx_cond_str..].starts_with(',') {
                return false;
            }
            idx_cond_str += 1;
        }
    }

    idx_cond_str == condition_str.len()
}

//分以下を切り捨て、一時間単位にする関数
fn truncate_after_hours<T>(datetime: T) -> T
where
    T: chrono::Timelike,
{
    datetime.with_minute(0).unwrap().with_second(0).unwrap()
}

// ISUの性格毎の最新のコンディション情報
#[actix_web::get("/api/trend")]
async fn get_trend(pool: web::Data<sqlx::MySqlPool>) -> actix_web::Result<HttpResponse> {
    let character_list: Vec<String> =
        sqlx::query_scalar("SELECT `character` FROM `isu` GROUP BY `character`")
            .fetch_all(pool.as_ref())
            .await
            .map_err(|e| {
                log::error!("db error: {}", e);
                SqlxError(e)
            })?;

    let mut res = Vec::new();

    // TODO: 処理が重すぎるのでなんとかする
    for character in character_list {
        let isu_list: Vec<Isu> = sqlx::query_as("SELECT * FROM `isu` WHERE `character` = ?")
            .bind(&character)
            .fetch_all(pool.as_ref())
            .await
            .map_err(|e| {
                log::error!("db error: {}", e);
                SqlxError(e)
            })?;

        let mut character_isu_conditions = Vec::new();
        for isu in isu_list {
            let conditions: Vec<IsuCondition> = sqlx::query_as(
                "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY timestamp DESC",
            )
            .bind(&isu.jia_isu_uuid)
            .fetch_all(pool.as_ref())
            .await
            .map_err(|e| {
                log::error!("db error: {}", e);
                SqlxError(e)
            })?;

            if !conditions.is_empty() {
                let isu_last_condition = &conditions[0];
                let condition_level = calculate_condition_level(&isu_last_condition.condition);
                if condition_level.is_none() {
                    log::error!("failed to get condition level: unexpected warn count");
                    return Ok(HttpResponse::InternalServerError()
                        .content_type(mime::TEXT_PLAIN)
                        .body(""));
                }
                let condition_level = condition_level.unwrap();
                let trend_condition = TrendCondition {
                    id: isu.id,
                    timestamp: isu_last_condition.timestamp.timestamp(),
                    condition_level,
                };
                character_isu_conditions.push(trend_condition);
            }
        }

        character_isu_conditions.sort_by_key(|condition| std::cmp::Reverse(condition.timestamp));
        res.push(TrendResponse {
            character,
            conditions: character_isu_conditions,
        });
    }

    Ok(HttpResponse::Ok().json(res))
}
