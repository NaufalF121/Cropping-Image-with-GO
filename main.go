package main

import (
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/sunshineplan/imgconv"
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

func separateRedChannel(img image.Image) *image.RGBA64 {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	redImage := image.NewRGBA64(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			r, _, _, a := img.At(x, y).RGBA()
			redImage.Set(x, y, color.RGBA64{R: uint16(r), G: 0, B: 0, A: uint16(a)})
		}
	}

	return redImage
}

func separateGreenChannel(img image.Image) *image.RGBA64 {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	greenImage := image.NewRGBA64(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			_, g, _, a := img.At(x, y).RGBA()
			greenImage.Set(x, y, color.RGBA64{R: 0, G: uint16(g), B: 0, A: uint16(a)})
		}
	}
	return greenImage
}

func separateBlueChannel(img image.Image) *image.RGBA64 {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	blueImage := image.NewRGBA64(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			_, _, b, a := img.At(x, y).RGBA()
			blueImage.Set(x, y, color.RGBA64{R: 0, G: 0, B: uint16(b), A: uint16(a)})
		}
	}

	return blueImage
}

func upload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := os.RemoveAll("./asset/gambar.jpeg")
	err = os.RemoveAll("./asset/gambar.png")
	err = os.RemoveAll("./asset/gambar.jpg")

	if _, err := os.Stat("./asset"); os.IsNotExist(err) {

		err := os.MkdirAll("./asset", 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		fmt.Println("Directory created successfully.")
	}

	if _, err := os.Stat("./output"); os.IsNotExist(err) {

		err := os.MkdirAll("./output", 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		fmt.Println("Directory created successfully.")
	}

	img, _, err := r.FormFile("file")

	if err != nil {
		panic(err.Error() + "Error in uploading file")
	}

	imgDecoded, _, err := image.Decode(img)
	if err != nil {
		panic(err.Error() + "Error decoding image file")
	}

	pngFile, err := os.Create("./asset/gambar.png")
	if err != nil {
		panic(err.Error() + "Error creating PNG file")
	}
	defer pngFile.Close()

	err = imgconv.Write(pngFile, imgDecoded, &imgconv.FormatOption{Format: imgconv.PNG})

	if err != nil {
		panic(err.Error() + "Error encoding image to PNG")
	}

	//contentType := http.DetectContentType(buffer)

	fileDir := "./asset/gambar.png"
	//file, err := os.Create(fileDir)
	//if err != nil {
	//	panic(err.Error() + "Error in creating file")
	//
	//}
	//defer file.Close()
	//buf := bytes.NewBuffer(nil)
	//if _, err := buf.ReadFrom(img); err != nil {
	//	panic(err.Error() + "Error in reading file")
	//}
	//if _, err := file.Write(buf.Bytes()); err != nil {
	//	panic(err.Error() + "Error in writing file")
	//}
	tmpl, err := template.ParseFiles("template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	imgFile, err := os.Open(fileDir)
	if err != nil {
		panic(err)
	}
	imek, err := png.Decode(imgFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	width := imek.Bounds().Dx()
	height := imek.Bounds().Dy()

	err = tmpl.ExecuteTemplate(w, "editing", Response{File: "asset/gambar.png", State: true, Width: width, Height: height})
	if err != nil {
		fmt.Println("Error in executing template")
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

	x1, _ := strconv.Atoi(r.PostFormValue("x1"))
	y1, _ := strconv.Atoi(r.PostFormValue("y1"))
	x2, _ := strconv.Atoi(r.PostFormValue("x2"))
	y2, _ := strconv.Atoi(r.PostFormValue("y2"))

	imgChannel := r.FormValue("imageChannel")

	files, err := ioutil.ReadDir("./asset")
	if err != nil {
		log.Fatal(err)

	}
	imgFile, err := os.Open("./asset" + "/" + files[0].Name())
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		img, err = png.Decode(imgFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	switch imgChannel {
	case "red":
		img = separateRedChannel(img)
	case "green":
		img = separateGreenChannel(img)
	case "blue":
		img = separateBlueChannel(img)
	case "grayscale":
		img = convertToGrayscale(img)
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	rect := image.Rect(x1, y1, x2, y2)
	croppedImage := rgba.SubImage(rect)
	outFile, err := os.Create("./output/cropped_gambar.png")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, croppedImage)
	if err != nil {
		return
	}
	tmpl, err := template.ParseFiles("template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "preview", Response{File: "output/cropped_gambar.png", State: true, Width: 0, Height: 0})
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
