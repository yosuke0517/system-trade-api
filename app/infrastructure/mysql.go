package infrastructure

import (
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func Connect() (db *gorm.DB, err error) {

	err = godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Memo:"DB_HOST"はdockerの場合データベースコンテナ名
	db, err = gorm.Open("mysql",
		os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+
			"@tcp("+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+")/"+
			os.Getenv("DB_DATABASE")+
			"?charset=utf8mb4&parseTime=True&loc=Local")

	if err != nil {
		log.Fatal(err)
	}

	return db, err
}
