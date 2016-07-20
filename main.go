package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
)

const pixelRaw = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
const defaultPort = ":8080"
const maxQueueSize = 2048

func makeHandlePixel(writeQueue chan []byte) http.HandlerFunc {
	pixel, err := base64.StdEncoding.DecodeString(pixelRaw)
	if err != nil {
		panic(err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		go LogPageEvent(req, writeQueue)
		out.Header().Set("Content-Type", "image/gif")
		out.WriteHeader(http.StatusOK)
		out.Write(pixel)
	}
}

func main() {
	path := os.Getenv("OUTFILE")

	var outfile *os.File
	if path == "" {
		outfile = os.Stdin
	} else {
		outfile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
	}

	port := os.Getenv("PORT")
	if port != "" {
		port = ":" + port
	} else {
		port = defaultPort
	}

	writeQueue := make(chan []byte, maxQueueSize)
	go WriteEventsToFile(outfile, writeQueue)

	http.HandleFunc("/a.gif", makeHandlePixel(writeQueue))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server")
	}
}
