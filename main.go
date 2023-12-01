package main

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func main() {

	c := getFlags()
	Debug(c, c)
	client, err := buildKubeClient(c)
	if err != nil {
		panic(err.Error())
	}

	cl := NewCluster(c, client)
	err = cl.NewQueue("test")
	if err != nil {
		panic(err.Error())
	}

	q, err := cl.GetQueue("test")
	if err != nil {
		panic(err.Error())
	}

	for i := 1; i <= 100; i++ {
		_, err = q.Push([]byte(rs(20)))
		if err != nil {
			panic(err.Error())
		}
	}
	fmt.Println(q.Length)
}

func rs(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
