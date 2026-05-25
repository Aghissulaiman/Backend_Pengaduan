package db

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() error {
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")

    if dbHost == "" {
        dbHost = "localhost"
    }
    if dbPort == "" {
        dbPort = "3306"
    }
    if dbUser == "" {
        dbUser = "root"
    }
    if dbName == "" {
        dbName = "pengaduan_provinsi"
    }

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        dbUser, dbPassword, dbHost, dbPort, dbName)

    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        return err
    }

    DB.SetMaxOpenConns(25)
    DB.SetMaxIdleConns(5)
    DB.SetConnMaxLifetime(5 * time.Minute)

    if err = DB.Ping(); err != nil {
        return err
    }

    log.Println("✅ Database connected")
    return nil
}

func CloseDB() {
    if DB != nil {
        DB.Close()
    }
}