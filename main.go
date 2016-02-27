//go:generate go run gen.go

package xc

import (
	"net/url"
	"strconv"
)

var (
	Token    string = ""
	Secret   string = ""
	Endpoint string = "https://api.example.com/"
)

func intToString(i int) string {
	var dst []byte
	strconv.AppendInt(dst, int64(i), 10)
	return string(dst)
}

func setAuth(v *url.Values) {
	v.Set("token", Token)
	v.Set("secret", Secret)
}

func Auth(token, secret string) {
	Token = token
	Secret = secret
}
