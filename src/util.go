package main

import (
	"math/rand"
)

type Map map[string]interface{}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789#$%&!"

func RandomID() (int) {
	return rand.Intn(999999999999999 - 1000000000000) + 100000000000
}

func RandomToken(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }

    return string(b)
}