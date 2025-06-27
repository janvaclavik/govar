package main

import (
	"os"

	"github.com/janvaclavik/govar"
)

func main() {
	data := map[string]any{
		"language": "Go",
		"year":     2009,
		"stable":   true,
	}

	file, err := os.Create("output.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString(govar.SdumpHTML(data))
}
