package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
	"time"
)

const pixelRaw = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
const defaultPort = ":8080"
const maxQueueSize = 2048

func makeHandlePixel(eventQueue chan *PageEvent) http.HandlerFunc {
	pixel, err := base64.StdEncoding.DecodeString(pixelRaw)
	if err != nil {
		panic(err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		eventQueue <- PageEventFromRequest(req)
		out.Header().Set("Content-Type", "image/gif")
		out.WriteHeader(http.StatusOK)
		out.Write(pixel)
	}
}

func runEventWriter(db *sql.DB, eventQueue chan *PageEvent) {
	queue := make([]*PageEvent, 0, 100000)
	for {
		select {
		case ev := <-eventQueue:
			queue = append(queue, ev)

		case <-time.After(time.Millisecond * 5):
			if len(queue) > 0 {
				db.Exec("BEGIN TRANSACTION;")
				for _, ev := range queue {
					ev.InsertIntoDB(db)
					fmt.Println(ev)
				}
				db.Exec("END TRANSACTION;")
				queue = queue[:0]
			}
		}
	}
}

func initDB() *sql.DB {
	path := os.Getenv("DB_FILE")
	if path == "" {
		path = "out.db"
	}

	db, err := sql.Open("sqlite3", path)

	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("DB is null")
	}

	_, err = db.Exec(SQLPageEventCreateTable)
	if err != nil {
		panic(err)
	}

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
	port := getPort()
	db := initDB()
	defer db.Close()
	queue := make(chan *PageEvent, maxQueueSize)

	go runEventWriter(db, queue)

	http.HandleFunc("/a.gif", makeHandlePixel(queue))

	if http.ListenAndServe(port, nil) != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server")
	}
}
