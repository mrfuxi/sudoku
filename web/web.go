package web

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image"
	_ "image/jpeg" // Enable processing JPEG files
	"image/png"
	"log"
	"net/http"

	"github.com/mrfuxi/sudoku"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func init() {
	http.HandleFunc("/", upload)
}

func imageToBase64(img image.Image) string {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func processForm(req *http.Request, context map[string]string) {
	file, handler, err := req.FormFile("uploadfile")
	if err != nil {
		context["Error"] = "Sudoku file missing"
		log.Println("Could not open sudoku file.", err.Error())
		return
	}

	if len(handler.Header["Content-Type"]) > 0 {
		context["ContentType"] = handler.Header["Content-Type"][0]
	}

	img, _, err := image.Decode(file)
	if err != nil {
		context["Error"] = "Could not read the file"
		log.Println("Could not read the file.", err.Error())
		return
	}
	s, err := sudoku.NewSudoku(img)
	if err != nil {
		context["Error"] = err.Error()
		log.Println(err.Error())
		return
	}
	context["Image"] = imageToBase64(s.Overlay())
}

func upload(rw http.ResponseWriter, req *http.Request) {
	context := map[string]string{}
	if req.Method == "POST" {
		processForm(req, context)
	}

	if err := templates.ExecuteTemplate(rw, "sudoku.html", context); err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
