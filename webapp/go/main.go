package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	sessionName              = "isucondition"
	searchLimit              = 20
	conditionLimit           = 20
	conditionTimestampFormat = "2006-01-02 15:04:05 -0700"
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
	JIAUserID string `json:"jia_user_id"`
}

type GraphResponse struct {
}

type GetIsuConditionResponse struct {
	JIAIsuUUID     string    `json:"jia_isu_uuid"`
	IsuName        string    `json:"isu_name"`
	Timestamp      time.Time `json:"timestamp"`
	IsSitting      bool      `json:"is_sitting"`
	Condition      string    `json:"condition"`
	ConditionLevel string    `json:"condition_level"`
	Message        string    `json:"message"`
}

type PostIsuConditionRequest struct {
	IsSitting bool   `json:"is_sitting"`
	Condition string `json:"condition"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"` //Format("2006-01-02 15:04:05 -0700")
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
	e.GET("/api/condition", getAllIsuConditions)
	e.GET("/api/condition/:jia_isu_uuid", getIsuConditions)

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

func getMe(c echo.Context) error {
	userID, err := getUserIdFromSession(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "you are not signed in") // TODO 記法が決まったら修正
	}

	response := GetMeResponse{JIAUserID: userID}
	return c.JSON(http.StatusOK, response)
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
	jiaUserID, err := getUserIdFromSession(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "you are not sign in")
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")

	// TODO: jia_user_id 判別はクエリに入れずその後のロジックとする？ (一通り完成した後に要考慮)
	var isu Isu
	err = db.Get(&isu, "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = ?",
		jiaUserID, jiaIsuUUID, false)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, "isu not found")
	}
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "db error")
	}

	return c.JSON(http.StatusOK, isu)
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

