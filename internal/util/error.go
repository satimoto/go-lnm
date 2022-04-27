package util

import (
	"log"
)

func LogOnError(code string, message string, err error) {
	if err != nil {
		log.Printf("%s: %s", code, message)
		log.Printf("%s: %v", code, err)
	}
}

func PanicOnError(code string, message string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", code, message)
		log.Fatalf("%s: %v", code, err)
	}
}
