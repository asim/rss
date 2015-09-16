package main

import (
	"io/ioutil"
	"net/http"
)

func rssHandler(w http.ResponseWriter, r *http.Request) {
	rsp, err := http.Get("http://malten.me/thoughts?" + r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", rsp.Header.Get("Content-Type"))
	w.Write(b)
	return
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("html")))
	http.HandleFunc("/rss", rssHandler)
	http.ListenAndServe(":8888", nil)
}
