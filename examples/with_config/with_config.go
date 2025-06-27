package main

import "github.com/janvaclavik/govar"

func main() {
	type Address struct {
		City  string
		State string
	}

	person := struct {
		Name    string
		Age     int
		Address Address
	}{
		Name: "Alan Turing",
		Age:  41,
		Address: Address{
			City:  "London",
			State: "UK",
		},
	}

	defaultConfig := govar.DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            150,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
	}
	d := govar.NewDumper(defaultConfig)
	d.Dump(person)
}
