package config

import (
	"io/ioutil"
	"log"
)

var (
	defaultHostname = "localhost"
	defaultPort     = "10666"
)

func mustWrite(path string, contents []byte) {
	if err := ioutil.WriteFile(path, contents, 0644); err != nil {
		log.Fatal(err)
	}
}

