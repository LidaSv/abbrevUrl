package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
)

func newCookie(r *http.Request, name string) (*http.Cookie, error) {
	ClientIP := r.RemoteAddr
	ClientIP = ClientIP[len(ClientIP)-5:]
	//ClientIP := "87654"

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

		defer func() {
			if err := recover(); err != nil {
				log.Println("panic occurred:", err)
			}
		}()

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
			//log.Println("создание куки")
			http.SetCookie(w, cookie)
		default:
			// проверка подлинности
			if IP, err := UnhashCookie(cookie.Value, name); err == nil {
				//log.Println("проверка подлинности")
				cookie.Value = IP
				//log.Println(IP)
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

func UnhashCookie(value, name string) (string, error) {
	key := sha256.Sum256([]byte(name))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		log.Print("NewCipher: ", err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Print("NewGCM: ", err)
		return "", err
	}

	nonce := key[len(key)-2*aesgcm.NonceSize() : len(key)-aesgcm.NonceSize()]

	encrypted, err := hex.DecodeString(value)
	if err != nil {
		log.Print("DecodeString: :", err)
		return "", err
	}

	src, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		log.Println("Open: ", err)
		return "", err
	}

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
