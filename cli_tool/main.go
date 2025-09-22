package main

import (
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
	"syscall/js"
	"github.com/boardshapes/boardshapes"
	"github.com/boardshapes/boardshapes/serialization"
)

var resizeImage string
var mode string
var binaryOutput bool
var outputPath string
var useStdOut bool
var optimizeShapeEpsilon float64

func init() {
	const resizeFlagDescription = "Resize any input image to fit a specific size while maintaining aspect ratio. " +
		"Value should be in the format [width]x[height] where both width and height " +
		"are optional and can be left empty. If neither are specified, it will default to 1920x1080."
	flag.StringVar(&resizeImage, "r", "no", resizeFlagDescription)
	flag.StringVar(&resizeImage, "resize", "no", resizeFlagDescription)

	const modeFlagDescription = "Determines what operation should be performed:\n" +
		"- \"g\"/\"generate\" (default) -> Generate shapes from an image file and output serialized Boardshapes data." +
		"- \"r\"/\"reserialize\" -> Deserialize data from a Boardshapes data file and then output the data after serializing it again. " +
		"Useful for converting between binary and JSON formats or upgrading old Boardshapes data to the latest version." +
		"- \"s\"/\"simplify\" -> Simplifies the color palette of an image file, giving you a preview of what color " +
		"each pixel is classified as when generating shapes."
	flag.StringVar(&mode, "m", "generate", modeFlagDescription)
	flag.StringVar(&mode, "mode", "generate", modeFlagDescription)

	const binaryFlagDescription = "Serializes shape data to the binary format instead of JSON."
	flag.BoolVar(&binaryOutput, "b", false, binaryFlagDescription)
	flag.BoolVar(&binaryOutput, "binary", false, binaryFlagDescription)

	const outputFileFlagDescription = "Path to the output file"
	flag.StringVar(&outputPath, "o", "", outputFileFlagDescription)
	flag.StringVar(&outputPath, "output", "", outputFileFlagDescription)

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

	w, shouldClose := getOutputWriter()
	if shouldClose {
		defer w.Close()
	}

	switch mode {
	case "g", "generate":
		img := getInputImage()
		boardShapesData := boardshapes.CreateShapes(img, boardshapes.ShapeCreationOptions{
			EpsilonRDP: optimizeShapeEpsilon,
		})

		serializeDataToWriter(w, boardShapesData)
	case "s", "simplify":
		img := getInputImage()
		outputSimplifiedImageToWriter(w, img)
	case "r", "reserialize":
		var boardShapesData *boardshapes.BoardshapesData
		boardShapesData = getInputData()
		serializeDataToWriter(w, boardShapesData)
	default:
		log.Fatalf("unknown mode: %s\n", mode)
	}

}

func serializeDataToWriter(w io.Writer, boardShapesData *boardshapes.BoardshapesData) {
	if binaryOutput {
		err := serialization.BinarySerialize(w, boardShapesData, nil)
		if err != nil {
			panic(err)
		}
	} else {
		err := serialization.JsonSerialize(w, boardShapesData)
		if err != nil {
			panic(err)
		}
	}
}

func getOutputWriter() (w io.WriteCloser, shouldClose bool) {
	if outputPath == "" {
		outputPath = getDefaultOutputFilename()
	}
	if useStdOut {
		w = os.Stdout
		shouldClose = false
	} else {
		err := os.MkdirAll(filepath.Dir(outputPath), 0755)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(outputPath)
		if err != nil {
			panic(err)
		}
		w = f
		shouldClose = true
	}

	return
}

func getDefaultOutputFilename() string {
	switch mode {
	case "g", "generate", "r", "reserialize":
		if binaryOutput {
			return "output.bshapes"
		} else {
			return "output.jshapes"
		}
	case "s", "simplify":
		return "output.png"
	default:
		log.Fatalf("unknown mode: %s\n", mode)
		return ""
	}
}

func getInputReader() io.ReadSeeker {
	args := flag.Args()


	if len(args) == 0 {
		panic(errors.New("no input file specified"))
	}
    stdInCheck := args[0] 
	if stdInCheck == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		return bytes.NewReader(data)
	}

	fileName := strings.Join(args, " ")
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	return f
}

func getInputImage() image.Image {
	r := getInputReader()

	img := decodeImageFromFile(r)
	img = resize(img)
	return img
}

func getInputData() *boardshapes.BoardshapesData {
	r := getInputReader()

	boardShapesData := deserializeBoardshapesData(r)
	return boardShapesData
}

func deserializeBoardshapesData(r io.ReadSeeker) *boardshapes.BoardshapesData {
	var boardShapesData *boardshapes.BoardshapesData
	var err error

	format := detectDataFormat(r)
	switch format {
	case "json":
		boardShapesData, err = serialization.JsonDeserialize(r, nil)
		if err != nil {
			panic(err)
		}
	case "binary":
		boardShapesData, err = serialization.BinaryDeserialize(r, nil)
		if err != nil {
			panic(err)
		}
	default:
		log.Fatalf("unknown data format: %s\n", format)
	}

	return boardShapesData
}

// todo: this should probably be in the serialization package.
func detectDataFormat(r io.ReadSeeker) string {
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		panic(err)
	}
	r.Seek(-1, io.SeekCurrent)

	if buf[0] == '{' {
		return "json"
	} else {
		return "binary"
	}
}

func decodeImageFromFile(r io.Reader) image.Image {
	img, _, err := image.Decode(r)
	if err != nil {
		panic(err)
	}

	return img
}

func encodeImageToWriter(w io.Writer, img image.Image) {
	ext := strings.ToLower(filepath.Ext(outputPath))
	var err error
	switch ext {
	case ".png":
		err = png.Encode(w, img)
	case ".jpeg", ".jpg":
		err = jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
	default:
		err = fmt.Errorf("unsupported file format: %s", ext)
	}
	if err != nil {
		panic(err)
	}
}

func resize(img image.Image) image.Image {

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

func outputSimplifiedImageToWriter(w io.Writer, img image.Image) {
	simplifiedImage := boardshapes.SimplifyImage(img, boardshapes.ShapeCreationOptions{EpsilonRDP: optimizeShapeEpsilon})

	encodeImageToWriter(w, simplifiedImage)
}
