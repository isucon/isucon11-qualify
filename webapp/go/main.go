package main

import (
	"bytes"
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	sessionName            = "isucondition"
	searchLimit            = 20
	conditionLimit         = 20
	isuListLimit           = 200 // TODO 修正が必要なら変更
	frontendContentsPath   = "../frontend/dist"
	jwtVerificationKeyPath = "../ec256-public.pem"
	defaultIconFilePath    = "../NoImage.png"
	defaultJIAServiceURL   = "http://localhost:5000"
)

var scorePerCondition = map[string]int{
	"is_dirty":      -1,
	"is_overweight": -1,
	"is_broken":     -5,
}

//"is_dirty=true/false,is_overweight=true/false,..."
var conditionFormat = regexp.MustCompile(`^[-a-zA-Z_]+=(true|false)(,[-a-zA-Z_]+=(true|false))*$`)

var (
	templates           *template.Template
	db                  *sqlx.DB
	sessionStore        sessions.Store
	mySQLConnectionData *MySQLConnectionEnv

	jwtVerificationKey *ecdsa.PublicKey

	isuConditionIP string
)

type Config struct {
	Name string `db:"name"`
	URL  string `db:"url"`
}

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

type IsuFromJIA struct {
	JIACatalogID string `json:"catalog_id"`
	Character    string `json:"character"`
}

type CatalogFromJIA struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LimitWeight int    `json:"limit_weight"`
	Weight      int    `json:"weight"`
	Size        string `json:"size"`
	Maker       string `json:"maker"`
	Features    string `json:"features"`
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

