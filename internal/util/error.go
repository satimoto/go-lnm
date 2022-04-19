package util

import (
	"log"
)

func LogOnError(message string, err error) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}

func PanicOnError(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
