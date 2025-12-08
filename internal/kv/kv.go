package kv

import (
	"os"

	"github.com/ondrejsikax/simple-key-value-storage/pkg/client"
)

const (
	DEFAULT_ORIGIN   = "https://simple-key-value-storage.sikalabs.io"
	DEFAULT_PASSWORD = "sikalabs"
)

func Set(key, value string) error {
	origin, password := getOriginAndPassword()
	return client.Set(origin, password, key, value)
}

func Get(key string) (string, error) {
	origin, password := getOriginAndPassword()
	return client.Get(origin, password, key)
}

func getOriginAndPassword() (string, string) {
	origin := DEFAULT_ORIGIN
	password := DEFAULT_PASSWORD

	if envOrigin := os.Getenv("SKVS_ORIGIN"); envOrigin != "" {
		origin = envOrigin
	}

	if envPassword := os.Getenv("SKVS_PASSWORD"); envPassword != "" {
		password = envPassword
	}

	if envNoPassword := os.Getenv("SKVS_NO_PASSWORD"); envNoPassword != "" {
		password = ""
	}

	return origin, password
}
