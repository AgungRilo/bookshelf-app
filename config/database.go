package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// cleanDatabaseURL membuang parameter channel_binding (bawaan connection string Neon)
// karena tidak didukung lib/pq, koneksi tetap aman lewat sslmode=require
func cleanDatabaseURL(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	q := u.Query()
	q.Del("channel_binding")
	u.RawQuery = q.Encode()
	return u.String()
}

func ConnectDatabase() {
	// DATABASE_URL (Neon / Railway) diutamakan, selain itu pakai DB_*
	dsn := cleanDatabaseURL(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' dbname='%s' sslmode='%s'",
			GetEnv("DB_HOST", "localhost"),
			GetEnv("DB_PORT", "5432"),
			GetEnv("DB_USER", "postgres"),
			GetEnv("DB_PASSWORD", ""),
			GetEnv("DB_NAME", "bookshelf_db"),
			GetEnv("DB_SSLMODE", "disable"),
		)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("gagal membuka koneksi database: ", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatal("gagal terhubung ke database: ", err)
	}

	DB = db
	fmt.Println("berhasil terhubung ke database")
}
