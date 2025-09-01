package main

import (
	"boardshapes/boardshapes"
	"boardshapes/boardshapes/serialization"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var resizeImage string
var outputSimplifiedImagePath string
var noShapes bool
var binaryOutput bool
var outputPath string
var useStdOut bool
var optimizeShapeEpsilon float64

func init() {
	const resizeFlagDescription = "Resize the image to fit a specific size while maintaining aspect ratio. " +
		"Value should be in the format [width]x[height] where both width and height " +
		"are optional and can be left empty. If neither are specified, it will default to 1920x1080."
	flag.StringVar(&resizeImage, "r", "no", resizeFlagDescription)
	flag.StringVar(&resizeImage, "resize", "no", resizeFlagDescription)

	const outputSimplifiedFlagDescription = "If set to a valid filepath, the simplified image generated will " +
		"be output to that filepath."
	flag.StringVar(&outputSimplifiedImagePath, "s", "", outputSimplifiedFlagDescription)
	flag.StringVar(&outputSimplifiedImagePath, "output-simplified", "", outputSimplifiedFlagDescription)

	const noShapesFlagDescription = "Skips the shape generation step and does not output any shape data."
	flag.BoolVar(&noShapes, "x", false, noShapesFlagDescription)
	flag.BoolVar(&noShapes, "no-shapes", false, noShapesFlagDescription)

	const binaryFlagDescription = "Serializes shape data to the binary format instead of JSON."
	flag.BoolVar(&binaryOutput, "b", false, binaryFlagDescription)
	flag.BoolVar(&binaryOutput, "binary", false, binaryFlagDescription)

	const outputFileFlagDescription = "Path to the output file"
	flag.StringVar(&outputPath, "o", "output.bshapes", outputFileFlagDescription)
	flag.StringVar(&outputPath, "output", "output.bshapes", outputFileFlagDescription)

	const useStdOutFlagDescription = "If set, the output will be written to stdout instead of a file."
	flag.BoolVar(&useStdOut, "c", false, useStdOutFlagDescription)
	flag.BoolVar(&useStdOut, "stdout", false, useStdOutFlagDescription)

	const optimizeShapeEpsilonDescription = "Sets the epsilon value for the Ramer-Douglas-Peucker optimization. " +
		"Generally, a smaller epsilon value will result in a more detailed shape, while a larger epsilon value will " +
		"result in a less complex shape. Will use the default epsilon value if not specified or set to 0." +
		"Will skip RDP optimization entirely if set to a negative value, but will never skip basic straight-line optimization."
	flag.Float64Var(&optimizeShapeEpsilon, "e", 0.0, optimizeShapeEpsilonDescription)
	flag.Float64Var(&optimizeShapeEpsilon, "epsilon", 0.0, optimizeShapeEpsilonDescription)
}

func main() {

	flag.Parse()
	fileInput := flag.Args()
	img, err := decodeImageFromFile(fileInput)

	if err != nil {
		panic(err)
	}

	img = resize(resizeImage, img)

	outputSimpified(outputSimplifiedImagePath, img)

	boardShapeData := noShape(noShapes, img, optimizeShapeEpsilon)

	r, err := serialization.JsonSerialize(boardShapeData)

	

}

func decodeImageFromFile(fileInput []string) (image.Image, error) {
	joinedFileName := strings.Join(fileInput, "")

	fileTaken, err := os.Open(joinedFileName)
	if err != nil {
		panic(err)
	}
	defer fileTaken.Close()

	fileExtension := filepath.Ext(joinedFileName)
	if fileExtension == ".jpg" {
		fileExtension = ".jpeg"
	}
	if fileExtension == ".png" || fileExtension == ".jpeg" {
		img, _, err := image.Decode(fileTaken)
		if err != nil {
			panic(err)
		}

		return img, nil
	}
	return nil, fmt.Errorf("unsuppported file format")
}

func encodeImageToFile(img image.Image, outPath string) *os.File {
	err := os.MkdirAll(filepath.Dir(outPath), 0755)
	if err != nil {
		panic(err)
	}
	outputFile, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}

	ext := strings.ToLower(filepath.Ext(outPath))
	switch ext {
	case ".png":
		err = png.Encode(outputFile, img)
	case ".jpeg", ".jpg":
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 100})
	}
	if err != nil {
		panic(err)
	}
	log.Printf("Output file path: %s\n", outPath)
	return outputFile
}

func resize(resizeImage string, img image.Image) image.Image {

	if resizeImage != "no" {
		if resizeImage == "" {
			img = boardshapes.ResizeImage(img)
		}
		dimensions := strings.Split(resizeImage, "x")
		var width, height int
		if len(dimensions) != 2 {
			panic(errors.New("invalid resize format: Use [width]x[height], e.g. 800x600, 800x, x600"))
		}
		if dimensions[0] != "" {
			var err error
			width, err = strconv.Atoi(dimensions[0])
			if err != nil {
				panic(errors.New("invalid width value"))
			}
		}
		if dimensions[1] != "" {
			var err error
			height, err = strconv.Atoi(dimensions[1])
			if err != nil {
				panic(errors.New("invalid height value"))
			}
		}
		if width == 0 && height == 0 {
			img = boardshapes.ResizeImage(img)
		} else {
			img = boardshapes.ResizeImageTo(img, width, height)
		}

		return img
	}
	return img
}

func outputSimpified(outputSimplifiedImagePath string, img image.Image) {
	if outputSimplifiedImagePath != "" {
		simplifiedImage := boardshapes.SimplifyImage(img, boardshapes.ShapeCreationOptions{})

		encodeImageToFile(simplifiedImage, outputSimplifiedImagePath)
	}
}

func outputPathArg(outputFile string, boarddata io.Reader) {
	if outputFile != "" {
		f, err := os.Create(outputFile)

		if err != nil {
			panic(err)
		}
		defer f.Close()

		_, err = io.Copy(f, boarddata)
		if err != nil {
			panic(err)
		}
	}
}

func noShape(noShapesFlag bool, img image.Image, optimizeShapeEpsilon float64) (data *boardshapes.BoardshapesData) {
	if noShapesFlag == false {
		boarddata := boardshapes.CreateShapes(img, boardshapes.ShapeCreationOptions{EpsilonRDP: optimizeShapeEpsilon})
		return boarddata
	}

	return nil
}
