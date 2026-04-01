package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Random() string {
	return fmt.Sprintf("0.%17v", rand.Int63n(100000000000000000))
}

func RandomRange(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func RandomUUID() string {
	uuid := RandomString(32)
	return uuid
}

func GenerateUUID() string {
	return RandomUUID()
}

func RandomWithPattern(pattern string) string {
	reg := regexp.MustCompile("[xy]")
	data := reg.ReplaceAllFunc([]byte(pattern), func(msg []byte) []byte {
		var i int64
		t := int64(16 * rand.Float32())
		if msg[0] == 'x' {
			i = t
		} else {
			i = 3&t | 8
		}
		return []byte(strconv.FormatInt(i, 16))
	})
	return string(data)
}
