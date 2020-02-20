package node

import (
	crypto_rand "crypto/rand"
	"encoding/base32"
	"hash/fnv"
	"strings"
)

func randomBytesString(length int) string {
	randomBytes := make([]byte, 32)
	_, err := crypto_rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}

func hashFnv64(s []string) uint64 {
	d := strings.Join(s, "")
	h := fnv.New64a()
	h.Write([]byte(d))
	return h.Sum64()
}
