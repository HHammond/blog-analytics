package main

import (
	"database/sql"
	"encoding/base64"
	// "encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"log"
	"net/http"
	"os"
)

const pixelRaw = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
const defaultPort = ":8080"
const maxQueueSize = 2048

func makeHandlePixel(db *reform.DB) http.HandlerFunc {
	pixel, err := base64.StdEncoding.DecodeString(pixelRaw)
	if err != nil {
		panic(err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		go LogPageEvent(req, db)
		out.Header().Set("Content-Type", "image/gif")
		out.WriteHeader(http.StatusOK)
		out.Write(pixel)
	}
}

func LogPageEvent(req *http.Request, db *reform.DB) {
	event := PageEventFromRequest(req)
	fmt.Println(event)
	if err := db.Save(event); err != nil {
		panic(err)
	}
}

func initDB() *reform.DB {
	path := os.Getenv("DB_FILE")
	if path == "" {
		path = "out.db"
	}

	conn, err := sql.Open("sqlite3", path)
	logger := log.New(os.Stderr, "SQL: ", log.Flags())
	db := reform.NewDB(conn, sqlite3.Dialect, reform.NewPrintfLogger(logger.Printf))

	if err != nil {
		panic(err)
	}

	db.Exec(CreatePageEventTable)

	return db
}

func getPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		port = ":" + port
	} else {
		port = defaultPort
	}
	return port
}

func main() {
	db := initDB()
	port := getPort()

	http.HandleFunc("/a.gif", makeHandlePixel(db))

	if http.ListenAndServe(port, nil) != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server")
	}
}
