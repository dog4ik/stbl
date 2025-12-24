package utils

import (
	"log"
	"os"
)

// Fetch env value, panic if the key is not present
func ExpectEnv(key string) string {
	value, present := os.LookupEnv(key)

	if !present {
		log.Fatalf("%s env variable is not present", key)
	}
	log.Printf("%s: %s", key, value)
	return value
}
