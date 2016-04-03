package server

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

// RandomString is far from perfect random string generator,
// it is based on underlying len number of bytes which are then base64 encoded
func randomStringFromBytes(len int) string {
	randData := make([]byte, len)
	_, err := rand.Read(randData)
	if err != nil {
		log.Fatalf("Could not generate random secret: %v", err)
	}
	return base64.StdEncoding.EncodeToString(randData)
}
