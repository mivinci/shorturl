package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mivinci/shorturl/link"
	"github.com/mivinci/ttl"
)

var (
	dbpath = flag.String("db", "shorturl.db", "path to boltdb file")
	domain = flag.String("domain", "http://localhost:5000", "your short domain")
	port   = flag.Int("port", 5000, "port to serve on")
)

func main() {
	flag.Parse()
	link.Init(*dbpath)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			path := r.URL.Path[1:]
			if len(path) != 5 {
				http.ServeFile(w, r, fmt.Sprintf("html/%s", path))
				return
			}
			linkGet(w, r)
		case "POST":
			linkPost(w, r)
		}
	})
	http.HandleFunc("/history", listLinkIP)

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func linkGet(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Path[1:]
	l, err := link.GetLink(alias)
	if errors.Is(err, ttl.ErrExpire) || errors.Is(err, ttl.ErrNotFound) {
		http.ServeFile(w, r, "html/404.html")
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	http.Redirect(w, r, l.Origin, 302)
}

func linkPost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ttl, err := strconv.Atoi(q.Get("ttl"))
	if err != nil {
		http.Error(w, "非法生存时间", 401)
		return
	}
	ip := link.RemoteIP(r)
	l, err := link.AddLink(q.Get("origin"), ip, time.Duration(ttl)*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s/%s", *domain, l.Alias)
}

func listLinkIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	ip := link.RemoteIP(r)
	links, err := link.ListLinkByIP(ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	lip := link.LinkIP{
		Note:  "不想写了，自己凑合着看吧 🙂",
		IP:    ip,
		Links: links,
	}
	buf, err := json.Marshal(&lip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(buf) // nolint: errcheck
}
