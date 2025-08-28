package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"boardshapes/boardshapes"
)

var rFlag = flag.Bool("r", false, "rflag this should resize an image")
var sFlag = flag.String("s", "", "sflag gets a user file output name.")
var mFlag = flag.Bool("m", false, "mflag simplifies an image into regions.")
var jFlag = flag.Bool("j", false, "jflag should meshify an image")

func fileOpenerDecoder(fileInput []string) (image.Image, error) {

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

func fileEncoder(img image.Image) *os.File {

	var outputPath string
	if *sFlag != "" {
		outputPath = *sFlag
	} else {
		outputPath = "output.png"
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == ".png" {
		err = png.Encode(outputFile, img)
	} else if ext == ".jpeg" || ext == ".jpg" {
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 100})

	}
	if err != nil {
		panic(err)
	}
	return outputFile
}

func image_output(fileToOutput *os.File, inputDir string) {

	outputDir := filepath.Join(inputDir, "output_files")
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}
	outputFilePath := filepath.Join(outputDir, filepath.Base(fileToOutput.Name()))
	fmt.Printf("Output file path: %s\n", outputFilePath)
}

func resize(rFlag *bool, imageResize image.Image) (flag *bool, img image.Image) {

	if rFlag != nil {
		imageResize, err := boardshapes.ResizeImage(imageResize) 
		if err != nil {
			fmt.Println("Error occurred")
			panic(err)
		}
		return rFlag, imageResize
	}
	return
}

/*func meshify(jFlag *bool, imageMeshify image.Image) (flag *bool, img image.Image) {

		if *jFlag {
			regionJunk := SimplifyImage(img, RegionMapOptions{})
			fmt.Println(regionJunk)

			for i := 0; i < regionCount-1; i++ {
				currRegion := regionToOutput.GetRegion(processing.RegionIndex(i))
				regionMeshCreated, err := currRegion.CreateMesh()

				if err != nil {
					panic(err)
				}
				for j := 0; j < len(regionMeshCreated); j++ {
					fmt.Println("X: ", regionMeshCreated[j].X, "Y: ", regionMeshCreated[j].Y)
				}
			}
		}
}
*/

func main() {

	flag.Parse()
	fileInput := flag.Args()
	imageProc, err := fileOpenerDecoder(fileInput)
	// take file input path
	if err != nil {
		panic(err)
	}

	rFlag, imageProc = resize(rFlag, imageProc)
	/*
		if rFlag != nil {}

			jFlag, img = meshify(jFlag)

		}
		if *mFlag {
			fileRegioned, regionCount, _ := processing.SimplifyImage(img, processing.RegionMapOptions{})
			fmt.Println(regionCount)

			outputFile := fileEncoder(fileRegioned)
			image_output(outputFile, filepath.Dir(fileInput[0]))
		} else {
			outputFile := fileEncoder(img)
			image_output(outputFile, filepath.Dir(fileInput[0]))
		}
	*/
	fmt.Println("1:", *rFlag)
	fmt.Println("2:", *sFlag)
	fmt.Println("3:", *mFlag)
	fmt.Println("3.5:", *jFlag)
	fmt.Println("4:", fileInput)
}
