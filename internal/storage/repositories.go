package storage

var (
	CacheNewID   map[string]string // cache map[newID]longURL
	CacheLongURL map[string]string // cache map[longURL]newID
	CacheDomen   map[string]int    // cache map[domen]id
)

func init() {
	CacheNewID = make(map[string]string)
	CacheLongURL = make(map[string]string)
	CacheDomen = make(map[string]int)
}
