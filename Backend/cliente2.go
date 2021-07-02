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
		KNearest: "5",
		GRUPO:    "9",
		EDAD:     "20",
		SEXO:     "0",
		DOSIS:    "2",
		UBIGEO:   "25",
	}
	go cliente(parametro)
	fmt.Println("Parametro",parametro)
	var input string
	fmt.Scanln(&input)

}
