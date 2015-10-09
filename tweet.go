package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/PuerkitoBio/goquery"
	"github.com/jbrukh/bayesian"
)

type Metadata struct {
	Created     int64
	Title       string
	Description string
	Type        string
	Image       string
	Url         string
	Site        string
}

type Object struct {
	Id       string
	Text     string
	Created  int64 `json:",string"`
	Stream   string
	Metadata *Metadata
}

var (
	files = flag.String("bayes_files", "", "comma separated list of bayes learning files")
)

func init() {
	flag.Parse()
	parts := strings.Split(*files, ",")
	if len(parts) == 0 || len(parts[0]) == 0 {
		panic("require at least one bayes file")
	}
}

func getMetadata(uri string) *Metadata {
	u, err := url.Parse(uri)
	if err != nil {
		return nil
	}

	d, err := goquery.NewDocument(u.String())
	if err != nil {
		return nil
	}

	g := &Metadata{
		Created: time.Now().UnixNano(),
	}

	for _, node := range d.Find("meta").Nodes {
		if len(node.Attr) < 2 {
			continue
		}

		p := strings.Split(node.Attr[0].Val, ":")
		if len(p) < 2 || (p[0] != "twitter" && p[0] != "og") {
			continue
		}

		switch p[1] {
		case "site_name":
			g.Site = node.Attr[1].Val
		case "site":
			if len(g.Site) == 0 {
				g.Site = node.Attr[1].Val
			}
		case "title":
			g.Title = node.Attr[1].Val
		case "description":
			g.Description = node.Attr[1].Val
		case "card", "type":
			g.Type = node.Attr[1].Val
		case "url":
			g.Url = node.Attr[1].Val
		case "image":
			if len(p) > 2 && p[2] == "src" {
				g.Image = node.Attr[1].Val
			} else if len(g.Image) == 0 {
				g.Image = node.Attr[1].Val
			}
		}
	}

	if len(g.Type) == 0 || len(g.Image) == 0 || len(g.Title) == 0 || len(g.Url) == 0 {
		return nil
	}

	return g
}

func main() {
	anaconda.SetConsumerKey("")
	anaconda.SetConsumerSecret("")
	api := anaconda.NewTwitterApi("", "")

	t := time.NewTicker(time.Second * 10)
	l := fmt.Sprintf("%d", time.Now().UnixNano())

	te := bayesian.Class("tech")
	ot := bayesian.Class("other")

	fparts := strings.Split(*files, ",")
	b, err := ioutil.ReadFile(fparts[0])
	if err != nil {
		panic(err.Error())
	}
	words := strings.Split(string(b), "\n")
	c := bayesian.NewClassifier(te, ot)
	c.Learn(words, te)

	posts := make(chan string, 10)
	post := time.NewTicker(time.Minute * 30)
	var cur []string
	var max float64
	var total float64
	var idx int

	go func() {
		for {
			select {
			case <-post.C:
				idx = -1
				max = -1000.0

				for i, p := range cur {
					total = 0.0
					parts := strings.Split(strings.ToLower(p), " ")
					scores, class, _ := c.LogScores(parts)

					if class != 0 {
						continue
					}

					for _, score := range scores {
						total += score
					}

					if total > max {
						max = total
						idx = i
					}
				}

				if idx >= 0 {
					api.PostTweet(cur[idx], url.Values{})
				}

				cur = []string{}
			case p := <-posts:
				cur = append(cur, p)
			}
		}
	}()

	for _ = range t.C {
		r, err := http.Get("http://127.0.0.1:8889/objects?stream=tech&last=" + l)
		if err != nil {
			fmt.Println(err)
			continue
		}

		b, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}
		var objects []Object
		err = json.Unmarshal(b, &objects)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, object := range objects {
			text := object.Text
			parts := strings.Split(object.Text, " ")
			for _, part := range parts {
				m := getMetadata(part)
				if m != nil && len(m.Title) > 0 && len(m.Url) > 0 {
					text = m.Title + " " + m.Url
				}
			}
			posts <- text
		}

		if len(objects) > 0 {
			l = fmt.Sprintf("%d", objects[len(objects)-1].Created)
		}
	}
}
