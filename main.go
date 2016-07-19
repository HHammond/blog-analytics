package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
)

const pixelRaw = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="

func makeRequestPixel() http.HandlerFunc {
	pixel, err := base64.StdEncoding.DecodeString(pixelRaw)
	if err != nil {
		panic(err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		go LogPageEvent(req)
		out.Header().Set("Content-Type", "image/gif")
		out.WriteHeader(http.StatusOK)
		out.Write(pixel)
	}
}

func main() {
	http.HandleFunc("/a.gif", makeRequestPixel())

	fmt.Fprintf(os.Stderr, "Starting server.")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server")
	}
}
