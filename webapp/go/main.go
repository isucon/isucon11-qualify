package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	// "github.com/labstack/echo"
	// "github.com/labstack/echo/middleware"
	// "github.com/labstack/gommon/log"
)

const NotificationLimit = 20

var db *sqlx.DB
var mySQLConnectionData *MySQLConnectionEnv

type InitializeResponse struct {
	Language string `json:"language"`
}

type User struct {
}

type Isu struct {
}

type IsuLog struct {
}

type Graph struct {
}

type MySQLConnectionEnv struct {
	Host     string
	Port     string
	User     string
	DBName   string
	Password string
}

func getEnv(key, defaultValue string) string {
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
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", mc.User, mc.Password, mc.Host, mc.Port, mc.DBName)
	return sqlx.Open("mysql", dsn)
}

func init() {
}

func main() {
	// // Echo instance
	// e := echo.New()
	// e.Debug = true
	// e.Logger.SetLevel(log.DEBUG)

	// // Middleware
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	// // Initialize
	// e.POST("/initialize", initialize)

	// // Chair Handler
	// e.GET("/api/chair/:id", getChairDetail)
	// e.POST("/api/chair", postChair)
	// e.GET("/api/chair/search", searchChairs)
	// e.GET("/api/chair/low_priced", getLowPricedChair)
	// e.GET("/api/chair/search/condition", getChairSearchCondition)
	// e.POST("/api/chair/buy/:id", buyChair)

	// // Estate Handler
	// e.GET("/api/estate/:id", getEstateDetail)
	// e.POST("/api/estate", postEstate)
	// e.GET("/api/estate/search", searchEstates)
	// e.GET("/api/estate/low_priced", getLowPricedEstate)
	// e.POST("/api/estate/req_doc/:id", postEstateRequestDocument)
	// e.POST("/api/estate/nazotte", searchEstateNazotte)
	// e.GET("/api/estate/search/condition", getEstateSearchCondition)
	// e.GET("/api/recommended_estate/:id", searchRecommendedEstateWithChair)

	mySQLConnectionData = NewMySQLConnectionEnv()

	var err error
	db, err = mySQLConnectionData.ConnectDB()
	if err != nil {
		//e.Logger.Fatalf("DB connection failed : %v", err)
		return
	}
	db.SetMaxOpenConns(10)
	defer db.Close()

	// Start server
	// serverPort := fmt.Sprintf(":%v", getEnv("SERVER_PORT", "3000"))
	// e.Logger.Fatal(e.Start(serverPort))
}

// func initialize(c echo.Context) error {
// 	sqlDir := filepath.Join("..", "mysql", "db")
// 	paths := []string{
// 		filepath.Join(sqlDir, "0_Schema.sql"),
// 	}

// 	for _, p := range paths {
// 		sqlFile, _ := filepath.Abs(p)
// 		cmdStr := fmt.Sprintf("mysql -h %v -u %v -p%v -P %v %v < %v",
// 			mySQLConnectionData.Host,
// 			mySQLConnectionData.User,
// 			mySQLConnectionData.Password,
// 			mySQLConnectionData.Port,
// 			mySQLConnectionData.DBName,
// 			sqlFile,
// 		)
// 		if err := exec.Command("bash", "-c", cmdStr).Run(); err != nil {
// 			c.Logger().Errorf("Initialize script error : %v", err)
// 			return c.NoContent(http.StatusInternalServerError)
// 		}
// 	}

// 	return c.JSON(http.StatusOK, InitializeResponse{
// 		Language: "go",
// 	})
// }
