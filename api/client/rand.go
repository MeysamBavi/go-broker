package main

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

const defaultCharset = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandomString(length int) string {
	return RandomStringWithCharset(length, defaultCharset)
}

func RandomStringWithCharset(length int, charset string) string {
	str := make([]byte, length)
	for i := 0; i < length; i++ {
		str[i] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

func RandomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func RandomItem[T any](src []T) T {
	return src[rand.Intn(len(src))]
}
