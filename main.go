package main

import (
	"bytes"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Response struct {
	File   string
	State  bool
	Width  int
	Height int
}

func convertToGrayscale(img image.Image) *image.Gray {
	// Create a new grayscale image
	grayImage := image.NewGray(img.Bounds())

	// Draw the source image onto the destination image
	draw.Draw(grayImage, grayImage.Bounds(), img, img.Bounds().Min, draw.Src)

	return grayImage
}

func upload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	img, _, err := r.FormFile("file")

	if err != nil {
		panic(err.Error() + "Error in uploading file")
	}

	buffer := make([]byte, 512)
	_, err = img.Read(buffer)
	if err != nil {
		panic(err.Error() + "Error reading file")
	}

	// Reset the read pointer to the start of the file
	_, err = img.Seek(0, 0)
	if err != nil {
		panic(err.Error() + "Error in seeking file")
	}

	// Get the content type of the file
	contentType := http.DetectContentType(buffer)

	if _, err := os.Stat("./asset"); os.IsNotExist(err) {
		// Create the directory
		err := os.MkdirAll("./asset", 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		fmt.Println("Directory created successfully.")
	}
	fileDir := "./asset/gambar." + contentType[6:]
	file, err := os.Create(fileDir)
	if err != nil {
		panic(err.Error() + "Error in creating file")

	}
	defer file.Close()
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(img); err != nil {
		panic(err.Error() + "Error in reading file")
	}
	if _, err := file.Write(buf.Bytes()); err != nil {
		panic(err.Error() + "Error in writing file")
	}
	tmpl, err := template.ParseFiles("template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	imgFile, err := os.Open(fileDir)
	imek, _, err := image.Decode(imgFile)
	if err != nil {
		imek, err = png.Decode(imgFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	width := imek.Bounds().Dx()
	height := imek.Bounds().Dy()

	err = tmpl.ExecuteTemplate(w, "editing", Response{File: "asset/gambar." + contentType[6:], State: true, Width: width, Height: height})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func renderTemplate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl, err := template.ParseFiles("template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func cropper(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	time.Sleep(1 * time.Second)
	x1, _ := strconv.Atoi(r.PostFormValue("x1"))
	y1, _ := strconv.Atoi(r.PostFormValue("y1"))
	x2, _ := strconv.Atoi(r.PostFormValue("x2"))
	y2, _ := strconv.Atoi(r.PostFormValue("y2"))
	gray := r.FormValue("grayscale") == "on"
	fmt.Println(r.PostFormValue("input1"), y1, x2, y2, gray)
	files, err := ioutil.ReadDir("./asset")
	if err != nil {
		log.Fatal(err)

	}
	imgFile, err := os.Open("./asset" + "/" + files[0].Name())
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()

	// Decode the image file
	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}
	if gray {
		img = convertToGrayscale(img)
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Define the rectangle for cropping
	rect := image.Rect(x1, y1, x2, y2)
	croppedImage := rgba.SubImage(rect)
	outFile, err := os.Create("./output/cropped_gambar.jpeg")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, croppedImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
	if err != nil {
		return
	}
	tmpl, err := template.ParseFiles("template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "preview", Response{File: "output" + "/" + files[0].Name(), State: true, Width: 0, Height: 0})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}

}

func main() {
	router := httprouter.New()

	router.POST("/upload", upload)
	router.GET("/", renderTemplate)
	router.POST("/cropper", cropper)
	router.ServeFiles("/asset/*filepath", http.Dir("./asset"))
	router.ServeFiles("/output/*filepath", http.Dir("./output"))

	log.Fatal(http.ListenAndServe(":8080", router))
}
