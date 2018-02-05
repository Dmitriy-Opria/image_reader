package main

import (
	"github.com/go-chi/chi"
	"net/http"
	"encoding/json"
	"os"
	"fmt"
	"encoding/binary"
	"io"
)
type (
	Query struct {
		X int	`json:"x"`
		Y int	`json:"y"`
		Images []string `json:"images"`
	}
	Response struct {
		Points []uint32 `json:"points"`
	}
)


func main (){
	r := chi.NewRouter()

	r.Post("/", ImageHandler)
	http.ListenAndServe(":3000", r)
}


func ImageHandler(w http.ResponseWriter, r *http.Request) {

	queryBody := new(Query)

	json.NewDecoder(r.Body).Decode(queryBody)

	if len(queryBody.Images) > 0 {
		return
	}

	res := Response{}

	res.Points = make([]uint32, 0, len(queryBody.Images))

	for _, fileName := range queryBody.Images {

		point := getPointValue(fileName, queryBody.X, queryBody.Y)

		res.Points = append(res.Points, point)
	}

	buf, err := json.Marshal(&res)

	if err != nil {
		fmt.Println("Can`t marshal json")
	}

	w.Write(buf)

}

func getPointValue(fileName string, queryX, queryY int) uint32{
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)

	if err != nil {
		fmt.Printf("Can`t read file : %s", err.Error())
	}

	var width, height uint32

	binary.Read(file, binary.BigEndian, width)
	binary.Read(file, binary.BigEndian, height)

	offset := int64((queryY * int(height)) + queryX)

	var point uint32

	file.Seek(offset, io.SeekCurrent)
	
	binary.Read(file, binary.BigEndian, point)
	return point
}