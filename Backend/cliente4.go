package main

import (
	"encoding/gob"
	"fmt"
	"net"
)

type Parametros struct {
	KNearest string
	GRUPO    string
	EDAD     string
	SEXO     string
	DOSIS    string
	UBIGEO   string
	Eleccion string
	//RESULTADO    sortedClassVotes

}

func cliente(parametros Parametros) {
	c, err := net.Dial("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return

	}

	err = gob.NewEncoder(c).Encode(parametros)
	if err != nil {
		fmt.Println(err)
	}
	c.Close()

}

func main() {
	parametro := Parametros{
		KNearest: "3",
		GRUPO:    "2",
		EDAD:     "65",
		SEXO:     "0",
		DOSIS:    "2",
		UBIGEO:   "24",
	}
	go cliente(parametro)
	var input string
	fmt.Scanln(&input)

}
