//package infrastructure
//
//import (
//	"github.com/jinzhu/gorm"
//	"github.com/joho/godotenv"
//	"log"
//	"os"
//)
//
//func Connect() (db *gorm.DB, err error) {
//
//	err = godotenv.Load()
//
//	if err != nil {
//		log.Fatal("Error loading .env file")
//	}
//	// Memo:"DB_HOST"はdockerの場合データベースコンテナ名
//	db, err = gorm.Open("mysql",
//		os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+
//			"@tcp("+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+")/"+
//			os.Getenv("DB_DATABASE")+
//			"?charset=utf8mb4&parseTime=True&loc=Local")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return db, err
//}
package infrastructure

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

var DB *sql.DB

func init() {
	envErr := godotenv.Load()
	if envErr != nil {
		logrus.Fatal("Error loading .env")
	}
	var err error
	fmt.Println(os.Getenv("DB_PASSWORD"))
	fmt.Println(os.Getenv("LOG_FILE"))
	fmt.Println(os.Getenv("DB_USERNAME"))
	DB, err = sql.Open("mysql", os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+
		"@tcp("+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+")/"+
		os.Getenv("DB_DATABASE")+
		"?charset=utf8mb4&parseTime=True&loc=Local")

	if err != nil {
		log.Fatal(err)
	}
	//DB, err = sql.Open("mysql", "root:pass@tcp(db:3306)/systemtrade?charset=utf8mb4&parseTime=True&loc=Local")
	//if err != nil {
	//	log.Fatal(err)
	//}
}
