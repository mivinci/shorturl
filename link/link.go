package link

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mivinci/shortid"
	"github.com/mivinci/ttl"
)

var (
	db     *bolt.DB
	cache  *ttl.Cache
	mtx    sync.RWMutex
	bkLink = []byte("link")

	ErrDup      = errors.New("该链接已存在")
	ErrNotFound = errors.New("该链接不存在")
)

type LinkIP struct {
	Note  string  `json:"note"`
	IP    string  `json:"ip"`
	Links []*Link `json:"links"`
}

type Link struct {
	ID      uint64        `json:"id"`
	Origin  string        `json:"origin"`
	Alias   string        `json:"alias"`
	IP      string        `json:"ip"`
	Ctime   time.Time     `json:"ctime"`
	Mtime   time.Time     `json:"mtime"`
	Expired bool          `json:"expired"`
	TTL     time.Duration `json:"ttl"`
}

func Init(dsn string) {
	var err error
	db, err = bolt.Open(dsn, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		log.Fatalf("open boltdb: %s", err)
	}
	err = db.Update(func(t *bolt.Tx) error {
		_, err := t.CreateBucketIfNotExists(bkLink)
		return err
	})
	if err != nil {
		log.Fatalf("create boltdb bucket: %s", err)
	}
	cache = ttl.New()
	cache.Evict = func(k, v interface{}) {
		l, ok := v.(*Link)
		if !ok {
			log.Printf("convert link: %s failed since %s, skipped\n", k, err)
			return
		}
		if err := l.Kill(db); err != nil {
			log.Printf("kill link: %s failed since %s, skipped\n", k, err)
			return
		}
		log.Printf("kill link: %s -> %s\n", l.Origin, l.Alias)
	}
	if err := loadCache(); err != nil {
		log.Fatalf("load cache failed: %s", err)
	}
}

func loadCache() error {
	return db.View(func(t *bolt.Tx) error {
		b := t.Bucket(bkLink)
		return b.ForEach(func(k, v []byte) error {
			link := new(Link)
			if err := json.Unmarshal(v, link); err != nil {
				log.Printf("list link unmarshal: %s since %s, skipped\n", k, err)
				return err
			}
			if link.Alive() {
				if err := cache.Add(link.Alias, link, link.TTL); err != nil {
					log.Printf("load cache %s failed since %s, skipped\n", link.Alias, err)
					return err
				}
				log.Printf("load cache: %v\n", link)
			}
			return nil
		})
	})
}

func (l Link) Alive() bool {
	return l.Mtime.Add(l.TTL).After(time.Now())
}

func (l *Link) Kill(db *bolt.DB) error {
	return db.Update(func(t *bolt.Tx) error {
		b := t.Bucket(bkLink)
		l.Expired = true
		buf, _ := json.Marshal(l)
		return b.Put([]byte(l.Origin), buf)
	})
}

func (l *Link) Cure(b *bolt.Bucket, cache *ttl.Cache, newTTL time.Duration) error {
	l.Expired = false
	l.Mtime = time.Now()
	l.TTL = newTTL
	buf, _ := json.Marshal(l)
	mtx.Lock()
	defer mtx.Unlock()
	if err := cache.Add(l.Alias, l, newTTL); err != nil {
		return err
	}
	return b.Put([]byte(l.Origin), buf)
}

func AddLink(origin, ip string, ttl time.Duration) (*Link, error) {
	link := new(Link)
	err := db.Update(func(t *bolt.Tx) error {
		b := t.Bucket(bkLink)

		if buf := b.Get([]byte(origin)); buf != nil {
			if err := json.Unmarshal(buf, link); err != nil {
				return err
			}
			if link.Expired {
				log.Printf("cure link: %s -> %s\n", origin, link.Alias)
				return link.Cure(b, cache, ttl)
			}
			return nil
		}

		link.ID, _ = b.NextSequence()
		link.Alias = shortid.String(int(link.ID))
		link.Origin = origin
		link.IP = ip
		link.TTL = ttl
		link.Ctime = time.Now()
		link.Mtime = link.Ctime
		// codec
		buf, err := json.Marshal(link)
		if err != nil {
			return err
		}
		// cache
		mtx.Lock()
		if err = cache.Add(link.Alias, link, link.TTL); err != nil {
			log.Printf("add link: %s cache failed since %s, skipped\n", link.Origin, err)
		}
		mtx.Unlock()
		log.Printf("new link: %v\n", link)
		// database
		return b.Put([]byte(link.Origin), buf)
	})
	return link, err
}

func GetLink(alias string) (*Link, error) {
	mtx.RLock()
	defer mtx.RUnlock()
	v, err := cache.Get(alias)
	if err != nil {
		log.Printf("get link: %s (cache miss)\n", alias)
		return nil, err
	}
	log.Printf("get link: %s (cache)\n", alias)
	return v.(*Link), nil
}

func ListLinkByIP(ip string) ([]*Link, error) {
	links := make([]*Link, 0)
	err := db.View(func(t *bolt.Tx) error {
		b := t.Bucket(bkLink)
		return b.ForEach(func(k, v []byte) error {
			link := new(Link)
			if err := json.Unmarshal(v, link); err != nil {
				log.Printf("list link unmarshal: %s since %s, skipped\n", k, err)
				return err
			}
			if link.IP == ip {
				links = append(links, link)
			}
			return nil
		})
	})
	return links, err
}
