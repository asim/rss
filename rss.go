package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/SlyMarbo/rss"
)

var (
	urls = map[string]*rss.Feed{
		"http://a16z.com/feed/":                                                  nil,
		"http://feeds.arstechnica.com/arstechnica/index":                         nil,
		"http://feeds.feedburner.com/fastcompany/headlines":                      nil,
		"http://feeds.mashable.com/Mashable":                                     nil,
		"http://qz.com/feed":                                                     nil,
		"http://recode.net/feed/":                                                nil,
		"http://feedproxy.google.com/typepad/alleyinsider/silicon_alley_insider": nil,
		"http://feeds.feedburner.com/TechCrunch/":                                nil,
		"http://thenewstack.io/rss-feeds/":                                       nil,
		"http://thenextweb.com/feed/":                                            nil,
		"http://valleywag.gawker.com/rss":                                        nil,
		"http://feeds.venturebeat.com/VentureBeat":                               nil,
		"http://www.wired.com/category/business/feed/":                           nil,
		"http://www.wired.com/category/gear/feed/":                               nil,
		"http://www.wsj.com/xml/rss/3_7455.xml":                                  nil,
		"http://www.bloomberg.com/feed/bview/":                                   nil,
	}
)

func think(stream, text string) {
	http.PostForm("http://127.0.0.1:8889/objects", url.Values{
		"text":   []string{text},
		"stream": []string{stream},
	})
}

func fetch(feed *rss.Feed) {
	feed.Unread = 0
	feed.Items = []*rss.Item{}
	if err := feed.Update(); err != nil {
		return
	}

	for _, item := range feed.Items {
		think("tech", feed.Title+": "+item.Title+"  "+item.Link)
	}
}

func fetchAll() {
	for url, feed := range urls {
		if feed == nil {
			fd, err := rss.Fetch(url)
			if err != nil {
				continue
			}
			urls[url] = fd
			feed = fd
		}
		fetch(feed)
	}
}

func main() {
	tick := time.NewTicker(time.Minute*10)
	for _ = range tick.C {
		fetchAll()
	}
}
