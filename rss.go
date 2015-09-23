package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SlyMarbo/rss"
)

var (
	urls = map[string]*rss.Feed{
		"http://www.forbes.com/technology/feed/":                                 nil,
		"http://fortune.com/feed/":                                               nil,
		"http://a16z.com/feed/":                                                  nil,
		"http://feeds.arstechnica.com/arstechnica/index":                         nil,
		"http://feeds.feedburner.com/fastcompany/headlines":                      nil,
		"http://feeds.mashable.com/Mashable":                                     nil,
		"http://qz.com/feed":                                                     nil,
		"http://recode.net/feed/":                                                nil,
		"http://feedproxy.google.com/typepad/alleyinsider/silicon_alley_insider": nil,
		"http://feeds.feedburner.com/TechCrunch/":                                nil,
		"http://thenewstack.io/blog/feed/":                                       nil,
		"http://thenextweb.com/feed/":                                            nil,
		"http://valleywag.gawker.com/rss":                                        nil,
		"http://feeds.venturebeat.com/VentureBeat":                               nil,
		"http://www.wired.com/category/business/feed/":                           nil,
		"http://www.wired.com/category/gear/feed/":                               nil,
		"http://www.wsj.com/xml/rss/3_7455.xml":                                  nil,
		"http://www.bloomberg.com/feed/bview/":                                   nil,
	}
)

var (
	backfill = flag.Bool("backfill", false, "")
)

func init() {
	flag.Parse()
}

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
		fmt.Println("error updating", feed.UpdateURL, err)
		return
	}

	for _, item := range feed.Items {
		think("tech", item.Title+"  "+item.Link)
	}
}

func fetchAll() {
	for url, feed := range urls {
		if feed == nil {
			fd, err := rss.Fetch(url)
			if err != nil {
				fmt.Println("error fetching", url, err)
				continue
			}
			urls[url] = fd
			feed = fd

			if *backfill {
				for _, item := range fd.Items {
					think("tech", item.Title+"  "+item.Link)
				}
			}
		}
		fetch(feed)
		feed.Refresh = time.Now()
	}
}

func main() {
	tick := time.NewTicker(time.Minute)
	for _ = range tick.C {
		fetchAll()
	}
}
