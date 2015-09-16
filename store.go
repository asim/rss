package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/groupcache/lru"
)

const (
	defaultStream  = "_"
	maxObjectSize = 512
	maxObjects    = 1000
	maxStreams     = 1000
	streamTTL      = 8.64e13
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

type Stream struct {
	Id       string
	Objects []*Object
	Updated  int64
}

type Object struct {
	Id      string
	Text    string
	Created int64 `json:",string"`
	Stream  string
	Metadata *Metadata
}

type Store struct {
	Created int64
	Updates chan *Object

	mtx      sync.RWMutex
	Streams  *lru.Cache
	streams  map[string]int64
	metadatas map[string]*Metadata
}

var (
	C = newStore()
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func newStore() *Store {
	return &Store{
		Created:  time.Now().UnixNano(),
		Streams:  lru.New(maxStreams),
		Updates:  make(chan *Object, 100),
		streams:  make(map[string]int64),
		metadatas: make(map[string]*Metadata),
	}
}

func newStream(id string) *Stream {
	return &Stream{
		Id:      id,
		Updated: time.Now().UnixNano(),
	}
}

func newObject(text, stream string) *Object {
	return &Object{
		Id:      uuid.New(),
		Text:    text,
		Created: time.Now().UnixNano(),
		Stream:  stream,
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

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	object := r.Form.Get("id")
	stream := r.Form.Get("stream")

	last, err := strconv.ParseInt(r.Form.Get("last"), 10, 64)
	if err != nil {
		last = 0
	}

	limit, err := strconv.ParseInt(r.Form.Get("limit"), 10, 64)
	if err != nil {
		limit = 25
	}

	direction, err := strconv.ParseInt(r.Form.Get("direction"), 10, 64)
	if err != nil {
		direction = 1
	}

	// default stream
	if len(stream) == 0 {
		stream = defaultStream
	}

	objects := C.Retrieve(object, stream, direction, last, limit)
	b, _ := json.Marshal(objects)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(b))
}

func getStreamsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	b, _ := json.Marshal(C.List())
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(b))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	object := r.Form.Get("text")
	stream := r.Form.Get("stream")

	if len(object) == 0 {
		http.Error(w, "Object cannot be blank", 400)
		return
	}

	// default stream
	if len(stream) == 0 {
		stream = defaultStream
	}

	// default length
	if len(object) > maxObjectSize {
		object = object[:maxObjectSize]
	}

	select {
	case C.Updates <- newObject(object, stream):
	case <-time.After(time.Second):
		http.Error(w, "Timed out creating object", 504)
	}
}

func (c *Store) Metadata(t *Object) {
	parts := strings.Split(t.Text, " ")
	for _, part := range parts {
		g := getMetadata(part)
		if g == nil {
			continue
		}
		c.mtx.Lock()
		c.metadatas[t.Id] = g
		c.mtx.Unlock()
		return
	}
}

func (c *Store) List() map[string]int64 {
	c.mtx.RLock()
	streams := c.streams
	c.mtx.RUnlock()
	return streams
}

func (c *Store) Save(object *Object) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	var stream *Stream

	if obj, ok := c.Streams.Get(object.Stream); ok {
		stream = obj.(*Stream)
	} else {
		stream = newStream(object.Stream)
		c.Streams.Add(object.Stream, stream)
	}

	stream.Objects = append(stream.Objects, object)
	if len(stream.Objects) > maxObjects {
		stream.Objects = stream.Objects[1:]
	}
	stream.Updated = time.Now().UnixNano()
}

func (c *Store) Retrieve(object string, streem string, direction, last, limit int64) []*Object {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	var stream *Stream

	if object, ok := c.Streams.Get(streem); ok {
		stream = object.(*Stream)
	} else {
		return []*Object{}
	}

	if len(object) == 0 {
		var objects []*Object

		if limit <= 0 {
			return objects
		}

		li := int(limit)

		// go back in time
		if direction < 0 {
			for i := len(stream.Objects) - 1; i >= 0; i-- {
				if len(objects) >= li {
					return objects
				}

				object := stream.Objects[i]

				if object.Created < last {
					if g, ok := c.metadatas[object.Id]; ok {
						tc := *object
						tc.Metadata = g
						objects = append(objects, &tc)
					} else {
						objects = append(objects, object)
					}
				}
			}
			return objects
		}

		start := 0
		if len(stream.Objects) > li {
			start = len(stream.Objects) - li
		}

		for i := start; i < len(stream.Objects); i++ {
			if len(objects) >= li {
				return objects
			}

			object := stream.Objects[i]

			if object.Created > last {
				if g, ok := c.metadatas[object.Id]; ok {
					tc := *object
					tc.Metadata = g
					objects = append(objects, &tc)
				} else {
					objects = append(objects, object)
				}
			}
		}
		return objects
	}

	// retrieve one
	for _, t := range stream.Objects {
		var objects []*Object
		if object == t.Id {
			if g, ok := c.metadatas[t.Id]; ok {
				tc := *t
				tc.Metadata = g
				objects = append(objects, &tc)
			} else {
				objects = append(objects, t)
			}
			return objects
		}
	}

	return []*Object{}
}

func (c *Store) Run() {
	t1 := time.NewTicker(time.Hour)
	t2 := time.NewTicker(time.Minute)
	streams := make(map[string]int64)

	for {
		select {
		case object := <-c.Updates:
			c.Save(object)
			streams[object.Stream] = time.Now().UnixNano()
			go c.Metadata(object)
		case <-t1.C:
			now := time.Now().UnixNano()
			for stream, u := range streams {
				if d := now - u; d > streamTTL {
					c.Streams.Remove(stream)
					delete(streams, stream)
				}
			}
			c.mtx.Lock()
			for metadata, g := range c.metadatas {
				if d := now - g.Created; d > streamTTL {
					delete(c.metadatas, metadata)
				}
			}
			c.mtx.Unlock()
		case <-t2.C:
			c.mtx.Lock()
			c.streams = streams
			c.mtx.Unlock()
		}
	}
}

func main() {
	go C.Run()

	http.HandleFunc("/streams", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getStreamsHandler(w, r)
		default:
			http.Error(w, "unsupported method "+r.Method, 400)
		}
	})

	http.HandleFunc("/objects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getHandler(w, r)
		case "POST":
			postHandler(w, r)
		default:
			http.Error(w, "unsupported method "+r.Method, 400)
		}
	})

	http.ListenAndServe("127.0.0.1:8889", nil)
}
