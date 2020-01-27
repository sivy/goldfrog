/*
https://www.alexedwards.net/blog/simple-flash-messages-in-golang
*/
package blog

import (
	"encoding/base64"
	"net/http"
	"time"
)

func SetFlash(w http.ResponseWriter, name string, value string) {
	byteValue := []byte(value)
	c := &http.Cookie{Name: name, Value: encode(byteValue)}
	http.SetCookie(w, c)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return "", nil
		default:
			return "", err
		}
	}
	byteValue, err := decode(c.Value)
	if err != nil {
		return "", err
	}
	value := string(byteValue)
	dc := &http.Cookie{Name: name, MaxAge: -1, Expires: time.Unix(1, 0)}
	logger.Debugf("setting dc on writer %v", w)
	http.SetCookie(w, dc)
	return value, nil
}

// -------------------------

func encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}
