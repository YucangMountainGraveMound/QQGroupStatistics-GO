package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"
)

// MD5 string->md5
func MD5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

// RandomString generate a random string
func RandomString(lengrh int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < lengrh; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
