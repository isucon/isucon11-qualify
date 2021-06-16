package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	sessionName                 = "isucondition"
	searchLimit                 = 20
	notificationLimit           = 20
	notificationTimestampFormat = "2006-01-02 15:04:05 -0700"
)

var scorePerCondition = map[string]int{
	"is_dirty":      -1,
	"is_overweight": -1,
	"is_broken":     -5,
}

var (
	db                  *sqlx.DB
	sessionStore        sessions.Store
	mySQLConnectionData *MySQLConnectionEnv
)

type Isu struct {
	JIAIsuUUID   string    `db:"jia_isu_uuid" json:"jia_isu_uuid"`
	Name         string    `db:"name" json:"name"`
	Image        []byte    `db:"image" json:"-"`
	JIACatalogID string    `db:"jia_catalog_id" json:"jia_catalog_id"`
	Character    string    `db:"character" json:"character"`
	JIAUserID    string    `db:"jia_user_id" json:"-"`
	IsDeleted    bool      `db:"is_deleted" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"-"`
	UpdatedAt    time.Time `db:"updated_at" json:"-"`
}

type CatalogFromJIA struct {
	JIACatalogID string `json:"catalog_id"`
	Name         string `json:"name"`
	LimitWeight  int    `json:"limit_weight"`
	Weight       int    `json:"weight"`
	Size         string `json:"size"`
	Maker        string `json:"maker"`
	Features     string `json:"features"`
}

type Catalog struct {
	JIACatalogID string `json:"jia_catalog_id"`
	Name         string `json:"name"`
	LimitWeight  int    `json:"limit_weight"`
	Weight       int    `json:"weight"`
	Size         string `json:"size"`
	Maker        string `json:"maker"`
	Tags         string `json:"tags"`
}

