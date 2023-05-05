package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func ReadBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	var reader io.Reader

	//Правильно ли так (через strings.Contains()) искать gzip,
	//если Content-Encoding множественный?
	if strings.Contains(r.Header.Get(`Content-Encoding`), `gzip`) {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}

	longURLByte, err := io.ReadAll(reader)
	defer r.Body.Close()
	if err != nil || len(longURLByte) == 0 {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("incorrect URL")
	}
	return longURLByte, nil
}
