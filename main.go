package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

func dataHandler(w http.ResponseWriter, r *http.Request) {
	rsp, err := http.Get("http://127.0.0.1:8889/objects?" + r.URL.RawQuery)
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

func getData(stream string) ([]byte, error) {
	rsp, err := http.Get("http://127.0.0.1:8889/objects?stream=" + stream)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	b, err := getData(vars["stream"])
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var m []map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	feed := &feeds.Feed{
		Title:       "asl.am rss feed",
		Link:        &feeds.Link{Href: "http://asl.am/rss.atom"},
		Description: "Startup and tech related news",
		Author:      &feeds.Author{"asl.am", ""},
		Created:     time.Now(),
		Copyright:   "Copyright © asl.am",
	}

	for i := len(m) - 1; i >= 0; i-- {
		if meta := m[i]["Metadata"]; meta != nil {
			mr := meta.(map[string]interface{})
			feed.Items = append(feed.Items, &feeds.Item{
				Title:   mr["Title"].(string),
				Link:    &feeds.Link{Href: mr["Url"].(string)},
				Created: time.Unix(int64(mr["Created"].(float64))/1e9, 0),
			})
		}
	}

	if strings.HasSuffix(r.URL.Path, ".atom") {
		w.Header().Set("Content-Type", "application/atom+xml")
		feed.WriteAtom(w)
	} else {
		w.Header().Set("Content-Type", "application/rss+xml")
		feed.WriteRss(w)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/rss", dataHandler)
	r.HandleFunc("/rss/{stream:[a-z]+}.xml", rssHandler)
	r.HandleFunc("/rss/{stream:[a-z]+}.atom", rssHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("html")))
	http.ListenAndServe(":8888", r)
}
