package serialization

import (
	main "boardshapes/boardshapes"
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

func loadImage(filepath string) image.Image {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

func TestBinarySerialization(t *testing.T) {
	type args struct {
		data    main.BoardshapesData
		options *SerializationOptions
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "lub",
			args: args{
				data: *main.CreateShapes(
					loadImage("../build_region_map_test_images/lub.png"),
					main.ShapeCreationOptions{}),
				options: &SerializationOptions{
					UseMasks: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.args.data
			w := &bytes.Buffer{}
			if err := BinarySerialize(w, data, tt.args.options); err != nil {
				t.Errorf("BinarySerialize() error = %v", err)
				return
			}
			t.Logf("serialized shapes: %d", len(data.Shapes))

			result, err := BinaryDeserialize(w, nil)
			if err != nil {
				t.Errorf("BinaryDeserialize() error = %v", err)
			}
			t.Logf("deserialized shapes: %d", len(result.Shapes))

			if data.Version != result.Version {
				t.Errorf("Version mismatch: got %v, want %v", result.Version, data.Version)
				return
			}

		outer:
			for _, outShape := range data.Shapes {
				for _, inShape := range result.Shapes {
					if inShape.Equal(outShape) {
						continue outer
					}
				}
				t.Errorf("Shape has no matching shape: %d", outShape.Number)
				return
			}
		})
	}
}
