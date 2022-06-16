package utils

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const letterBytesStringOnly = "abcdefghijklmnopqrstuvwxyz"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

//var src = rand.NewSource(time.Now().UnixNano())

type lockedSource struct {
	source rand.Source
	lock   sync.Mutex
}

func newLockSource(seed int64) *lockedSource {
	return &lockedSource{
		source: rand.NewSource(seed),
	}
}

func (ls *lockedSource) Int63() int64 {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	return ls.source.Int63()
}

func (ls *lockedSource) Seed(seed int64) {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	ls.source.Seed(seed)
}

var src = newLockSource(time.Now().UnixNano())

// randStringBytesMaskImprSrcUnsafe generates a random string of a given length.
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go/31832326#31832326
func randStringBytesMaskImprSrcUnsafe(n int, strSet string) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(strSet) {
			b[i] = strSet[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func RandString(n int) string {
	return randStringBytesMaskImprSrcUnsafe(n, letterBytes)
}

func RandStringLowerCase(n int) string {
	return randStringBytesMaskImprSrcUnsafe(n, letterBytesStringOnly)
}
