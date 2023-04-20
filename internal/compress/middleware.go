package compress

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
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

	if r.Header.Get(`Content-Encoding`) == `gzip` {
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
		w.WriteHeader(http.StatusBadRequest)
		return []byte("Incorrect URL"), err
	}
	return longURLByte, nil
}

func newCookie(r *http.Request, name string) (*http.Cookie, error) {
	ClientIP := r.RemoteAddr
	ClientIP = ClientIP[len(ClientIP)-5:]

	hashClientIP, err := hashCookie(ClientIP, name)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	newCookie := &http.Cookie{
		Name:     name,
		Value:    ClientIP,
		Path:     "/",
		HttpOnly: true,
	}
	r.AddCookie(newCookie)

	cookie := &http.Cookie{
		Name:  name,
		Value: hashClientIP,
		Path:  "/",
		//MaxAge:   -1,
		HttpOnly: true,
	}
	return cookie, nil
}

func CookieHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate;")

		name := "clientCookie"
		cookie, err := r.Cookie(name)

		switch {
		case err != nil:
			// создание куки
			cookie, err = newCookie(r, name)
			if err != nil {
				io.WriteString(w, err.Error())
				return
			}
			http.SetCookie(w, cookie)
		default:
			// проверка подлинности
			if IP, err := unhashCookie(cookie.Value, name, r); err == nil {
				//http.SetCookie(w, cookie)
				cookie.Value = IP
				r.AddCookie(cookie)
			} else {
				cookie, err = newCookie(r, name)
				if err != nil {
					io.WriteString(w, err.Error())
					return
				}
				http.SetCookie(w, cookie)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func unhashCookie(value, name string, r *http.Request) (string, error) {
	key := sha256.Sum256([]byte(name))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		log.Print(err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Print(err)
		return "", err
	}

	nonce := key[len(key)-2*aesgcm.NonceSize() : len(key)-aesgcm.NonceSize()]

	encrypted, err := hex.DecodeString(value)
	if err != nil {
		log.Print(err)
		return "", err
	}

	src, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	////ClientIP := r.RemoteAddr
	////ClientIP = ClientIP[len(ClientIP)-5:]
	//
	//if !hmac.Equal(src, []byte(ClientIP)) {
	//	log.Print(errors.New("invalid cookie value "), string(src), " ", ClientIP)
	//	return "", errors.New("invalid cookie value")
	//}
	return string(src), nil
}

func hashCookie(ip, name string) (string, error) {
	key := sha256.Sum256([]byte(name))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		log.Print(err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Print(err)
		return "", err
	}

	nonce := key[len(key)-2*aesgcm.NonceSize() : len(key)-aesgcm.NonceSize()]

	dst := hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(ip), nil))
	return dst, nil
}
