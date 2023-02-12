package app

import (
	"io"
	"net/http"
	"strconv"
)

var cache map[string]string

func init() {
	cache = make(map[string]string)
}

func check(longUrl string) int {
	resp, err := http.Get(longUrl)
	if err != nil {
		return 404
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 400
	}
	return 200
}

func MyShortUrl(longUrl string) string {
	newId := "/new" + strconv.Itoa(len(cache)+1) + "/"
	shortUrl := "http://localhost:8080" + newId
	cache[newId] = longUrl
	return shortUrl

}

func Abbrevurl(w http.ResponseWriter, r *http.Request) {
	cache["/new1/"] = "https://vk.com"
	switch r.Method {
	case http.MethodPost:
		longUrlByte, _ := io.ReadAll(r.Body)
		longUrl := string(longUrlByte)

		status := check(longUrl)
		if status != 200 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect URL"))
			return
		}
		shortUrl := MyShortUrl(longUrl)
		//w.Header().Add("Location", longUrl)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortUrl))
		//fmt.Fprintln(w, 201, cache[longUrl].shortUrl)
	case http.MethodGet:

		newId := r.URL.Path
		//shortUrl := string(w) + newId
		if longUrl, ok := cache[newId]; ok {
			w.Header().Set("Location", longUrl)
			//w.Header().Add("Location", longUrl)
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte(longUrl))
			return
		}
		//longUrl := cache[newId]
		//w.Header().Set("Pir", longUrl)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Short URL not in memory"))
	}
}
