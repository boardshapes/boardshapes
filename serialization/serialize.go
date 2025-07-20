package serialization

import (
	main "boardshapes/boardshapes"
	"io"
)

func BinarySerialize(w io.Writer, data main.BoardshapesData) error {
	// write version chunk
	_, err := w.Write([]byte{0})
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(main.VERSION))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte{0})
	if err != nil {
		return err
	}

	return nil
}

func BinaryDeserialize(r io.Reader) (*main.BoardshapesData, error) {
	return nil, nil
}