type IsuLog struct {
	JIAIsuUUID string    `db:"jia_isu_uuid" json:"jia_isu_uuid"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"`
	IsSitting  bool      `db:"is_sitting" json:"is_sitting"`
	Condition  string    `db:"condition" json:"condition"`
	Message    string    `db:"message" json:"message"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

//グラフ表示用  一時間のsummry 詳細
type GraphData struct {
	Score   int            `json:"score"`
	Sitting int            `json:"sitting"`
	Detail  map[string]int `json:"detail"`
}

//グラフ表示用  一時間のsummry
type Graph struct {
	JIAIsuUUID string    `db:"jia_isu_uuid"`
	StartAt    time.Time `db:"start_at"`
	Data       string    `db:"data"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type User struct {
	JIAUserID string    `db:"jia_user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type MySQLConnectionEnv struct {
	Host     string
	Port     string
	User     string
	DBName   string
	Password string
}

type InitializeResponse struct {
	Language string `json:"language"`
}

type GetMeResponse struct {
}

type GraphResponse struct {
}

type IsuCondition struct {
	IsDirty      bool `json:"is_dirty"`
	IsOverweight bool `json:"is_overweight"`
	IsBroken     bool `json:"is_broken"`
}
type NotificationRequest struct {
	IsSitting bool         `json:"is_sitting"`
	Condition IsuCondition `json:"condition"`
	Message   string       `json:"message"`
	Timestamp string       `json:"timestamp"` //Format("2006-01-02 15:04:05 -0700")
}
type NotificationResponse struct {
}

func getEnv(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}

func NewMySQLConnectionEnv() *MySQLConnectionEnv {
	return &MySQLConnectionEnv{
		Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
		Port:     getEnv("MYSQL_PORT", "3306"),
		User:     getEnv("MYSQL_USER", "isucon"),
		DBName:   getEnv("MYSQL_DBNAME", "isucondition"),
		Password: getEnv("MYSQL_PASS", "isucon"),
	}
}

//ConnectDB データベースに接続する
func (mc *MySQLConnectionEnv) ConnectDB() (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true&loc=Local", mc.User, mc.Password, mc.Host, mc.Port, mc.DBName)
	return sqlx.Open("mysql", dsn)
}

func init() {
	sessionStore = sessions.NewCookieStore([]byte(getEnv("SESSION_KEY", "isucondition")))
}

func main() {
	// Echo instance
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize
	e.POST("/initialize", postInitialize)

	e.POST("/api/auth", postAuthentication)
	e.POST("/api/signout", postSignout)
	e.GET("/api/user/me", getMe)

	e.GET("/api/catalog/:jia_catalog_id", getCatalog)
	e.GET("/api/isu", getIsuList)
	e.POST("/api/isu", postIsu)
	e.GET("/api/isu/search", getIsuSearch)
	e.GET("/api/isu/:jia_isu_uuid", getIsu)
	e.PUT("/api/isu/:jia_isu_uuid", putIsu)
	e.DELETE("/api/isu/:jia_isu_uuid", deleteIsu)
	e.GET("/api/isu/:jia_isu_uuid/icon", getIsuIcon)
	e.PUT("/api/isu/:jia_isu_uuid/icon", putIsuIcon)
	e.GET("/api/isu/:jia_isu_uuid/graph", getIsuGraph)
	e.GET("/api/notification", getNotifications)
	e.GET("/api/notification/:jia_isu_uuid", getIsuNotifications)

	e.POST("/api/isu/:jia_isu_uuid/condition", postIsuCondition)

	mySQLConnectionData = NewMySQLConnectionEnv()

	var err error
	db, err = mySQLConnectionData.ConnectDB()
	if err != nil {
		e.Logger.Fatalf("DB connection failed : %v", err)
		return
	}
	db.SetMaxOpenConns(10)
	defer db.Close()

	// Start server
	serverPort := fmt.Sprintf(":%v", getEnv("SERVER_PORT", "3000"))
	e.Logger.Fatal(e.Start(serverPort))
}

func getSession(r *http.Request) *sessions.Session {
	session, _ := sessionStore.Get(r, sessionName)
	return session
}

func getUserIdFromSession(r *http.Request) (string, error) {
	session := getSession(r)
	userID, ok := session.Values["jia_user_id"]
	if !ok {
		return "", fmt.Errorf("no session")
	}
	return userID.(string), nil
}

func postInitialize(c echo.Context) error {
	sqlDir := filepath.Join("..", "mysql", "db")
	paths := []string{
		filepath.Join(sqlDir, "0_Schema.sql"),
	}

	for _, p := range paths {
		sqlFile, _ := filepath.Abs(p)
		cmdStr := fmt.Sprintf("mysql -h %v -u %v -p%v -P %v %v < %v",
			mySQLConnectionData.Host,
			mySQLConnectionData.User,
			mySQLConnectionData.Password,
			mySQLConnectionData.Port,
			mySQLConnectionData.DBName,
			sqlFile,
		)
		err := exec.Command("bash", "-c", cmdStr).Run()
		if err != nil {
			c.Logger().Errorf("Initialize script error : %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	return c.JSON(http.StatusOK, InitializeResponse{
		Language: "go",
	})
}

//  POST /api/auth
func postAuthentication(c echo.Context) error {
	// ユーザからの入力
	// * jwt

	// jwt の verify
	// NG だったら resp 400(Bad Request) or 403(期限切れとか)

	// jwt から username を取得
	//なかったら 400

	// トランザクション開始

	// DB にて存在確認
	// SELECT COUNT(*) FROM users WHERE username = `username`;

	// もし見つからなかったら
	// INSERT INTO users(username) VALUES(`username`); //uniqueなのでtx無しでこれだけでもOK(ダミーの改善点)
	// すでに存在するユーザー名なら409 ← 上で存在確認してるけど、このレベルのハンドリングつける？
	// 失敗したら500

	// トランザクション終了

	// Cookieを付与
	// 見つかったら 200
	return fmt.Errorf("not implemented")
}

//  POST /api/signout
func postSignout(c echo.Context) error {
	// ユーザからの入力
	// * session
	// session が存在しなければ 401

	// cookie の max-age を -1 にして Set-Cookie

	// response 200
	return fmt.Errorf("not implemented")
}

// TODO
// GET /api/user/{jia_user_id}
// ユーザ情報を取得
// day2 実装のため skip
// func getUser(c echo.Context) error {
// }

//  GET /api/user/me
// 自分のユーザー情報を取得
func getMe(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// SELECT jia_user_id, user_name FROM users WHERE jia_user_id = {jia_user_id};

	//response 200
	// * jia_user_id
	// * user_name
	return fmt.Errorf("not implemented")
}

//  GET /api/catalog/{jia_catalog_id}
// MEMO: 外部APIのドキュメントに「競技期間中は不変」「read onlyである」「叩く頻度は変えて良い」を明記
// MEMO: ISU協会のrespと当Appのrespのフォーマットは合わせる?  → 合わせない
// MEMO: day2: ISU 協会の問い合わせ内容をまるっとキャッシュするだけだと減点する仕組みを実装する
//         (ISU 協会の response に updated_at を加え、ベンチでここをチェックするようにする)
//		 ref: https://scrapbox.io/ISUCON11/2021.06.02_%E4%BA%88%E9%81%B8%E5%AE%9A%E4%BE%8B#60b764b725829c0000426896
func getCatalog(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// ISU 協会に問い合わせる  (想定解はキャッシュ)
	// request
	// * jia_catalog_id
	// response
	// * isu_catalog_name
	// * isu_catalog_limit_weight
	// * isu_catalog_size
	// * isu_catalog_weight
	// * isu_catalog_maker
	// * isu_catalog_tag: (headrest, personality, ...などジェイウォークになってるもの。種類は固定でドキュメントに記載しておく)
	// GET /api/isu/search で使用

	// ISU 協会からのレスポンス
	// 404 なら404を返す

	// 一つずつ変数を代入
	// validation（Nginxで横流しをOKにするかで決める）
	// ゼロ値が入ってないかの軽いチェック
	// 違反したら 500(初期実装では書くけどチェックはしない？)
	// 要検討

	// response 200
	return fmt.Errorf("not implemented")
}

//  GET /api/isu?limit=5
// 自分の ISU 一覧を取得
func getIsuList(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input
	//     * limit: 取得件数（利用用途的には固定だが一般的な話で指定可能にする。ベンチもやる）

	// SELECT * FROM isu WHERE jia_user_id = {jia_user_id} and is_deleted=false LIMIT {limit} order by created_at;
	// (catalogは取らない)
	// 画像までSQLで取ってくるボトルネック
	// imageも最初はとってるけどレスポンスに含まれてないからselect時に持ってくる必要ない

	// response 200
	// * id
	// * name
	// * jia_catalog_id
	// * charactor  // MEMO: この値を使うのは day2 実装だが、ひとまずフィールドは用意する
	return fmt.Errorf("not implemented")
}

//  POST /api/isu
// 自分のISUの登録
func postIsu(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input
	// 		jia_isu_uuid: 椅子固有のID（衝突しないようにUUID的なもの設定）
	// 		isu_name: 椅子の名前

	// req := contextからいい感じにinputとuser_idを取得
	// 形式が違うかったら400
	// (catalog,charactor), err := 外部API
	// ISU 協会にactivate
	// request
	// 	* jia_isu_uuid
	// response
	// 	* jia_catalog_id
	// 	* charactor
	// レスポンスが200以外なら横流し
	// 404, 403(認証拒否), 400, 5xx
	// 403はday2

	// imageはデフォルトを挿入
	// INSERT INTO isu VALUES (jia_isu_uuid, isu_name, image, catalog_, charactor, jia_user_id);
	// jia_isu_uuid 重複時 409

	// SELECT (*) FROM isu WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;
	// 画像までSQLで取ってくるボトルネック
	// imageも最初はとってるけどレスポンスに含まれてないからselect時に持ってくる必要ない

	// response 200
	//{
	// * id
	// * name
	// * jia_catalog_id
	// * charactor
	//]
	return fmt.Errorf("not implemented")
}

//  GET /api/isu/search
func getIsuSearch(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input (query_param) (required field はなし, 全て未指定の場合 /api/isu と同じクエリが発行される)
	//  * name
	//  * charactor
	//	* catalog_name
	//	* min_limit_weight
	//	* max_limit_weight
	//	* catalog_tags (ジェイウォーク)
	//  * page: （default = 1)
	//	* MEMO: 二つのカラムを併せて計算した値での検索（x*yの面積での検索とか）

	// 想定解
	// whereを使う、インデックスを頑張って張る
	// SQLにcatalogを入れてJOINする
	// ISUテーブルにカラム追加してcatalog情報入れちゃうパターンならJOIN不要かも？

	// 持っている椅子を数件取得 (非効率な検索ロジック)
	// SELECT (*) FROM isu WHERE jia_user_id = `jia_user_id` AND is_deleted=false
	//   AND name LIKE CONCAT('%', ?, '%')
	//   AND charactor = `charactor`
	//   ORDER BY created_at;

	//for _, isu range isuList {
	// catalog := isu.catalogを使ってISU協会に問い合わせ(http request)

	// isu_calatog情報を使ってフィルター
	//if !(min_limit_weight < catalog.limit_weight && catalog.limit_weight < max_limit_weight) {
	//	continue
	//}

	// ... catalog_name,catalog_tags

	// res の 配列 append
	//}
	// 指定pageからlimit件数（固定）だけ取り出す(ボトルネックにしたいのでbreakは行わない）

	//response 200
	//[{
	// * id
	// * name
	// * jia_catalog_id
	// * charactor
	//}]
	return fmt.Errorf("not implemented")
}

//  GET /api/isu/{jia_isu_uuid}
// 椅子の情報を取得する
func getIsu(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input
	//		* jia_isu_uuid: 椅子固有のID

	// SELECT (*) FROM isu WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;
	// 見つからなければ404
	// user_idがリクエストユーザーのものでなければ404
	// 画像までSQLで取ってくるボトルネック
	// imageも最初はとってるけどレスポンスに含まれてないからselect時に持ってくる必要ない
	// MEMO: jia_user_id 判別はクエリに入れずその後のロジックとする？ (一通り完成した後に要考慮)

	// response  200
	//{
	// * id
	// * name
	// * jia_catalog_id
	// * charactor
	//}

	return fmt.Errorf("not implemented")
}

//  PUT /api/isu/{jia_isu_uuid}
// 自分の所有しているISUの情報を変更する
func putIsu(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input (path_param)
	// 	* jia_isu_uuid: 椅子固有のID
	// input (body)
	// 	* isu_name: 椅子の名称
	// invalid ならば 400

	// トランザクション開始

	// MEMO: dummy のボトルネック
	// current_userが認可された操作か確認
	//     * isu_idがcurrent_userのものか
	// SELECT COUNT(*) FROM isu WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;
	// NGならエラーを返す
	//   404 not found

	// DBを更新
	// UPDATE isu SET isu_name=? WHERE jia_isu_uuid = `jia_isu_uuid`;

	//更新後の値取得
	// SELECT (*) FROM isu WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;
	// 画像までSQLで取ってくるボトルネック
	// imageも最初はとってるけどレスポンスに含まれてないからselect時に持ってくる必要ない

	//トランザクション終了

	// response  200
	//{
	// * id
	// * name
	// * jia_catalog_id
	// * charactor
	//}
	return fmt.Errorf("not implemented")
}

//  DELETE /api/isu/{jia_isu_uuid}
// 所有しているISUを削除する
func deleteIsu(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input
	// 		* jia_isu_uuid: 椅子の固有ID

	// トランザクション開始

	// DBから当該のISUが存在するか検索
	// SELECT (*) FROM isu WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;
	// 存在しない場合 404 を返す

	// 存在する場合 ISU　の削除フラグを有効にして 204 を返す
	// UPDATE isu SET is_deleted = true WHERE jia_isu_uuid = `jia_isu_uuid`;

	// ISU協会にdectivateを送る
	// MEMO: ISU協会へのリクエストが失敗した時に DB をロールバックできるから

	// トランザクション終了
	// MEMO: もしコミット時にエラーが発生しうるならば、「ISU協会側はdeactivate済みだがDBはactive」という不整合が発生しうる

	//response 204
	return fmt.Errorf("not implemented")
}

//  GET /api/isu/{jia_isu_uuid}/icon
// ISUのアイコンを取得する
// MEMO: ヘッダーとかでキャッシュ効くようにするのが想定解？(ただしPUTはあることに注意)
//       nginxで認証だけ外部に投げるみたいなのもできるっぽい？（ちゃんと読んでいない）
//       https://tech.jxpress.net/entry/2018/08/23/104123
// MEMO: DB 内の image は longblob
func getIsuIcon(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// SELECT image FROM isu WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;
	// 見つからなければ404
	// user_idがリクエストユーザーのものでなければ404

	// response 200
	// image
	// MEMO: とりあえず未指定... Content-Type: image/png image/jpg
	return fmt.Errorf("not implemented")
}

//  PUT /api/isu/{jia_isu_uuid}/icon
// ISUのアイコンを登録する
// multipart/form-data
// MEMO: DB 内の image は longblob
func putIsuIcon(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	//トランザクション開始

	// SELECT image FROM isu WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;
	// 見つからなければ404
	// user_idがリクエストユーザーのものでなければ404

	// UPDATE isu SET image=? WHERE jia_user_id = `jia_user_id` and jia_isu_uuid = `jia_isu_uuid` and is_deleted=false;

	//トランザクション終了

	// response 200
	// {}
	return fmt.Errorf("not implemented")
}

//  GET /api/isu/{jia_isu_uuid}/graph
// グラフ描画のための情報を計算して返却する
// ユーザーがISUの機嫌を知りたい
// この時間帯とか、この日とかの機嫌を知りたい
// 日毎時間単位グラフ
// conditionを何件か集めて、ISUにとっての快適度数みたいな値を算出する
func getIsuGraph(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input (path_param)
	//	* jia_isu_uuid: 椅子の固有ID
	// input (query_param)
	//	* date (required)
	//		YYYY-MM-DD
	//

	// 自分のISUかチェック
	// SELECT count(*) from isu where jia_user_id=`jia_user_id` and id = `jia_isu_uuid` and is_deleted=false;
	// エラー: response 404

	// MEMO: シナリオ的にPostIsuConditionでgraphを生成する方がボトルネックになる想定なので初期実装はgraphテーブル作る
	// DBを検索。グラフ描画に必要な情報を取得
	// ボトルネック用に事前計算したものを取得
	// graphは POST /api/isu/{jia_isu_uuid}/condition で生成
	// SELECT * from graph
	//   WHERE jia_isu_uuid = `jia_isu_uuid` AND date<=start_at AND start_at < (date+1day)

	//SQLのレスポンスを成形
	//nullチェック

	// response 200:
	//    グラフ描画のための情報のjson
	//        * フロントで表示しやすい形式
	// [{
	// start_at: ...
	// end_at: ...
	// data: {
	//   score: 70,//>=0
	//   sitting: 50 (%),
	//   detail: {
	//     dirty: -10,
	//     over_weight: -20
	//   }
	// }},
	// {
	// start_at: …,
	// end_at: …,
	// data: null,
	// },
	// {...}, ...]
	return fmt.Errorf("not implemented")
}

//  GET /api/notification?
// 自分の所持椅子の通知を取得する
// MEMO: 1970/1/1みたいな時を超えた古代からのリクエストは表示するか → する
// 順序は最新順固定
func getNotifications(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input {jia_isu_uuid}が無い以外は、/api/notification/{jia_isu_uuid}と同じ
	//

	// cookieからユーザID取得
	// ユーザの所持椅子取得
	// SELECT * FROM isu where jia_user_id = ?;

	// ユーザの所持椅子毎に /api/notificaiton/{jia_isu_uuid} を叩く（こことマージ含めてボトルネック）
	// query_param は GET /api/notification (ここ) のリクエストと同じものを使い回す

	// ユーザの所持椅子ごとのデータをマージ（ここと個別取得部分含めてボトルネック）
	// 通知時間帯でソートして、limit件数（固定）該当するデータを返す
	// MEMO: 改善後はこんな感じのSQLで一発でとる
	// select * from isu_log where (isu_log.created_at, jia_isu_uuid) < (cursor.end_time, cursor.jia_isu_uuid)
	//  order by created_at desc,jia_isu_uuid desc limit ?
	// 10.1.36-MariaDB で確認

	//memo（没）
	// (select * from isu_log where (isu_log.created_at=cursor.end_time and jia_isu_uuid < cursor.jia_isu_uuid)
	//   or isu_log.created_at<cursor.end_time order by created_at desc,jia_isu_uuid desc limit ?)

	// response: 200
	// /api/notification/{jia_isu_uuid}と同じ
	return fmt.Errorf("not implemented")
}

//  GET /api/notification/{jia_isu_uuid}?start_time=
// 自分の所持椅子のうち、指定したisu_idの通知を取得する
func getIsuNotifications(c echo.Context) error {
	// * session
	// session が存在しなければ 401

	// input
	//     * jia_isu_uuid: 椅子の固有番号(path_param)
	//     * start_time: 開始時間
	//     * cursor_end_time: 終了時間 (required)
	//     * cursor_isu_id (required)
	//     * condition_level: critical,warning,info (csv)
	//               critical: conditions (is_dirty,is_overweight,is_broken) のうちtrueが3個
	//               warning: conditionsのうちtrueのものが1 or 2個
	//               info: warning無し
	//     * MEMO day2実装: message (文字列検索)

	// memo
	// 例: Google Cloud Logging の URL https://console.cloud.google.com/logs/query;query=resource.type%3D%22gce_instance%22%20resource.labels.instance_id%3D%22<DB_NAME>%22?authuser=1&project=<PROJECT_ID>&query=%0A

	// cookieからユーザID取得

	// isu_id存在確認、ユーザの所持椅子か確認
	// 対象isu_idのisu_name取得
	// 存在しなければ404

	// 対象isu_idの通知を取得(limit, cursorで絞り込み）
	// select * from isu_log where jia_isu_uuid = {jia_isu_uuid} AND (isu_log.created_a, jia_isu_uuid) < (cursor.end_time, cursor.jia_isu_uuid) order by created_at desc, jia_isu_uuid desc limit ?
	// MEMO: ↑で実装する

	//for {
	// conditions を元に condition_level (critical,warning,info) を算出
	//}

	// response: 200
	// [{
	//     * jia_isu_uuid
	//     * isu_name
	//     * timestamp
	//     * conditions: {"is_dirty": boolean, "is_overweight": boolean,"is_broken": boolean}
	//     * condition_level
	//     * message
	// },...]
	return fmt.Errorf("not implemented")
}

// POST /api/isu/{jia_isu_uuid}/condition
// ISUからのセンサデータを受け取る
func postIsuCondition(c echo.Context) error {
	// input (path_param)
	//	* jia_isu_uuid
	// input (body)
	//  * is_sitting:  true/false,
	// 	* condition: {
	//      is_dirty:    true/false,
	//      is_overweight: true/false,
	//      is_broken:   true/false,
	//    }
	//  * message
	//	* timestamp（秒まで）

	jiaIsuUUID := c.Param("isu_uuid")
	var request NotificationRequest
	err := c.Bind(&request)
	if jiaIsuUUID == "" || err != nil {
		// invalid ならば 400
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	//Parse
	timestamp, err := time.Parse(notificationTimestampFormat, request.Timestamp)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid timestamp")
	}
	conditionStr, err := json.Marshal(request.Condition)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "condition has bad format")
	}

	// トランザクション開始
	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin tx")
	}
	defer tx.Rollback()

	// jia_isu_uuid が存在するかを確認
	count := 0
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?  and `is_deleted`=false", jiaIsuUUID)
	if count == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "isu not found")
	}
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select")
	}

	//isu_logに記録
	//confilct確認
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu_log` WHERE (`timestamp`, `jia_isu_uuid`) = (?, ?)  FOR UPDATE",
		timestamp, jiaIsuUUID,
	)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select")
	}
	if count != 0 {
		return echo.NewHTTPError(http.StatusConflict, "isu_log already exist")
	}
	//insert
	_, err = tx.Exec("INSERT INTO `isu_log`"+
		"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`) VALUES (?, ?, ?, ?, ?)",
		jiaIsuUUID, timestamp, request.IsSitting, conditionStr, request.Message,
	)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert")
	}

	// getGraph用のデータを計算
	// IsuLogを一時間ごとの区切りに分け、区切りごとにスコアを計算する
	var isuLogCluster []IsuLog // 一時間ごとの纏まり
	var tmpIsuLog IsuLog
	rows, err := tx.Queryx("SELECT * FROM `isu_log` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC", jiaIsuUUID)
	if err != nil || !rows.Next() {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select")
	}
	//分以下を切り捨て、一時間単位にする関数
	truncateAfterHours := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	}
	//スコア計算をする関数
	calculateGraph := func(isuLogCluster []IsuLog) (*GraphData, error) {
		graph := new(GraphData)

		//sitting
		var sittingCount int = 0
		for _, log := range isuLogCluster {
			if log.IsSitting {
				sittingCount += 1
			}
		}
		graph.Sitting = sittingCount * 100 / len(isuLogCluster)

		//score&detail
		graph.Score = 100
		graph.Detail = map[string]int{}
		graph.Detail["is_dirty"] = 0
		graph.Detail["is_overweight"] = 0
		graph.Detail["is_broken"] = 0
		for _, log := range isuLogCluster {
			var conditoin IsuCondition
			err := json.Unmarshal([]byte(log.Condition), &conditoin)
			if err != nil {
				return nil, err
			}
			if conditoin.IsDirty {
				graph.Score += scorePerCondition["is_dirty"]
				graph.Detail["is_dirty"] += scorePerCondition["is_dirty"]
			}
			if conditoin.IsOverweight {
				graph.Score += scorePerCondition["is_overweight"]
				graph.Detail["is_overweight"] += scorePerCondition["is_overweight"]
			}
			if conditoin.IsBroken {
				graph.Score += scorePerCondition["is_broken"]
				graph.Detail["is_broken"] += scorePerCondition["is_broken"]
			}
		}
		if graph.Detail["is_dirty"] == 0 {
			delete(graph.Detail, "is_dirty")
		}
		if graph.Detail["is_overweight"] == 0 {
			delete(graph.Detail, "is_overweight")
		}
		if graph.Detail["is_broken"] == 0 {
			delete(graph.Detail, "is_broken")
		}

		return graph, nil
	}
	//スコアをDBに保存する関数
	saveGraph := func(tx *sqlx.Tx, isuLogCluster []IsuLog, startTime time.Time) error {
		data, err := calculateGraph(isuLogCluster)
		if err != nil {
			return fmt.Errorf("failed to calculate graph: %v", err)
		}
		dataStr, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to encode json: %v", err)
		}

		//insert
		_, err = tx.Exec("INSERT INTO `graph` (`jia_isu_uuid`, `start_at`, `data`) VALUES (?, ?, ?)"+
			"	ON DUPLICATE KEY UPDATE `data` = VALUES(`data`)",
			jiaIsuUUID, startTime, dataStr,
		)
		if err != nil {
			return fmt.Errorf("failed to insert: %v", err)
		}
		return nil
	}
	//一時間ごとに区切る
	err = rows.StructScan(&tmpIsuLog)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select")
	}
	startTime := truncateAfterHours(tmpIsuLog.Timestamp)
	isuLogCluster = []IsuLog{tmpIsuLog}
	for rows.Next() {
		err = rows.StructScan(&tmpIsuLog)
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to select")
		}
		tmpTime := truncateAfterHours(tmpIsuLog.Timestamp)
		if startTime != tmpTime {
			//tmpTimeは次の一時間なので、それ以外を使ってスコア計算
			err = saveGraph(tx, isuLogCluster, startTime)
			if err != nil {
				c.Logger().Error(err)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to save graph")
			}

			//次の一時間の探索
			startTime = tmpTime
			isuLogCluster = isuLogCluster[:0]
		}
		isuLogCluster = append(isuLogCluster, tmpIsuLog)
	}
	//最後の一時間分
	err = saveGraph(tx, isuLogCluster, startTime)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save graph")
	}

	// トランザクション終了
	err = tx.Commit()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit tx")
	}

	return c.NoContent(http.StatusCreated)
}
