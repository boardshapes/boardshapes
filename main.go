package boardshapes

import (
	"image"

	"github.com/boardshapes/boardshapes/colors"
	"github.com/boardshapes/boardshapes/data"
	"github.com/boardshapes/boardshapes/imageops"
	"github.com/boardshapes/boardshapes/internal/optimization"
	"github.com/boardshapes/boardshapes/internal/processing"
	"github.com/boardshapes/boardshapes/internal/regions"
)

const Version = "0.1.1"

type ShapeCreationOptions struct {
	NoColorSeparation,
	AllowWhite,
	PreserveColor,
	KeepSmallRegions bool
	EpsilonRDP float64
}

func isRegionLargeEnough(region *regions.Region) bool {
	const minimumNumberOfPixelsForNonSmallRegion = 50
	return len(*region) >= minimumNumberOfPixelsForNonSmallRegion
}

func CreateShapes(img image.Image, opts ShapeCreationOptions) (boardshapesData *data.BoardshapesData) {
	boardshapesData = &data.BoardshapesData{
		Version: Version,
	}

	img = imageops.ResizeImage(img)

	newImg := imageops.SimplifyImage(img, imageops.ImageSimplificationOptions{
		AllowWhite: opts.AllowWhite,
	})

	var filter func(*regions.Region) bool
	if opts.KeepSmallRegions {
		filter = nil
	} else {
		filter = isRegionLargeEnough
	}

	regionMap := regions.BuildRegionMap(newImg, regions.RegionMappingOptions{
		NoColorSeparation: opts.NoColorSeparation,
		AllowWhite:        opts.AllowWhite,
		PreserveColor:     opts.PreserveColor,
	}, filter)

	regions := regionMap.GetRegions()
	numRegions := len(regions)

	boardshapesData.Shapes = make([]data.ShapeData, 0, numRegions)

	for i := range numRegions {
		region := regionMap.GetRegionByIndex(i)

		minX, minY := region.FindRegionPosition()
		regionColor := region.GetColorOfRegion(newImg, opts.NoColorSeparation)
		var regionColorName string

		switch regionColor {
		case colors.Red:
			regionColorName = "Red"
		case colors.Green:
			regionColorName = "Green"
		case colors.Blue:
			regionColorName = "Blue"
		case colors.Black:
			regionColorName = "Black"
		case colors.White:
			regionColorName = "White"
		}

		regionImage := image.NewNRGBA(region.GetBounds())

		if opts.PreserveColor {
			for j := 0; j < len(*region); j++ {
				regionImage.Set(int((*region)[j].X), int((*region)[j].Y), img.At(int((*region)[j].X), int((*region)[j].Y)))
			}
		} else {
			for j := 0; j < len(*region); j++ {
				regionImage.Set(int((*region)[j].X), int((*region)[j].Y), regionColor)
			}
		}

		shape, err := processing.CreateShapeFromRegion(region)
		if err != nil {
			continue
		}

		if opts.EpsilonRDP == 0 {
			shape = optimization.OptimizeShape(shape)
		} else {
			shape = optimization.OptimizeShapeWithEpsilon(shape, opts.EpsilonRDP)
		}

		shapeData := data.ShapeData{
			Number:    i,
			Color:     regionColor,
			ColorName: regionColorName,
			CornerX:   minX,
			CornerY:   minY,
			Image:     regionImage,
			Path:      shape,
		}

		boardshapesData.Shapes = append(boardshapesData.Shapes, shapeData)
	}

	return
}
