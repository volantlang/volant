package error

import (
	"fmt"
	"log"
)

type langError string

func New(message string, line int, column int) {
	log.Fatal(fmt.Sprintf("Error: line %d column %d: %s\n", line, column, message))
}

// for general (non-code) errors
func NewGenError(message string) {
	log.Fatal(message)
}
