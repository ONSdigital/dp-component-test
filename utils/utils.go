package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// randomDBLength is the length of the string returned by RandomDatabase()
const randomDBLength = 15

// randomDBChars is the set of characters used by RandomDatabase()
const randomDBChars = "abcdefghijklmnopqrstuvwxyz"

// RandomDatabase returns a random valid mongo database name. You can use to
// to pick a new database name for each test to isolate tests from each other
// without having to tear down the whole server.
//
// This function will panic if it cannot generate a random number.
func RandomDatabase() string {
	b := make([]byte, randomDBLength)
	for i := range b {
		bigN, err := rand.Int(rand.Reader, big.NewInt(int64(len(randomDBChars))))
		if err != nil {
			panic(fmt.Errorf("error getting a random int: %s", err))
		}
		b[i] = randomDBChars[int(bigN.Int64())]
	}
	return string(b)
}
