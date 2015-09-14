package main

import (
	"net/http"
)

func rssHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("html")))
	http.HandleFunc("/rss", rssHandler)
	http.ListenAndServe(":8888", nil)
}