//  GET /api/condition?
// 自分の所持椅子の通知を取得する
func getAllIsuConditions(c echo.Context) error {
	// input
	//     * start_time: 開始時間
	//     * cursor_end_time: 終了時間 (required)
	//     * cursor_jia_isu_uuid (required)
	//     * condition_level: critical,warning,info (csv)
	//               critical: conditionsのうちtrueが3個
	//               warning: conditionsのうちtrueのものが1 or 2個
	//               info: warning無し
	//     * TODO: day2実装: message (文字列検索)

	jiaUserID, err := getUserIdFromSession(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "you are not sign in")
	}
	//required query param
	cursorEndTimeStr := c.QueryParam("cursor_end_time")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "cursor_end_time is missing")
	}
	cursorJIAIsuUUID := c.QueryParam("cursor_jia_isu_uuid")
	if cursorJIAIsuUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "cursor_jia_isu_uuid is missing")
	}
	conditionLevel := c.QueryParam("condition_level")
	if conditionLevel == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "condition_level is missing")
	}
	//optional query param
	startTimeStr := c.QueryParam("start_time")

	// ユーザの所持椅子取得
	isuList := []Isu{}
	err = db.Select(&isuList,
		"SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `is_deleted` = false",
		jiaUserID,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			c.Logger().Errorf("failed to select: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	// ユーザの所持椅子毎に http://localhost:3000/api/condition/{jia_isu_uuid} を叩く
	conditionsResponse := []*GetIsuConditionResponse{}
	for _, isu := range isuList {
		conditionsTmp, err := getIsuConditionsFromLocalhost(
			isu.JIAIsuUUID, cursorEndTimeStr, cursorJIAIsuUUID, conditionLevel, startTimeStr,
		)
		if err != nil {
			c.Logger().Errorf("failed to http request: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		// ユーザの所持椅子ごとのデータをマージ
		conditionsResponse = append(conditionsResponse, conditionsTmp...)
	}

	// (`timestamp`, `jia_isu_uuid`)のペアで降順ソート
	sort.Slice(conditionsResponse, func(i int, j int) bool {
		// return [i] > [j]

		if conditionsResponse[i].Timestamp.After(conditionsResponse[j].Timestamp) {
			return true
		}
		if conditionsResponse[i].Timestamp.Equal(conditionsResponse[j].Timestamp) {
			return conditionsResponse[i].JIAIsuUUID > conditionsResponse[j].JIAIsuUUID
		}
		return false
	})

	//limitを取る
	conditionsResponse = conditionsResponse[:conditionLimit]

	return c.JSON(http.StatusOK, conditionsResponse)
}

//http requestを飛ばし、そのレスポンスを[]GetIsuConditionResponseに変換する
func getIsuConditionsFromLocalhost(
	jiaIsuUUID string, cursorEndTimeStr string, cursorJIAIsuUUID string, conditionLevel string, startTimeStr string,
) ([]*GetIsuConditionResponse, error) {

	targetURL, err := url.Parse(fmt.Sprintf(
		"http://localhost:%s/api/condition/%s",
		getEnv("SERVER_PORT", "3000"), jiaIsuUUID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %v ;(%s,%s)", err, getEnv("SERVER_PORT", "3000"), jiaIsuUUID)
	}

	q := targetURL.Query()
	q.Set("cursor_end_time", cursorEndTimeStr)
	q.Set("cursor_jia_isu_uuid", cursorJIAIsuUUID)
	q.Set("condition_level", conditionLevel)
	if startTimeStr != "" {
		q.Set("start_time", startTimeStr)
	}
	targetURL.RawQuery = q.Encode()

	res, err := http.Get(targetURL.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to `GET %s` with status=`%s`", targetURL.String(), res.Status)
	}

	condition := []*GetIsuConditionResponse{}
	err = json.NewDecoder(res.Body).Decode(&condition)
	if err != nil {
		return nil, err
	}
	return condition, nil
}

//  GET /api/condition/{jia_isu_uuid}?
// 自分の所持椅子のうち、指定した椅子の通知を取得する
func getIsuConditions(c echo.Context) error {
	// input
	//     * jia_isu_uuid: 椅子の固有番号(path_param)
	//     * start_time: 開始時間
	//     * cursor_end_time: 終了時間 (required)
	//     * cursor_jia_isu_uuid (required)
	//     * condition_level: critical,warning,info (csv)
	//               critical: conditions (is_dirty,is_overweight,is_broken) のうちtrueが3個
	//               warning: conditionsのうちtrueのものが1 or 2個
	//               info: warning無し
	//     * TODO: day2実装: message (文字列検索)

	jiaUserID, err := getUserIdFromSession(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "you are not sign in")
	}
	jiaIsuUUID := c.Param("jia_isu_uuid")
	if jiaIsuUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "jia_isu_uuid is missing")
	}
	//required query param
	cursorEndTime, err := time.Parse(conditionTimestampFormat, c.QueryParam("cursor_end_time"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad format: cursor_end_time")
	}
	cursorJIAIsuUUID := c.QueryParam("cursor_jia_isu_uuid")
	if cursorJIAIsuUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "cursor_jia_isu_uuid is missing")
	}
	conditionLevel := c.QueryParam("condition_level")
	if conditionLevel == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "condition_level is missing")
	}
	//optional query param
	startTimeStr := c.QueryParam("start_time")
	var startTime time.Time
	if startTimeStr != "" {
		startTime, err = time.Parse(conditionTimestampFormat, startTimeStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "bad format: cursor_end_time")
		}
	}

	// isu_id存在確認、ユーザの所持椅子か確認
	var isuName string
	err = db.Get(&isuName,
		"SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ? AND `is_deleted` = false",
		jiaIsuUUID, jiaUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, "isu not found")
	}
	if err != nil {
		c.Logger().Errorf("failed to select: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// 対象isu_idの通知を取得(limit, cursorで絞り込み）
	conditions := []IsuLog{}
	if startTimeStr == "" {
		err = db.Select(&conditions,
			"SELECT * FROM `isu_log` WHERE `jia_isu_uuid` = ?"+
				"	AND (`timestamp`, `jia_isu_uuid`) < (?, ?)"+
				"	ORDER BY `created_at` desc, `jia_isu_uuid` desc limit ?",
			jiaIsuUUID, cursorEndTime, cursorJIAIsuUUID, conditionLimit,
		)
	} else {
		err = db.Select(&conditions,
			"SELECT * FROM `isu_log` WHERE `jia_isu_uuid` = ?"+
				"	AND (`timestamp`, `jia_isu_uuid`) < (?, ?)"+
				"	AND ? <= `timestamp`"+
				"	ORDER BY `created_at` desc, `jia_isu_uuid` desc limit ?",
			jiaIsuUUID, cursorEndTime, cursorJIAIsuUUID, startTime, conditionLimit,
		)
	}
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			c.Logger().Errorf("failed to select: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	//condition_levelでの絞り込み
	conditionsResponse := []GetIsuConditionResponse{}
	for _, c := range conditions {

		var cLevel string
		add := false
		warnCount := strings.Count(c.Condition, "=true")
		if strings.Contains(conditionLevel, "critical") && warnCount == 3 {
			cLevel = "critical"
			add = true
		} else if strings.Contains(conditionLevel, "warning") && (warnCount == 1 || warnCount == 2) {
			cLevel = "warning"
			add = true
		} else if strings.Contains(conditionLevel, "info") && warnCount == 0 {
			cLevel = "info"
			add = true
		}

		//GetIsuConditionResponseに変換
		data := GetIsuConditionResponse{
			JIAIsuUUID:     c.JIAIsuUUID,
			IsuName:        isuName,
			Timestamp:      c.Timestamp,
			IsSitting:      c.IsSitting,
			Condition:      c.Condition,
			ConditionLevel: cLevel,
			Message:        c.Message,
		}
		if add {
			conditionsResponse = append(conditionsResponse, data)
		}
	}

	return c.JSON(http.StatusOK, conditionsResponse)
}

// POST /api/isu/{jia_isu_uuid}/condition
// ISUからのセンサデータを受け取る
func postIsuCondition(c echo.Context) error {
	// input (path_param)
	//	* jia_isu_uuid
	// input (body)
	//  * is_sitting:  true/false,
	// 	* condition: "is_dirty=true/false,is_overweight=true/false,..."
	//  * message
	//	* timestamp（秒まで）

	jiaIsuUUID := c.Param("jia_isu_uuid")
	if jiaIsuUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "jia_isu_uuid is missing")
	}
	var request PostIsuConditionRequest
	err := c.Bind(&request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	//Parse
	timestamp, err := time.Parse(conditionTimestampFormat, request.Timestamp)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid timestamp")
	}

	// トランザクション開始
	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	// jia_isu_uuid が存在するかを確認
	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?  and `is_deleted`=false", jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("failed to select: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if count == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "isu not found")
	}

	//isu_logに記録
	//confilct確認
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu_log` WHERE (`timestamp`, `jia_isu_uuid`) = (?, ?)  FOR UPDATE",
		timestamp, jiaIsuUUID,
	)
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if count != 0 {
		return echo.NewHTTPError(http.StatusConflict, "isu_log already exist")
	}
	//insert
	_, err = tx.Exec("INSERT INTO `isu_log`"+
		"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`) VALUES (?, ?, ?, ?, ?)",
		jiaIsuUUID, timestamp, request.IsSitting, request.Condition, request.Message,
	)
	if err != nil {
		c.Logger().Errorf("failed to insert: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert")
	}

	// getGraph用のデータを計算
	// IsuLogを一時間ごとの区切りに分け、区切りごとにスコアを計算する
	var isuLogCluster []IsuLog // 一時間ごとの纏まり
	var tmpIsuLog IsuLog
	var valuesForInsert []interface{} = []interface{}{}
	rows, err := tx.Queryx("SELECT * FROM `isu_log` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC", jiaIsuUUID)
	if err != nil || !rows.Next() {
		c.Logger().Errorf("failed to select: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	//一時間ごとに区切る
	err = rows.StructScan(&tmpIsuLog)
	if err != nil {
		c.Logger().Errorf("failed to select: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	startTime := truncateAfterHours(tmpIsuLog.Timestamp)
	isuLogCluster = []IsuLog{tmpIsuLog}
	for rows.Next() {
		err = rows.StructScan(&tmpIsuLog)
		if err != nil {
			c.Logger().Errorf("failed to select: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		tmpTime := truncateAfterHours(tmpIsuLog.Timestamp)
		if startTime != tmpTime {
			//tmpTimeは次の一時間なので、それ以外を使ってスコア計算
			data, err := calculateGraph(isuLogCluster)
			if err != nil {
				c.Logger().Errorf("failed to calculate graph: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			dataJSON, err := json.Marshal(data)
			if err != nil {
				c.Logger().Errorf("failed to encode json: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			valuesForInsert = append(valuesForInsert, jiaIsuUUID, startTime, dataJSON)

			//次の一時間の探索
			startTime = tmpTime
			isuLogCluster = isuLogCluster[:0]
		}
		isuLogCluster = append(isuLogCluster, tmpIsuLog)
	}
	//最後の一時間分
	data, err := calculateGraph(isuLogCluster)
	if err != nil {
		c.Logger().Errorf("failed to calculate graph: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		c.Logger().Errorf("failed to encode json: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	valuesForInsert = append(valuesForInsert, jiaIsuUUID, startTime, dataStr)
	//insert
	_, err = tx.Exec("INSERT INTO `graph` (`jia_isu_uuid`, `start_at`, `data`) VALUES "+
		"	(?,?,?)"+strings.Repeat(",(?,?,?)", len(valuesForInsert)/3-1)+
		"	ON DUPLICATE KEY UPDATE `data` = VALUES(`data`)",
		valuesForInsert...,
	)
	if err != nil {
		c.Logger().Errorf("failed to insert: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// トランザクション終了
	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("failed to commit tx: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

//分以下を切り捨て、一時間単位にする関数
func truncateAfterHours(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

//スコア計算をする関数
func calculateGraph(isuLogCluster []IsuLog) (*GraphData, error) {
	graph := &GraphData{}

	//sitting
	sittingCount := 0
	for _, log := range isuLogCluster {
		if log.IsSitting {
			sittingCount += 1
		}
	}
	graph.Sitting = sittingCount * 100 / len(isuLogCluster)

	//score&detail
	graph.Score = 100
	graph.Detail = map[string]int{}
	for key := range scorePerCondition {
		graph.Detail[key] = 0
	}
	for _, log := range isuLogCluster {
		conditions := map[string]bool{}
		for _, cond := range strings.Split(log.Condition, ",") {
			keyValue := strings.Split(cond, "=")
			if len(keyValue) != 2 {
				return nil, fmt.Errorf("invalid condition %s", cond)
			}
			if _, ok := scorePerCondition[keyValue[0]]; !ok {
				return nil, fmt.Errorf("invalid condition %s", cond)
			}
			conditions[keyValue[0]] = (keyValue[1] != "false")
		}

		for key, enabled := range conditions {
			if enabled {
				graph.Score += scorePerCondition[key]
				graph.Detail[key] += scorePerCondition[key]
			}
		}
	}
	for key := range scorePerCondition {
		if graph.Detail[key] == 0 {
			delete(graph.Detail, key)
		}
	}

	return graph, nil
}
