package ico

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"image"
	"image/draw"
	"image/png"
	"io"
)

func Encode(w io.Writer, im image.Image) error {
	var (
		b = im.Bounds()
		m = image.NewRGBA(b)
	)

	draw.Draw(m, b, im, b.Min, draw.Src)

	var (
		header = head{0, 1, 1}
		entry  = direntry{Plane: 1, Bits: 32, Offset: 22} //nolint:mnd

		pngbuffer = new(bytes.Buffer)
		pngwriter = bufio.NewWriter(pngbuffer)
	)

	if err := png.Encode(pngwriter, m); err != nil {
		return err
	}

	if err := pngwriter.Flush(); err != nil {
		return err
	}

	entry.Size = uint32(len(pngbuffer.Bytes())) //nolint:gosec

	var bounds = m.Bounds()

	entry.Width = uint8(bounds.Dx())
	entry.Height = uint8(bounds.Dy())

	var bb = new(bytes.Buffer)

	if err := binary.Write(bb, binary.LittleEndian, header); err != nil {
		return err
	}

	if err := binary.Write(bb, binary.LittleEndian, entry); err != nil {
		return err
	}

	if _, err := w.Write(bb.Bytes()); err != nil {
		return err
	}

	if _, err := w.Write(pngbuffer.Bytes()); err != nil {
		return err
	}

	return nil
}
