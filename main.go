package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

type (
	Query struct {
		X      int      `json:"x"`
		Y      int      `json:"y"`
		Images []string `json:"images"`
	}
	Response struct {
		Points []uint32 `json:"points"`
	}
)

func main() {

	fileConverter("./kml.png")
	//testFileConverter("./kml.converted")

	/*r := chi.NewRouter()

	r.Post("/", ImageHandler)
	http.ListenAndServe(":3000", r)*/
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
		return
	}

	w.Write(buf)

}

func getPointValue(fileName string, queryX, queryY int) (point uint32) {

	if queryY < 0 || queryX < 0 {
		fmt.Println("Invalid point coordinates")
		return
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)

	defer file.Close()

	if err != nil {
		fmt.Printf("Can`t read file : %s\n", err.Error())
		return
	}

	image, err := png.Decode(file)

	if err != nil {
		fmt.Printf("Can`t read file, it`s not png : %s\n", err.Error())
		return
	}

	rect := image.Bounds()

	if rect.Max.Y < queryY || rect.Max.X < queryX {
		fmt.Println("Invalid query point size")
		return
	}

	var width, height uint32

	binary.Read(file, binary.BigEndian, width)
	binary.Read(file, binary.BigEndian, height)

	offset := int64((queryX * int(height)) + queryY)

	file.Seek(offset, io.SeekCurrent)

	binary.Read(file, binary.BigEndian, point)
	return point
}

func fileConverter(fileName string) {

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)

	if err != nil {
		fmt.Printf("Can`t read file : %s\n", err.Error())
		return
	}

	defer file.Close()

	image, err := png.Decode(file)

	if err != nil {
		fmt.Printf("Can`t read file, it`s not png : %s\n", err.Error())
		return
	}

	rect := image.Bounds()

	width := make([]byte, 4)
	height := make([]byte, 4)

	buf := bytes.Buffer{}

	binary.BigEndian.PutUint32(width, uint32(rect.Max.X))
	binary.BigEndian.PutUint32(height, uint32(rect.Max.Y))

	fmt.Println(width)
	fmt.Println(height)

	buf.Write(width)
	buf.Write(height)

	for y := 0; y < rect.Max.Y; y++ {

		for x := 0; x < rect.Max.X; x++ {

			c, _ := color.NRGBAModel.Convert(rect.At(x, y)).(color.NRGBA)

			pointByte := []byte{c.R, c.G, c.B, c.B}

			fmt.Println(c.R, c.G, c.B, c.B)

			buf.Write(pointByte)
		}
	}

	newFile, err := os.Create(strings.TrimSuffix(fileName, ".png") + ".converted")

	if _, err := newFile.Write(buf.Bytes()); err != nil {
		fmt.Printf("Can`t write converted file: %s\n", err.Error())
	}

	newFile.Close()
}

func testFileConverter(fileName string) {

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)

	if err != nil {
		fmt.Printf("Can`t read file : %s\n", err.Error())
		return
	}

	defer file.Close()

	var width = make([]byte, 4)
	var height = make([]byte, 4)

	binary.Read(file, binary.BigEndian, width)
	binary.Read(file, binary.BigEndian, height)

	newRect := image.Rectangle{}
	newRect.Max.X = int(binary.BigEndian.Uint32(width))
	newRect.Max.Y = int(binary.BigEndian.Uint32(height))

	im := image.NewNRGBA(newRect)

	for y := newRect.Max.X; y > 0; y-- {
		for x := newRect.Max.Y; x > 0; x-- {
			buf := make([]uint8, 4)
			if err := binary.Read(file, binary.BigEndian, buf); err != nil {
				return
			}
			colour := color.NRGBA{R: buf[0], G: buf[1], B: buf[2], A: buf[3]}
			im.Set(int(x), int(y), colour)
		}
	}

	fileTest, err := os.Create("test.png")

	if err != nil {
		fmt.Println("Can`t create test file", err.Error())
	}

	png.Encode(fileTest, im)

	fileTest.Close()
}
