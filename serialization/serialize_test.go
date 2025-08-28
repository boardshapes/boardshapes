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
			name: "no-masks",
			args: args{
				data: *main.CreateShapes(
					loadImage("../test_images/lub.png"),
					main.ShapeCreationOptions{}),
				options: &SerializationOptions{
					UseMasks: false,
				},
			},
		},
		{
			name: "masks",
			args: args{
				data: *main.CreateShapes(
					loadImage("../test_images/lub.png"),
					main.ShapeCreationOptions{}),
				options: &SerializationOptions{
					UseMasks: true,
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

			if equal, reason := data.Equal(*result); !equal {
				t.Errorf("Data mismatch: %v", reason)
			}
		})
	}
}

func TestJsonSerialization(t *testing.T) {
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
					loadImage("../test_images/lub.png"),
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
			b, err := JsonSerialize(&data)
			if err != nil {
				t.Errorf("JsonSerialize() error = %v", err)
				return
			}
			t.Logf("serialized shapes: %d", len(data.Shapes))

			result, err := JsonDeserialize(b, nil)
			if err != nil {
				t.Errorf("JsonDeserialize() error = %v", err)
			}
			t.Logf("deserialized shapes: %d", len(result.Shapes))

			if equal, reason := data.Equal(*result); !equal {
				t.Errorf("Data mismatch: %v", reason)
			}
		})
	}
}
