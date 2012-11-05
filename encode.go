package tga

import (
	"encoding/binary"
	"errors"
	"image"
	"image/draw"
	"io"
)

// Encode encodes an image into TARGA format.
func Encode(w io.Writer, m image.Image) (err error) {
	b := m.Bounds()
	mw, mh := b.Dx(), b.Dy()

	h := rawHeader{
		Width:  uint16(mw),
		Height: uint16(mh),
	}

	if int(h.Width) != mw || int(h.Height) != mh {
		return errors.New("uint16 width/height overflow")
	}

	h.Flags = flagOriginTop

	switch tm := m.(type) {
	case *image.Gray:
		h.ImageType = imageTypeMonoChrome
		err = encodeGray(w, tm, h)

	case *image.NRGBA:
		h.ImageType = imageTypeTrueColor
		err = encodeRGBA(w, tm, h, attrTypeAlpha)

	case *image.RGBA:
		h.ImageType = imageTypeTrueColor
		err = encodeRGBA(w, (*image.NRGBA)(tm), h, attrTypePremultipliedAlpha)

	default:
		// convert to non-premultiplied alpha by default
		newm := image.NewNRGBA(b)
		draw.Draw(newm, b, m, b.Min, draw.Src)
		err = encodeRGBA(w, newm, h, attrTypeAlpha)
	}

	return
}

func encodeGray(w io.Writer, m *image.Gray, h rawHeader) (err error) {
	h.BPP = 8 // 8-bit monochrome

	if err = binary.Write(w, binary.LittleEndian, &h); err != nil {
		return
	}

	offset := -(m.Rect.Min.Y*m.Stride + m.Rect.Min.X)
	max := offset + int(h.Height)*m.Stride

	for ; offset < max; offset += m.Stride {
		if _, err = w.Write(m.Pix[offset : offset+int(h.Width)]); err != nil {
			return
		}
	}

	// no extension area, only a footer
	err = binary.Write(w, binary.LittleEndian, newFooter())

	return
}

func encodeRGBA(w io.Writer, m *image.NRGBA, h rawHeader, attrType byte) (err error) {
	h.BPP = 32   // always save as 32-bit (faster this way)
	h.Flags |= 8 // 8-bit alpha channel

	if err = binary.Write(w, binary.LittleEndian, &h); err != nil {
		return
	}

	lineSize := int(h.Width) * 4
	offset := -m.Rect.Min.Y*m.Stride - m.Rect.Min.X*4
	max := offset + int(h.Height)*m.Stride
	b := make([]byte, lineSize)

	for ; offset < max; offset += m.Stride {
		copy(b, m.Pix[offset:offset+lineSize])

		for i := 0; i < lineSize; i += 4 {
			b[i+0], b[i+2] = b[i+2], b[i+0] // RGBA -> BGRA
		}

		if _, err = w.Write(b); err != nil {
			return
		}
	}

	// add extension area and footer to define attribute type
	_, err = w.Write(newExtArea(attrType))
	footer := newFooter()
	footer.ExtAreaOffset = uint32(tgaRawHeaderSize + int(h.Height)*lineSize)
	err = binary.Write(w, binary.LittleEndian, footer)

	return
}
