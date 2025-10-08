package id

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// New returns a random hex identifier fallback to timestamp derived string.
func New() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(buf)
}
