package main

import (
	"html/template"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)
type Resulta struct {
	Eleccion     string
	RESULTADO    sortedClassVotes
}
var tpl *template.Template
var r sortedClassVotes

func init(){
	tpl = template.Must((template.ParseGlob("templates/*")))
}


func foo(w http.ResponseWriter, req *http.Request){

	err := tpl.ExecuteTemplate(w,"index.gohtml",Resulta{r[0].key,r})

	if err != nil{
		http.Error(w,err.Error(),500)
		log.Fatalln(err)
	}


}


type Parametros struct {
	KNearest string
	GRUPO    string
	EDAD     string
	SEXO     string
	DOSIS    string
	UBIGEO   string
	Eleccion string
}

// Escucha 
func servidor() {
	s, err := net.Listen("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			continue

		}
		go handleClient(c)

	}
}
func handleClient(c net.Conn) {
	var parametros Parametros
	err := gob.NewDecoder(c).Decode(&parametros)
	m := make(chan sortedClassVotes)
	wg := sync.WaitGroup{}
	wg.Add(2)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		wg.Done()

		kint, _ := strconv.Atoi(parametros.KNearest)
		gf, _ := strconv.ParseFloat(parametros.GRUPO, 64)
		ef, _ := strconv.ParseFloat(parametros.EDAD, 64)
		sf, _ := strconv.ParseFloat(parametros.SEXO, 64)
		df, _ := strconv.ParseFloat(parametros.DOSIS, 64)
		uf, _ := strconv.ParseFloat(parametros.UBIGEO, 64)

		time.Sleep(8 * time.Second)
		go Data(kint, gf, ef, sf, df, uf, m)

		msg := <-m
		fmt.Println(msg)
		r = msg


	}

}

type Vacunacion struct {
	GRUPO_RIESGO float64
	EDAD         float64
	SEXO         float64
	DOSIS        float64
	UBIGEO       float64
	FABRICANTE   string
}

// Lectura de dataset
func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ','

	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

//Conversion de la estructura
func parseVacunacion(record []string) Vacunacion {
	var vacuna Vacunacion

	vacuna.GRUPO_RIESGO, _ = strconv.ParseFloat(record[0], 64)
	vacuna.EDAD, _ = strconv.ParseFloat(record[1], 64)
	vacuna.SEXO, _ = strconv.ParseFloat(record[2], 64)
	vacuna.DOSIS, _ = strconv.ParseFloat(record[3], 64)
	vacuna.UBIGEO, _ = strconv.ParseFloat(record[4], 64)
	vacuna.FABRICANTE = record[5]

	return vacuna
}

type classVote struct {
	key   string
	value int
}

type sortedClassVotes []classVote

func (scv sortedClassVotes) Len() int           { return len(scv) }
func (scv sortedClassVotes) Less(i, j int) bool { return scv[i].value < scv[j].value }
func (scv sortedClassVotes) Swap(i, j int)      { scv[i], scv[j] = scv[j], scv[i] }

func getResponse(neighbors []Vacunacion) sortedClassVotes {
	classVotes := make(map[string]int)

	for x := range neighbors {
		response := neighbors[x].FABRICANTE
		if contains(classVotes, response) {
			classVotes[response] += 1
		} else {
			classVotes[response] = 1
		}
	}

	scv := make(sortedClassVotes, len(classVotes))
	i := 0
	for k, v := range classVotes {
		scv[i] = classVote{k, v}
		i++
	}

	sort.Sort(sort.Reverse(scv))
	return scv
}

type distancePair struct {
	record   Vacunacion
	distance float64
}

type distancePairs []distancePair

func (slice distancePairs) Len() int           { return len(slice) }
func (slice distancePairs) Less(i, j int) bool { return slice[i].distance < slice[j].distance }
func (slice distancePairs) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }

func getNeighbors(trainingSet []Vacunacion, testRecord Vacunacion, k int, c chan []Vacunacion) []Vacunacion {
	var distances distancePairs
	for i := range trainingSet {
		dist := Manhattan(testRecord, trainingSet[i])
		distances = append(distances, distancePair{trainingSet[i], dist})
	}

	sort.Sort(distances)

	var neighbors []Vacunacion

	for x := 0; x < k; x++ {
		neighbors = append(neighbors, distances[x].record)
	}

	c <- neighbors

	return neighbors

}

func Manhattan(instanceOne Vacunacion, instanceTwo Vacunacion) float64 {
	var distance float64

	distance += math.Abs(instanceOne.GRUPO_RIESGO-instanceTwo.GRUPO_RIESGO) +
		math.Abs(instanceOne.EDAD-instanceTwo.EDAD) +
		math.Abs(instanceOne.SEXO-instanceTwo.SEXO) +
		math.Abs(instanceOne.DOSIS-instanceTwo.DOSIS) +
		math.Abs(instanceOne.UBIGEO-instanceTwo.UBIGEO)

	return distance
}

func errHandle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func contains(votesMap map[string]int, name string) bool {
	for s, _ := range votesMap {
		if s == name {
			return true
		}
	}

	return false
}

func Data(k int, Grupo float64, Edad float64, Sexo float64, Dosis float64, Ubigeo float64, m chan sortedClassVotes) {
	url := "https://raw.githubusercontent.com/Furtherron/TA2-Concurrente/main/Vacunacion.csv"

	var recordSet []Vacunacion

	data, err := readCSVFromUrl(url)

	if err != nil {
		panic(err)
	}

	for idx, row := range data {
		if idx == 0 {
			continue
		}

		recordSet = append(recordSet, parseVacunacion(row))

	}
	var testSet []Vacunacion
	var trainSet []Vacunacion
	for i := range recordSet {

		trainSet = append(trainSet, recordSet[i])

	}

	testSet = append(testSet, Vacunacion{Grupo, Edad, Sexo, Dosis, Ubigeo, ""})

	c := make(chan []Vacunacion)

	go getNeighbors(trainSet, testSet[0], k, c)

	neighbors := <-c
	result := getResponse(neighbors)

	fmt.Printf("Actual: %s\n", result[0].key)

	m <- result
	

}
func h(){
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/",foo)
	http.ListenAndServe(":9998", nil)
}
func main() {

	go h()
	go servidor()
	var input string
	fmt.Scanln(&input)
	
}