type IsuCondition struct {
	JIAIsuUUID string    `db:"jia_isu_uuid"`
	Timestamp  time.Time `db:"timestamp"`
	IsSitting  bool      `db:"is_sitting"`
	Condition  string    `db:"condition"`
	Message    string    `db:"message"`
	CreatedAt  time.Time `db:"created_at"`
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

type PostIsuRequest struct {
	JIAIsuUUID string `json:"jia_isu_uuid"`
	IsuName    string `json:"isu_name"`
}

type PutIsuRequest struct {
	Name string `json:"name"`
}

type GraphResponse struct {
	StartAt int64      `json:"start_at"`
	EndAt   int64      `json:"end_at"`
	Data    *GraphData `json:"data"`
}

type GetIsuConditionResponse struct {
	JIAIsuUUID     string `json:"jia_isu_uuid"`
	IsuName        string `json:"isu_name"`
	Timestamp      int64  `json:"timestamp"`
	IsSitting      bool   `json:"is_sitting"`
	Condition      string `json:"condition"`
	ConditionLevel string `json:"condition_level"`
	Message        string `json:"message"`
}

type PostIsuConditionRequest struct {
	IsSitting bool   `json:"is_sitting"`
	Condition string `json:"condition"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type JIAServiceRequest struct {
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	IsuUUID    string `json:"isu_uuid"`
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

	templates = template.Must(template.ParseFiles(
		frontendContentsPath + "/index.html",
	))

	key, err := ioutil.ReadFile(jwtVerificationKeyPath)
	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}
	jwtVerificationKey, err = jwt.ParseECPublicKeyFromPEM(key)
	if err != nil {
		log.Fatalf("Unable to parse ECDSA public key: %v", err)
	}
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
	// API for User
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
	// API for Isu
	e.POST("/api/isu/:jia_isu_uuid/condition", postIsuCondition)
	// Frontend
	e.GET("/", getIndex)
	e.GET("/condition", getIndex)
	e.GET("/search", getIndex)
	e.GET("/isu/:jia_isu_uuid", getIndex)
	e.GET("/register", getIndex)
	e.GET("/login", getIndex)
	// Assets
	e.Static("/assets", frontendContentsPath + "/assets")

	mySQLConnectionData = NewMySQLConnectionEnv()

	var err error
	db, err = mySQLConnectionData.ConnectDB()
	if err != nil {
		e.Logger.Fatalf("DB connection failed : %v", err)
		return
	}
	db.SetMaxOpenConns(10)
	defer db.Close()

	isuConditionIP = os.Getenv("ISU_CONDITION_IP")
	if isuConditionIP == "" {
		e.Logger.Fatalf("env ver ISU_CONDITION_IP is missing: %v", err)
		return
	}

	// Start server
	serverPort := fmt.Sprintf(":%v", getEnv("SERVER_PORT", "3000"))
	e.Logger.Fatal(e.Start(serverPort))
}

func getSession(r *http.Request) *sessions.Session {
	session, _ := sessionStore.Get(r, sessionName)
	return session
}

func getUserIDFromSession(r *http.Request) (string, error) {
	session := getSession(r)
	userID, ok := session.Values["jia_user_id"]
	if !ok {
		return "", fmt.Errorf("no session")
	}
	return userID.(string), nil
}

func getJIAServiceURL() string {
	config := Config{}
	err := db.Get(&config, "SELECT * FROM `isu_association_config` WHERE `name` = ?", "jia_service_url")
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Print(err)
		}
		return defaultJIAServiceURL
	}
	return config.URL
}

func getIndex(c echo.Context) error {
	err := templates.ExecuteTemplate(c.Response().Writer, "index.html", struct{}{})
	if err != nil {
		c.Logger().Errorf("getIndex error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return nil
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
	reqJwt := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	// verify JWT
	token, err := jwt.Parse(reqJwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, jwt.NewValidationError(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]), jwt.ValidationErrorSignatureInvalid)
		}
		return jwtVerificationKey, nil
	})
	if err != nil {
		switch err.(type) {
		case *jwt.ValidationError:
			c.Logger().Errorf("jwt validation error: %v", err)
			return c.String(http.StatusForbidden, "forbidden")
		default:
			c.Logger().Errorf("unknown error: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	// get jia_user_id from JWT Payload
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Logger().Errorf("type assertion error")
		return c.NoContent(http.StatusInternalServerError)
	}
	jiaUserIDVar, ok := claims["jia_user_id"]
	if !ok {
		c.Logger().Errorf("invalid JWT payload")
		return c.String(http.StatusBadRequest, "invalid JWT payload")
	}
	jiaUserID, ok := jiaUserIDVar.(string)
	if !ok {
		c.Logger().Errorf("invalid JWT payload")
		return c.String(http.StatusBadRequest, "invalid JWT payload")
	}

	_, err = db.Exec("INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)", jiaUserID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	session := getSession(c.Request())
	session.Values["jia_user_id"] = jiaUserID
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		c.Logger().Errorf("failed to set cookie: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

//  POST /api/signout
func postSignout(c echo.Context) error {
	_, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	session := getSession(c.Request())
	session.Options = &sessions.Options{MaxAge: -1, Path: "/"}
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		c.Logger().Errorf("cannot delete session: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// TODO
// GET /api/user/{jia_user_id}
// ユーザ情報を取得
// day2 実装のため skip
// func getUser(c echo.Context) error {
// }

func getMe(c echo.Context) error {
	userID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	res := GetMeResponse{JIAUserID: userID}
	return c.JSON(http.StatusOK, res)
}

// GET /api/catalog/{jia_catalog_id}
// カタログ情報を取得
func getCatalog(c echo.Context) error {
	// ログインチェック
	_, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	jiaCatalogID := c.Param("jia_catalog_id")
	if jiaCatalogID == "" {
		c.Logger().Errorf("jia_catalog_id is missing")
		return c.String(http.StatusBadRequest, "jia_catalog_id is missing")
	}

	// 日本ISU協会に問い合わせる(http request)
	catalogFromJIA, statusCode, err := fetchCatalogFromJIA(jiaCatalogID)
	if err != nil {
		c.Logger().Errorf("failed to get catalog from JIA: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if statusCode == http.StatusNotFound {
		c.Logger().Errorf("invalid jia_catalog_id")
		return c.String(http.StatusNotFound, "invalid jia_catalog_id")
	}

	catalog, err := castCatalogFromJIA(catalogFromJIA)
	if err != nil {
		c.Logger().Errorf("failed to cast catalog: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, catalog)
}

// 日本ISU協会にカタログ情報を問い合わせる
// 日本ISU協会のAPIについては資料を参照
func fetchCatalogFromJIA(catalogID string) (*CatalogFromJIA, int, error) {
	targetURLStr := getJIAServiceURL() + "/api/catalog/" + catalogID
	res, err := http.Get(targetURLStr)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, nil
	}

	catalog := &CatalogFromJIA{}
	err = json.NewDecoder(res.Body).Decode(&catalog)
	if err != nil {
		return nil, http.StatusOK, err
	}
	return catalog, http.StatusOK, nil
}

//CatalogFromJIAからCatalogへのキャスト
func castCatalogFromJIA(catalogFromJIA *CatalogFromJIA) (*Catalog, error) {
	//TODO: バリデーション
	catalog := &Catalog{}
	catalog.JIACatalogID = catalogFromJIA.ID
	catalog.Name = catalogFromJIA.Name
	catalog.LimitWeight = catalogFromJIA.LimitWeight
	catalog.Weight = catalogFromJIA.Weight
	catalog.Size = catalogFromJIA.Size
	catalog.Maker = catalogFromJIA.Maker
	catalog.Tags = catalogFromJIA.Features
	return catalog, nil
}

func getIsuList(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	limitStr := c.QueryParam("limit")
	limit := isuListLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.Logger().Errorf("invalid value: limit: (limit = %v) %v", limit, err)
			return c.String(http.StatusBadRequest, "invalid value: limit")
		}
	}

	isuList := []Isu{}
	err = db.Select(
		&isuList,
		"SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `is_deleted` = false ORDER BY `created_at` DESC LIMIT ?",
		jiaUserID, limit)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, isuList)
}

//  POST /api/isu
// 自分のISUの登録
func postIsu(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	var req PostIsuRequest
	err = c.Bind(&req)
	if err != nil {
		c.Logger().Errorf("bad request body: %v", err)
		return c.String(http.StatusBadRequest, "bad request body")
	}

	jiaIsuUUID := req.JIAIsuUUID
	isuName := req.IsuName

	// 既に登録されたisuでないか確認
	var count int
	// TODO 再activate時もエラー起こすため、 `is_deleted` は見ない
	err = db.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count != 0 {
		// TODO 再activate時もここでエラー; day2で再検討
		c.Logger().Errorf("duplicated isu: %v", err)
		return c.String(http.StatusConflict, "duplicated isu")
	}

	// JIAにisuのactivateをリクエスト
	targetURL := getJIAServiceURL() + "/api/activate"
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "3000"))
	if err != nil {
		c.Logger().Errorf("bad port number: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	body := JIAServiceRequest{isuConditionIP, port, jiaIsuUUID}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		c.Logger().Errorf("failed to marshal data: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	reqJIA, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		c.Logger().Errorf("failed to build request: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	reqJIA.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(reqJIA)
	if err != nil {
		c.Logger().Errorf("failed to request to JIAService: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		c.Logger().Errorf("JIAService returned error: status code %v", res.StatusCode)
		return c.String(res.StatusCode, "JIAService returned error")
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.Logger().Errorf("error occured while reading JIA response: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var isuFromJIA IsuFromJIA
	err = json.Unmarshal(resBody, &isuFromJIA)
	if err != nil {
		c.Logger().Errorf("cannot unmarshal JIA response: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// デフォルト画像を準備
	image, err := ioutil.ReadFile(defaultIconFilePath)
	if err != nil {
		c.Logger().Errorf("failed to read default icon: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// 新しいisuのデータをinsert
	_, err = db.Exec("INSERT INTO `isu`"+
		"	(`jia_isu_uuid`, `name`, `image`, `character`, `jia_catalog_id`, `jia_user_id`) VALUES (?, ?, ?, ?, ?, ?)",
		jiaIsuUUID, isuName, image, isuFromJIA.Character, isuFromJIA.JIACatalogID, jiaUserID)
	if err != nil {
		c.Logger().Errorf("cannot insert record: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var isu Isu
	err = db.Get(
		&isu,
		"SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, isu)
}

// GET /api/isu/search
// 自分の ISU 一覧から検索
func getIsuSearch(c echo.Context) error {
	// input (query_param)
	//  * name
	//  * character
	//	* catalog_name
	//	* min_limit_weight
	//	* max_limit_weight
	//	* catalog_tags (ジェイウォーク)
	//  * page: （default = 1)
	//	* TODO: day2 二つのカラムを併せて計算した値での検索（x*yの面積での検索とか）

	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}
	//optional query param
	name := c.QueryParam("name")
	character := c.QueryParam("character")
	catalogName := c.QueryParam("catalog_name")
	minLimitWeightStr := c.QueryParam("min_limit_weight")
	maxLimitWeightStr := c.QueryParam("max_limit_weight")
	catalogTagsStr := c.QueryParam("catalog_tags")
	pageStr := c.QueryParam("page")
	//parse
	minLimitWeight := math.MinInt32
	maxLimitWeight := math.MaxInt32
	page := 1
	var requiredCatalogTags []string
	if minLimitWeightStr != "" {
		minLimitWeight, err = strconv.Atoi(minLimitWeightStr)
		if err != nil {
			c.Logger().Errorf("bad format: min_limit_weight: %v", err)
			return c.String(http.StatusBadRequest, "bad format: min_limit_weight")
		}
	}
	if maxLimitWeightStr != "" {
		maxLimitWeight, err = strconv.Atoi(maxLimitWeightStr)
		if err != nil {
			c.Logger().Errorf("bad format: max_limit_weight: %v", err)
			return c.String(http.StatusBadRequest, "bad format: max_limit_weight")
		}
	}
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			c.Logger().Errorf("bad format: page : page = %v,  %v", page, err)
			return c.String(http.StatusBadRequest, "bad format: page")
		}
	}
	if catalogTagsStr != "" {
		requiredCatalogTags = strings.Split(catalogTagsStr, ",")
	}

	// DBからISU一覧を取得
	query := "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `is_deleted` = false"
	queryParam := []interface{}{jiaUserID}
	if name != "" {
		query += " AND `name` LIKE ? "
		queryParam = append(queryParam, name)
	}
	if character != "" {
		query += " AND `character` = ? "
		queryParam = append(queryParam, character)
	}
	query += " ORDER BY `created_at` DESC "
	isuList := []*Isu{}
	err = db.Select(&isuList,
		query,
		queryParam...,
	)
	if err != nil {
		c.Logger().Errorf("failed to select: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	isuListResponse := []*Isu{}
	//残りのフィルター処理
	for _, isu := range isuList {
		// ISU協会に問い合わせ(http request)
		catalogFromJIA, statusCode, err := fetchCatalogFromJIA(isu.JIACatalogID)
		if err != nil {
			c.Logger().Errorf("failed to fetch catalog from JIA: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		if statusCode == http.StatusNotFound {
			c.Logger().Errorf("catalog not found: %s", isu.JIACatalogID)
			return c.NoContent(http.StatusInternalServerError)
		}
		catalog, err := castCatalogFromJIA(catalogFromJIA)
		if err != nil {
			c.Logger().Errorf("failed to cast catalog: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		//catalog.Tagsをmapに変換
		catalogTags := map[string]interface{}{}
		for _, tag := range strings.Split(catalog.Tags, ",") {
			catalogTags[tag] = struct{}{}
		}

		// isu_catalog情報を使ってフィルター
		if !(minLimitWeight <= catalog.LimitWeight && catalog.LimitWeight <= maxLimitWeight) {
			continue
		}
		if catalogName != "" && catalogName != catalog.Name {
			continue
		}
		if requiredCatalogTags != nil {
			containsAll := true
			for _, tag := range requiredCatalogTags {
				if _, ok := catalogTags[tag]; !ok {
					containsAll = false
					break
				}
			}
			if !containsAll {
				continue
			}
		}

		//条件一致したのでレスポンスに追加
		isuListResponse = append(isuListResponse, isu)
	}

	//offset
	if pageStr != "" {
		offset := searchLimit * (page - 1)
		if len(isuListResponse) < offset {
			offset = len(isuListResponse)
		}
		isuListResponse = isuListResponse[offset:]
	}
	//limit
	if searchLimit < len(isuListResponse) {
		isuListResponse = isuListResponse[:searchLimit]
	}
	return c.JSON(http.StatusOK, isuListResponse)
}

//  GET /api/isu/{jia_isu_uuid}
// 椅子の情報を取得する
func getIsu(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")

	// TODO: jia_user_id 判別はクエリに入れずその後のロジックとする？ (一通り完成した後に要考慮)
	var isu Isu
	err = db.Get(&isu, "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		jiaUserID, jiaIsuUUID)
	if errors.Is(err, sql.ErrNoRows) {
		c.Logger().Errorf("isu not found: %v", err)
		return c.String(http.StatusNotFound, "isu not found")
	}
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, isu)
}

//  PUT /api/isu/{jia_isu_uuid}
// 自分の所有しているISUの情報を変更する
func putIsu(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")

	var req PutIsuRequest
	err = c.Bind(&req)
	if err != nil {
		c.Logger().Errorf("invalid request: %v", err)
		return c.String(http.StatusBadRequest, "invalid request")
	}

	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count == 0 {
		c.Logger().Errorf("isu not found")
		return c.String(http.StatusNotFound, "isu not found")
	}

	_, err = tx.Exec("UPDATE `isu` SET `name` = ? WHERE `jia_isu_uuid` = ?", req.Name, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var isu Isu
	err = tx.Get(&isu, "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, isu)
}

//  DELETE /api/isu/{jia_isu_uuid}
// 所有しているISUを削除する
func deleteIsu(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")

	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var count int
	err = tx.Get(
		&count,
		"SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count == 0 {
		c.Logger().Errorf("isu not found")
		return c.String(http.StatusNotFound, "isu not found")
	}

	_, err = tx.Exec("UPDATE `isu` SET `is_deleted` = true WHERE `jia_isu_uuid` = ?", jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// JIAにisuのdeactivateをリクエスト
	targetURL := getJIAServiceURL() + "/api/deactivate"
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "3000"))
	if err != nil {
		c.Logger().Errorf("bad port number: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	body := JIAServiceRequest{isuConditionIP, port, jiaIsuUUID}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		c.Logger().Errorf("failed to marshal data: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		c.Logger().Errorf("failed to build request: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Logger().Errorf("failed to request to JIAService: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		c.Logger().Errorf("JIAService returned error: status code %v", res.StatusCode)
		return c.NoContent(http.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("failed to commit tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

//  GET /api/isu/{jia_isu_uuid}/icon
// ISUのアイコンを取得する
// MEMO: ヘッダーとかでキャッシュ効くようにするのが想定解？(ただしPUTはあることに注意)
//       nginxで認証だけ外部に投げるみたいなのもできるっぽい？（ちゃんと読んでいない）
//       https://tech.jxpress.net/entry/2018/08/23/104123
// MEMO: DB 内の image は longblob
func getIsuIcon(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")

	var image []byte
	err = db.Get(&image, "SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Logger().Errorf("isu not found: %v", err)
			return c.String(http.StatusNotFound, "isu not found")
		}
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// TODO: putIsuIconでjpgも受け付けるなら対応が必要
	return c.Blob(http.StatusOK, "image/png", image)
}

//  PUT /api/isu/{jia_isu_uuid}/icon
// ISUのアイコンを登録する
// multipart/form-data
// MEMO: DB 内の image は longblob
func putIsuIcon(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")
	fh, err := c.FormFile("image")
	if err != nil {
		c.Logger().Errorf("jia_isu_uuid is invalid: %v", err)
		return c.String(http.StatusBadRequest, "jia_isu_uuid is invalid")
	}

	contentType := fh.Header.Get("Content-Type")
	// TODO: jpeg画像も受け付けるなら対応必要
	if contentType != "image/png" {
		c.Logger().Errorf("invalid image format: %v", contentType)
		return c.String(http.StatusBadRequest, "invalid image format")
	}

	file, err := fh.Open()
	if err != nil {
		c.Logger().Errorf("failed to open fh: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer file.Close()

	image, err := ioutil.ReadAll(file)
	if err != nil {
		c.Logger().Errorf("failed to read file: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false FOR UPDATE",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count == 0 {
		c.Logger().Errorf("isu not found")
		return c.String(http.StatusNotFound, "isu not found")
	}

	_, err = tx.Exec("UPDATE `isu` SET `image` = ? WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		image, jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

//  GET /api/isu/{jia_isu_uuid}/graph
// グラフ描画のための情報を計算して返却する
// ユーザーがISUの機嫌を知りたい
// この時間帯とか、この日とかの機嫌を知りたい
// 日毎時間単位グラフ
// conditionを何件か集めて、ISUにとっての快適度数みたいな値を算出する
// TODO: 文面の変更
func getIsuGraph(c echo.Context) error {
	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")
	dateStr := c.QueryParam("date")
	if dateStr == "" {
		c.Logger().Errorf("date is required")
		return c.String(http.StatusBadRequest, "date is required")
	}
	dateInt64, err := strconv.ParseInt(dateStr, 10, 64)
	if err != nil {
		c.Logger().Errorf("date is invalid format")
		return c.String(http.StatusBadRequest, "date is invalid format")
	}
	date := time.Unix(dateInt64, 0)

	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ? AND `is_deleted` = false",
		jiaUserID, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count == 0 {
		c.Logger().Errorf("isu not found")
		return c.String(http.StatusNotFound, "isu not found")
	}

	var graphList []Graph
	err = tx.Select(&graphList, "SELECT * FROM `graph` WHERE `jia_isu_uuid` = ? AND ? <= `start_at` AND `start_at` <= ? ORDER BY `start_at` ASC ",
		jiaIsuUUID, date, date.Add(time.Hour*24))
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	res := []GraphResponse{}
	index := 0
	tmpTime := date
	var tmpGraph Graph

	// dateから24時間分のグラフ用データを1時間単位で作成
	for tmpTime.Before(date.Add(time.Hour * 24)) {
		inRange := index < len(graphList)
		if inRange {
			tmpGraph = graphList[index]
		}

		var data *GraphData
		if inRange && tmpGraph.StartAt.Equal(tmpTime) {
			err = json.Unmarshal([]byte(tmpGraph.Data), &data)
			if err != nil {
				c.Logger().Errorf("failed to unmarshal json: %v", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			index++
		}

		graphResponse := GraphResponse{
			StartAt: tmpTime.Unix(),
			EndAt:   tmpTime.Add(time.Hour).Unix(),
			Data:    data,
		}
		res = append(res, graphResponse)
		tmpTime = tmpTime.Add(time.Hour)
	}

	// TODO: 必要以上に長めにトランザクションを取っているので後で検討
	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, res)
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

	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}
	sessionCookie, err := c.Cookie(sessionName)
	if err != nil {
		c.Logger().Errorf("failed to http request: %v", err)
		return c.String(http.StatusBadRequest, "cookie is missing")
	}
	//required query param
	cursorEndTimeInt64, err := strconv.ParseInt(c.QueryParam("cursor_end_time"), 10, 64)
	if err != nil {
		c.Logger().Errorf("bad format: cursor_end_time: %v", err)
		return c.String(http.StatusBadRequest, "bad format: cursor_end_time")
	}
	cursorEndTime := time.Unix(cursorEndTimeInt64, 0)

	cursorJIAIsuUUID := c.QueryParam("cursor_jia_isu_uuid")
	if cursorJIAIsuUUID == "" {
		c.Logger().Errorf("cursor_jia_isu_uuid is missing")
		return c.String(http.StatusBadRequest, "cursor_jia_isu_uuid is missing")
	}
	cursor := &GetIsuConditionResponse{
		JIAIsuUUID: cursorJIAIsuUUID,
		Timestamp:  cursorEndTime.Unix(),
	}
	conditionLevel := c.QueryParam("condition_level")
	if conditionLevel == "" {
		c.Logger().Errorf("condition_level is missing")
		return c.String(http.StatusBadRequest, "condition_level is missing")
	}
	//optional query param
	startTimeStr := c.QueryParam("start_time")
	if startTimeStr != "" {
		_, err = strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			c.Logger().Errorf("bad format: start_time: %v", err)
			return c.String(http.StatusBadRequest, "bad format: start_time")
		}
	}

	limitStr := c.QueryParam("limit")
	limit := conditionLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.Logger().Errorf("bad format: limit: limit = %v, %v", limit, err)
			return c.String(http.StatusBadRequest, "bad format: limit")
		}
	}

	// ユーザの所持椅子取得
	isuList := []Isu{}
	err = db.Select(&isuList,
		"SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `is_deleted` = false",
		jiaUserID,
	)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if len(isuList) == 0 {
		return c.JSON(http.StatusOK, isuList)
	}

	// ユーザの所持椅子毎に /api/condition/{jia_isu_uuid} を叩く
	conditionsResponse := []*GetIsuConditionResponse{}
	for _, isu := range isuList {
		//cursorのjia_isu_uuidで決まる部分は、とりあえず全部取得しておく
		//  cursorEndTime >= timestampを取りたいので、
		//  cursorEndTime + 1sec > timestampとしてリクエストを送る
		//この一要素はフィルターにかかるかどうか分からないので、limitも+1しておく
		conditionsTmp, err := getIsuConditionsFromLocalhost(
			isu.JIAIsuUUID, strconv.FormatInt(cursorEndTime.Add(1*time.Second).Unix(), 10),
			conditionLevel, startTimeStr, strconv.Itoa(limit+1),
			sessionCookie,
		)
		if err != nil {
			c.Logger().Errorf("failed to http request: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		// ユーザの所持椅子ごとのデータをマージ
		conditionsResponse = append(conditionsResponse, conditionsTmp...)
	}

	// (`timestamp`, `jia_isu_uuid`)のペアで降順ソート
	sort.Slice(conditionsResponse, func(i int, j int) bool { return conditionGreaterThan(conditionsResponse[i], conditionsResponse[j]) })
	// (cursor_end_time, cursor_jia_isu_uuid) > (`timestamp`, `jia_isu_uuid`)でフィルター
	removeIndex := 0
	for removeIndex < len(conditionsResponse) {
		if conditionGreaterThan(cursor, conditionsResponse[removeIndex]) {
			break
		}
		removeIndex++
	}
	//[0,index)は「(cursor_end_time, cursor_jia_isu_uuid) > (`timestamp`, `jia_isu_uuid`)」を満たしていないので取り除く
	conditionsResponse = conditionsResponse[removeIndex:]

	//limitを取る
	if len(conditionsResponse) > limit {
		conditionsResponse = conditionsResponse[:limit]
	}

	return c.JSON(http.StatusOK, conditionsResponse)
}

//http requestを飛ばし、そのレスポンスを[]GetIsuConditionResponseに変換する
func getIsuConditionsFromLocalhost(
	jiaIsuUUID string, cursorEndTimeStr string, conditionLevel string, startTimeStr string, limitStr string,
	cookie *http.Cookie,
) ([]*GetIsuConditionResponse, error) {
	targetURLStr := fmt.Sprintf(
		"http://localhost:%s/api/condition/%s",
		getEnv("SERVER_PORT", "3000"), jiaIsuUUID,
	)
	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %v ;(%s,%s)", err, getEnv("SERVER_PORT", "3000"), jiaIsuUUID)
	}

	q := url.Values{}
	q.Set("cursor_end_time", cursorEndTimeStr)
	q.Set("condition_level", conditionLevel)
	if startTimeStr != "" {
		q.Set("start_time", startTimeStr)
	}
	if limitStr != "" {
		q.Set("limit", limitStr)
	}
	targetURL.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, targetURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to `GET %s` with status=`%s`", targetURL.String(), res.Status)
	}

	condition := []*GetIsuConditionResponse{}
	err = json.NewDecoder(res.Body).Decode(&condition)
	if err != nil {
		return nil, err
	}
	return condition, nil
}

// left > right を計算する関数
func conditionGreaterThan(left *GetIsuConditionResponse, right *GetIsuConditionResponse) bool {
	//(`timestamp`, `jia_isu_uuid`)のペアを辞書順に比較

	if left.Timestamp > right.Timestamp {
		return true
	}
	if left.Timestamp == right.Timestamp {
		return left.JIAIsuUUID > right.JIAIsuUUID
	}
	return false
}

//  GET /api/condition/{jia_isu_uuid}?
// 自分の所持椅子のうち、指定した椅子の通知を取得する
func getIsuConditions(c echo.Context) error {
	// input
	//     * jia_isu_uuid: 椅子の固有番号(path_param)
	//     * start_time: 開始時間
	//     * cursor_end_time: 終了時間 (required)
	//     * condition_level: critical,warning,info (csv)
	//               critical: conditions (is_dirty,is_overweight,is_broken) のうちtrueが3個
	//               warning: conditionsのうちtrueのものが1 or 2個
	//               info: warning無し
	//     * TODO: day2実装: message (文字列検索)

	jiaUserID, err := getUserIDFromSession(c.Request())
	if err != nil {
		c.Logger().Errorf("you are not signed in: %v", err)
		return c.String(http.StatusUnauthorized, "you are not signed in")
	}
	jiaIsuUUID := c.Param("jia_isu_uuid")
	if jiaIsuUUID == "" {
		c.Logger().Errorf("jia_isu_uuid is missing")
		return c.String(http.StatusBadRequest, "jia_isu_uuid is missing")
	}
	//required query param
	cursorEndTimeInt64, err := strconv.ParseInt(c.QueryParam("cursor_end_time"), 10, 64)
	if err != nil {
		c.Logger().Errorf("bad format: cursor_end_time: %v", err)
		return c.String(http.StatusBadRequest, "bad format: cursor_end_time")
	}
	cursorEndTime := time.Unix(cursorEndTimeInt64, 0)
	conditionLevelCSV := c.QueryParam("condition_level")
	if conditionLevelCSV == "" {
		c.Logger().Errorf("condition_level is missing")
		return c.String(http.StatusBadRequest, "condition_level is missing")
	}
	conditionLevel := map[string]interface{}{}
	for _, level := range strings.Split(conditionLevelCSV, ",") {
		conditionLevel[level] = struct{}{}
	}
	//optional query param
	startTimeStr := c.QueryParam("start_time")
	startTime := time.Time{}
	if startTimeStr != "" {
		startTimeInt64, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			c.Logger().Errorf("bad format: start_time: %v", err)
			return c.String(http.StatusBadRequest, "bad format: start_time")
		}
		startTime = time.Unix(startTimeInt64, 0)
	}
	limitStr := c.QueryParam("limit")
	limit := conditionLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.Logger().Errorf("bad format: limit: limit = %v, %v", limit, err)
			return c.String(http.StatusBadRequest, "bad format: limit")
		}
	}

	// isu_id存在確認、ユーザの所持椅子か確認
	var isuName string
	err = db.Get(&isuName,
		"SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ? AND `is_deleted` = false",
		jiaIsuUUID, jiaUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		c.Logger().Errorf("isu not found: %v", err)
		return c.String(http.StatusNotFound, "isu not found")
	}
	if err != nil {
		c.Logger().Errorf("failed to select: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// 対象isu_idの通知を取得(limit, cursorで絞り込み）
	conditionsResponse, err := getIsuConditionsFromDB(jiaIsuUUID, cursorEndTime, conditionLevel, startTime, limit, isuName)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, conditionsResponse)
}

func getIsuConditionsFromDB(jiaIsuUUID string, cursorEndTime time.Time, conditionLevel map[string]interface{}, startTime time.Time,
	limit int, isuName string) ([]GetIsuConditionResponse, error) {

	conditions := []IsuCondition{}
	var err error

	if startTime.IsZero() {
		err = db.Select(&conditions,
			"SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?"+
				"	AND `timestamp` < ?"+
				"	ORDER BY `timestamp` DESC",
			jiaIsuUUID, cursorEndTime,
		)
	} else {
		err = db.Select(&conditions,
			"SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?"+
				"	AND `timestamp` < ?"+
				"	AND ? <= `timestamp`"+
				"	ORDER BY `timestamp` DESC",
			jiaIsuUUID, cursorEndTime, startTime,
		)
	}
	if err != nil {
		return nil, err
	}

	//condition_levelでの絞り込み
	conditionsResponse := []GetIsuConditionResponse{}
	for _, c := range conditions {
		var cLevel string
		warnCount := strings.Count(c.Condition, "=true")
		switch warnCount {
		case 0:
			cLevel = "info"
		case 1, 2:
			cLevel = "warning"
		case 3:
			cLevel = "critical"
		}

		if _, ok := conditionLevel[cLevel]; ok {
			//GetIsuConditionResponseに変換
			data := GetIsuConditionResponse{
				JIAIsuUUID:     c.JIAIsuUUID,
				IsuName:        isuName,
				Timestamp:      c.Timestamp.Unix(),
				IsSitting:      c.IsSitting,
				Condition:      c.Condition,
				ConditionLevel: cLevel,
				Message:        c.Message,
			}
			conditionsResponse = append(conditionsResponse, data)
		}
	}

	//limit
	if len(conditionsResponse) > limit {
		conditionsResponse = conditionsResponse[:limit]
	}

	return conditionsResponse, nil
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

	//TODO: 記法の統一
	jiaIsuUUID := c.Param("jia_isu_uuid")
	if jiaIsuUUID == "" {
		c.Logger().Errorf("jia_isu_uuid is missing")
		return c.String(http.StatusBadRequest, "jia_isu_uuid is missing")
	}
	var req PostIsuConditionRequest
	err := c.Bind(&req)
	if err != nil {
		c.Logger().Errorf("bad request body: %v", err)
		return c.String(http.StatusBadRequest, "bad request body")
	}

	//Parse
	timestamp := time.Unix(req.Timestamp, 0)

	if !conditionFormat.Match([]byte(req.Condition)) {
		c.Logger().Errorf("bad request body")
		return c.String(http.StatusBadRequest, "bad request body")
	}

	// トランザクション開始
	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	// jia_isu_uuid が存在するかを確認
	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?  and `is_deleted` = false", jiaIsuUUID) //TODO: 記法の統一
	if err != nil {
		c.Logger().Errorf("failed to select: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count == 0 {
		c.Logger().Errorf("isu not found")
		return c.String(http.StatusNotFound, "isu not found")
	}

	//isu_conditionに記録
	//confilct確認
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu_condition` WHERE (`timestamp`, `jia_isu_uuid`) = (?, ?)  FOR UPDATE", //TODO: 記法の統一
		timestamp, jiaIsuUUID,
	)
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count != 0 {
		c.Logger().Errorf("isu condition already exist")
		return c.String(http.StatusConflict, "isu_condition already exist")
	}
	//insert
	_, err = tx.Exec("INSERT INTO `isu_condition`"+
		"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`) VALUES (?, ?, ?, ?, ?)",
		jiaIsuUUID, timestamp, req.IsSitting, req.Condition, req.Message,
	)
	if err != nil {
		c.Logger().Errorf("failed to insert: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// getGraph用のデータを計算し、DBを更新する
	err = updateGraph(tx, jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("failed to update graph: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// トランザクション終了
	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("failed to commit tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

// getGraph用のデータを計算し、DBを更新する
func updateGraph(tx *sqlx.Tx, jiaIsuUUID string) error {
	// IsuConditionを一時間ごとの区切りに分け、区切りごとにスコアを計算する
	IsuConditionCluster := []IsuCondition{} // 一時間ごとの纏まり
	var tmpIsuCondition IsuCondition
	valuesForUpdate := []interface{}{} //3個1組、更新するgraphの各行のデータ
	rows, err := tx.Queryx("SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC", jiaIsuUUID)
	if err != nil {
		return err
	}
	//一時間ごとに区切る
	var startTime time.Time
	for rows.Next() {
		err = rows.StructScan(&tmpIsuCondition)
		if err != nil {
			return err
		}
		tmpTime := truncateAfterHours(tmpIsuCondition.Timestamp)
		if startTime != tmpTime {
			if len(IsuConditionCluster) > 0 {
				//tmpTimeは次の一時間なので、それ以外を使ってスコア計算
				data, err := calculateGraphData(IsuConditionCluster)
				if err != nil {
					return fmt.Errorf("failed to calculate graph: %v", err)
				}
				valuesForUpdate = append(valuesForUpdate, jiaIsuUUID, startTime, data)
			}

			//次の一時間の探索
			startTime = tmpTime
			IsuConditionCluster = []IsuCondition{}
		}
		IsuConditionCluster = append(IsuConditionCluster, tmpIsuCondition)
	}
	if len(IsuConditionCluster) > 0 {
		//最後の一時間分
		data, err := calculateGraphData(IsuConditionCluster)
		if err != nil {
			return fmt.Errorf("failed to calculate graph: %v", err)
		}
		valuesForUpdate = append(valuesForUpdate, jiaIsuUUID, startTime, data)
	}

	//insert or update
	params := strings.Repeat("(?,?,?),", len(valuesForUpdate)/3)
	params = params[:len(params)-1]
	_, err = tx.Exec("INSERT INTO `graph` (`jia_isu_uuid`, `start_at`, `data`) VALUES "+
		params+
		"	ON DUPLICATE KEY UPDATE `data` = VALUES(`data`)",
		valuesForUpdate...,
	)
	if err != nil {
		return err
	}

	return nil
}

//分以下を切り捨て、一時間単位にする関数
func truncateAfterHours(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

//スコア計算をする関数
func calculateGraphData(IsuConditionCluster []IsuCondition) ([]byte, error) {
	graph := &GraphData{}

	//sitting
	sittingCount := 0
	for _, log := range IsuConditionCluster {
		if log.IsSitting {
			sittingCount++
		}
	}
	graph.Sitting = sittingCount * 100 / len(IsuConditionCluster)

	//score&detail
	graph.Score = 100
	//condition要因の減点
	graph.Detail = map[string]int{}
	for key := range scorePerCondition {
		graph.Detail[key] = 0
	}
	for _, log := range IsuConditionCluster {
		conditions := map[string]bool{}
		//DB上にある is_dirty=true/false,is_overweight=true/false,... 形式のデータを
		//map[string]bool形式に変換
		for _, cond := range strings.Split(log.Condition, ",") {
			keyValue := strings.Split(cond, "=")
			if len(keyValue) != 2 {
				continue //形式に従っていないものは無視
			}
			conditions[keyValue[0]] = (keyValue[1] != "false")
		}

		//trueになっているものは減点
		for key, enabled := range conditions {
			if enabled {
				score, ok := scorePerCondition[key]
				if ok {
					graph.Score += score
					graph.Detail[key] += score
				}
			}
		}
	}
	//スコアに影響がないDetailを削除
	for key := range scorePerCondition {
		if graph.Detail[key] == 0 {
			delete(graph.Detail, key)
		}
	}
	//個数減点
	if len(IsuConditionCluster) < 50 {
		minus := -(50 - len(IsuConditionCluster)) * 2
		graph.Score += minus
		graph.Detail["missing_data"] = minus
	}
	if graph.Score < 0 {
		graph.Score = 0
	}

	//JSONに変換
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return nil, err
	}
	return graphJSON, nil
}
