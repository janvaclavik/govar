package main

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/janvaclavik/govar"
)

func main() {

	// Http request
	req, _ := http.NewRequest("POST", "https://space.api/launch", strings.NewReader("ignite=true"))
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Add("User-Agent", "govar-bot/1.0")
	govar.Dump(req)

	// Writer
	f := os.Stdout
	govar.Dump(f)

	// LONG AND SUPER GOOFY
	// tmpl, _ := template.New("test").Parse("Hello {{.Name}}!")
	// govar.Dump(tmpl)
	// html := govar.SdumpHTML(tmpl)

	// Zip
	type ZipWrapper struct {
		Header        zip.FileHeader
		privateWriter io.Writer
		privateLogs   []byte
	}
	header := zip.FileHeader{
		Name:     "data.json",
		Comment:  "Have you tried govar yet? It's the best dumper out there!",
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	zipWrapper := ZipWrapper{privateLogs: []byte("Zip logs corrupted, correction required! Transmission over."), Header: header, privateWriter: nil}
	govar.Dump(zipWrapper)

	// Time stuff (can be long if not as stringer)
	now := time.Now()
	duration := 2*time.Hour + 34*time.Minute
	govar.Dump(now, duration)

	// Cyclic references
	type Person struct {
		Name   string
		Avatar []byte
		Loves  *Person
	}
	p1 := &Person{Name: "Alice", Avatar: []byte{12, 43, 53}}
	p2 := &Person{Name: "Bob", Avatar: []byte{54, 23, 13}}
	p1.Loves = p2
	p2.Loves = p1
	pList := []*Person{p1, p2}
	govar.Dump(pList)

	// Self-dump
	defaultConfig := govar.DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           false,
		UseColors:           true,
		TrackReferences:     true,
		HTMLtagToken:        "span",
		HTMLtagSection:      "pre",
		EmbedTypeMethods:    false,
		ShowMetaInformation: false,
		ShowHexdump:         true,
	}
	govar.Dump(govar.NewDumper(defaultConfig))
}
