package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() (*gorm.DB, error) {
	usePostgres := strings.ToLower(os.Getenv("USE_POSTGRES")) == "true"

	var db *gorm.DB
	var err error

	config := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	if usePostgres {
		dsn := os.Getenv("POSTGRES_DSN") // example: "host=localhost user=postgres password=mysecret dbname=hammondDB port=5432 sslmode=disable"
		if dsn == "" {
			return nil, errors.New("no Postgres DSN set")
		}
		log.Println("Using Postgres")
		db, err = gorm.Open(postgres.Open(dsn), config)
	} else {
		configPath := os.Getenv("CONFIG")
		dbPath := path.Join(configPath, "hammond.db")
		log.Println("Using SQLite at:", dbPath)
		db, err = gorm.Open(sqlite.Open(dbPath), config)
	}

	if err != nil {
		fmt.Println("db err:", err)
		return nil, err
	}

	localDB, _ := db.DB()
	localDB.SetMaxIdleConns(10)

	//db.LogMode(true)
	DB = db
	return DB, nil
}

// Migrate Database
func Migrate() {
	err := DB.AutoMigrate(&Attachment{}, &QuickEntry{}, &User{}, &Vehicle{}, &UserVehicle{}, &VehicleAttachment{}, &Fillup{}, &Expense{}, &Setting{}, &JobLock{}, &Migration{})
	if err != nil {
		fmt.Println("1 " + err.Error())
	}
	err = DB.SetupJoinTable(&User{}, "Vehicles", &UserVehicle{})
	if err != nil {
		fmt.Println(err.Error())
	}
	err = DB.SetupJoinTable(&Vehicle{}, "Attachments", &VehicleAttachment{})
	if err != nil {
		fmt.Println(err.Error())
	}
	RunMigrations()
}

// Using this function to get a connection, you can create your connection pool here.
func GetDB() *gorm.DB {
	return DB
}
