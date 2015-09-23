package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/PuerkitoBio/goquery"
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
			api.PostTweet(text, url.Values{})
		}

		if len(objects) > 0 {
			l = fmt.Sprintf("%d", objects[len(objects)-1].Created)
		}
	}
}
