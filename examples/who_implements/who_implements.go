package main

import (
	"fmt"

	"github.com/janvaclavik/govar"
	"github.com/janvaclavik/govar/who"
)

type Dreamer interface {
	Dream(int) int
}

type MyType struct {
	id   int
	name string
	text string
}

func (o *MyType) Dream(int) int {
	return 5
}

func (o *MyType) String() string {
	return "MyType object"
}

type MyOtherType struct {
	isAlive bool
	cost    float32
}

func (o *MyOtherType) String() string {
	return "MyOtherType object"
}

func main() {

	listTypes, err := who.Implements("github.com/janvaclavik/govar/examples/who_implements.Dreamer")
	if err != nil {
		fmt.Println("ERROR who.Implements(): ", err)
	} else {
		fmt.Println("Types implementing 'example.Dreamer' interface in current project + STDlib + external imports: ", len(listTypes))
		govar.Dump(listTypes)
	}
	fmt.Println()
	listTypes2, err := who.Implements("fmt.Stringer")
	if err != nil {
		fmt.Println("ERROR who.Implements(): ", err)
	} else {
		fmt.Println("Types implementing 'fmt.Stringer' interface in current project + STDlib + external imports: ", len(listTypes2))
		govar.Dump(listTypes2)
	}

}
