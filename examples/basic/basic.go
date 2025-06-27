package main

import (
	"github.com/janvaclavik/govar"
)

func main() {
	// Sample data
	person := struct {
		Name    string
		Age     int
		Hobbies []string
	}{
		Name:    "Ada Lovelace",
		Age:     36,
		Hobbies: []string{"math", "poetry", "programming"},
	}

	// Basic dump with default settings
	govar.Dump(person)
}
