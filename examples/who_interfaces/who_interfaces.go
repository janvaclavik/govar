package main

import (
	"fmt"

	"github.com/janvaclavik/govar"
	"github.com/janvaclavik/govar/who"
)

type Dreamer interface {
	Dream(int) int
}

type MyOtherType struct {
	isAlive bool
	cost    float32
}

func (o *MyOtherType) String() string {
	return "MyOtherType object"
}

func main() {
	listCodebase, err := who.Interfaces("github.com/janvaclavik/govar/main/main.MyOtherType")
	if err != nil {
		fmt.Println("ERROR who.Interfaces(): ", err)
	} else {
		fmt.Println("Satisfied interfaces in current codebase for MyOtherType: ", len(listCodebase))
		govar.Dump(listCodebase)
	}

	fmt.Println("InterfacesExt() for MyOtherType [has String() string]:")
	listExt, err := who.InterfacesExt("github.com/janvaclavik/govar/main/main.MyOtherType")
	if err != nil {
		fmt.Println("ERROR who.InterfacesExt(): ", err)
	} else {
		fmt.Println("Satisfied interfaces in STDlib + external imports for MyOtherType: ", len(listExt))
		govar.Dump(listExt)
	}

}
