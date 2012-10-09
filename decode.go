package tga

import (
  "bytes"
  "encoding/binary"
  "errors"
  "image"
  "image/color"
  "io"
)

type rawHeader struct {
  IdLength      uint8
  PaletteType   uint8
  ImageType     uint8
  PaletteFirst  uint16
  PaletteLength uint16
  PaletteBPP    uint8
  OriginX       uint16
  PriginY       uint16
  Width         uint16
  Height        uint16
  BPP           uint8
  Flags         uint8
}

type rawFooter struct {
  ExtAreaOffset uint32
  DevDirOffset  uint32
  Signature     [18]byte // tgaSignature
}

type tga struct {
  r             *bytes.Reader
  raw           rawHeader
  isPaletted    bool
  hasAlpha      bool
  width         int
  height        int
  pixelSize     int
  palette       []byte
  paletteLength int
  ColorModel    color.Model
  tmp           [4]byte
  decode        func(tga *tga, out []byte) (err error)
}

const (
  flagOriginRight = uint8(1 << 4)
  flagOriginTop   = 1 << 5
  flagMask        = 7
)

const (
  imageTypePaletted   = uint8(1)
  imageTypeTrueColor  = 2
  imageTypeMonoChrome = 3
  imageTypeRLE        = 1 << 3
  imageTypeMask       = 3
)

const (
  tgaRawHeaderSize = 18
  tgaRawFooterSize = 26
  tgaSignature     = "TRUEVISION-XFILE.\x00"
)

func decode(r io.Reader) (outImage image.Image, err error) {
  var tga tga
  var data bytes.Buffer

  if _, err = data.ReadFrom(r); err != nil {
    return
  }

  tga.r = bytes.NewReader(data.Bytes())

  if err = tga.getHeader(); err != nil {
    return
  }

  // skip header
  if _, err = tga.r.Seek(int64(tgaRawHeaderSize+tga.raw.IdLength), 0); err != nil {
    return
  }

  if tga.isPaletted {
    // read palette
    entrySize := int((tga.raw.PaletteBPP + 1) >> 3)
    tga.paletteLength = int(tga.raw.PaletteLength - tga.raw.PaletteFirst)
    tga.palette = make([]byte, entrySize*tga.paletteLength)

    // skip to colormap
    if _, err = tga.r.Seek(int64(entrySize)*int64(tga.raw.PaletteFirst), 1); err != nil {
      return
    }

    if _, err = io.ReadFull(tga.r, tga.palette); err != nil {
      return
    }
  }

  rect := image.Rect(0, 0, tga.width, tga.height)
  var pixels []byte

  // choose a right color model
  if tga.ColorModel == color.NRGBAModel {
    im := image.NewNRGBA(rect)
    outImage = im
    pixels = im.Pix
  } else {
    im := image.NewRGBA(rect)
    outImage = im
    pixels = im.Pix
  }

  if err = tga.decode(&tga, pixels); err == nil {
    tga.flip(pixels)
  }

  return
}

func decodeConfig(r io.Reader) (cfg image.Config, err error) {
  var tga tga
  var data bytes.Buffer

  if _, err = data.ReadFrom(r); err != nil {
    return
  }

  tga.r = bytes.NewReader(data.Bytes())

  if err = tga.getHeader(); err == nil {
    cfg = image.Config{
      ColorModel: tga.ColorModel,
      Width:      tga.width,
      Height:     tga.height,
    }
  }

  return
}

func init() {
  image.RegisterFormat("tga", "", decode, decodeConfig)
}

// applyExtensions reads extensions section (if it exists) and parses attribute type.
func (tga *tga) applyExtensions() (err error) {
  var rawFooter rawFooter

  if _, err = tga.r.Seek(int64(-tgaRawFooterSize), 2); err != nil {
    return
  } else if err = binary.Read(tga.r, binary.LittleEndian, &rawFooter); err != nil {
    return
  } else if bytes.Equal(rawFooter.Signature[:], []byte(tgaSignature)) && rawFooter.ExtAreaOffset != 0 {
    offset := int64(rawFooter.ExtAreaOffset + 0x1ee)

    var n int64
    var t byte

    if n, err = tga.r.Seek(offset, 0); err != nil || n != offset {
      return
    } else if t, err = tga.r.ReadByte(); err != nil {
      return
    } else if t == 3 {
      // alpha
      tga.hasAlpha = true
      tga.ColorModel = color.NRGBAModel
    } else if t == 4 {
      // premultiplied alpha
      tga.hasAlpha = true
      tga.ColorModel = color.RGBAModel
    } else {
      // attribute is not an alpha channel value, ignore it
      tga.hasAlpha = false
    }
  }

  return
}

// decodeRaw decodes a raw (uncompressed) data.
func decodeRaw(tga *tga, out []byte) (err error) {
  for i := 0; i < len(out) && err == nil; i += 4 {
    err = tga.getPixel(out[i:])
  }

  return
}

// decodeRLE decodes run-length encoded data.
func decodeRLE(tga *tga, out []byte) (err error) {
  size := tga.width * tga.height * 4

  for i := 0; i < size && err == nil; {
    var b byte

    if b, err = tga.r.ReadByte(); err != nil {
      break
    }

    count := uint(b)

    if count&(1<<7) != 0 {
      // encoded packet
      count &= ^uint(1 << 7)

      if err = tga.getPixel(tga.tmp[:]); err == nil {
        for count++; count > 0 && i < size; count-- {
          copy(out[i:], tga.tmp[:])
          i += 4
        }
      }
    } else {
      // raw packet
      for count++; count > 0 && i < size && err == nil; count-- {
        err = tga.getPixel(out[i:])
        i += 4
      }
    }
  }

  return
}

// flip flips pixels of image based on its origin.
func (tga *tga) flip(out []byte) {
  flipH := tga.raw.Flags&flagOriginRight != 0
  flipV := tga.raw.Flags&flagOriginTop == 0
  rowSize := tga.width * 4

  if flipH {
    for y := 0; y < tga.height; y++ {
      for x, offset := 0, y*rowSize; x < tga.width/2; x++ {
        a := out[offset+x*4:]
        b := out[offset+(tga.width-x-1)*4:]

        a[0], a[1], a[2], a[3], b[0], b[1], b[2], b[3] = b[0], b[1], b[2], b[3], a[0], a[1], a[2], a[3]
      }
    }
  }

  if flipV {
    for y := 0; y < tga.height/2; y++ {
      for x := 0; x < tga.width; x++ {
        a := out[y*rowSize+x*4:]
        b := out[(tga.height-y-1)*rowSize+x*4:]

        a[0], a[1], a[2], a[3], b[0], b[1], b[2], b[3] = b[0], b[1], b[2], b[3], a[0], a[1], a[2], a[3]
      }
    }
  }
}

// getHeader reads and validates TGA header.
func (tga *tga) getHeader() (err error) {
  if err = binary.Read(tga.r, binary.LittleEndian, &tga.raw); err != nil {
    return
  }

  if tga.raw.ImageType&imageTypeRLE != 0 {
    tga.decode = decodeRLE
  } else {
    tga.decode = decodeRaw
  }

  tga.raw.ImageType &= imageTypeMask
  flags := tga.raw.Flags & flagMask

  if flags != 0 && flags != 1 && flags != 8 {
    err = errors.New("invalid alpha size")
    return
  }

  tga.hasAlpha = ((flags != 0 || tga.raw.BPP == 32) ||
    (tga.raw.ImageType == imageTypeMonoChrome && tga.raw.BPP == 16) ||
    (tga.raw.ImageType == imageTypePaletted && tga.raw.PaletteBPP == 32))

  tga.width = int(tga.raw.Width)
  tga.height = int(tga.raw.Height)
  tga.pixelSize = int(tga.raw.BPP) >> 3

  // default is NOT premultiplied alpha model
  tga.ColorModel = color.NRGBAModel

  if err = tga.applyExtensions(); err != nil {
    return
  }

  var formatIsInvalid bool

  switch tga.raw.ImageType {
  case imageTypePaletted:
    formatIsInvalid = (tga.raw.PaletteType == 0 ||
      tga.raw.BPP != 8 ||
      tga.raw.PaletteFirst >= tga.raw.PaletteLength ||
      (tga.raw.PaletteBPP != 15 && tga.raw.PaletteBPP != 16 && tga.raw.PaletteBPP != 24 && tga.raw.PaletteBPP != 32))
    tga.isPaletted = true

  case imageTypeTrueColor:
    formatIsInvalid = (tga.raw.BPP != 32 &&
      tga.raw.BPP != 16 &&
      (tga.raw.BPP != 24 || tga.hasAlpha))

  case imageTypeMonoChrome:
    formatIsInvalid = ((tga.hasAlpha && tga.raw.BPP != 16) ||
      (!tga.hasAlpha && tga.raw.BPP != 8))

  default:
    err = errors.New("invalid or unsupported image type")
  }

  if err == nil && formatIsInvalid {
    err = errors.New("invalid image format")
  }

  return
}

func (tga *tga) getPixel(dst []byte) (err error) {
  var R, G, B, A uint8 = 0xff, 0xff, 0xff, 0xff
  src := tga.tmp

  if _, err = io.ReadFull(tga.r, src[0:tga.pixelSize]); err != nil {
    return
  }

  switch tga.pixelSize {
  case 4:
    if tga.hasAlpha {
      A = src[3]
    }
    fallthrough

  case 3:
    B, G, R = src[0], src[1], src[2]

  case 2:
    if tga.raw.ImageType == imageTypeMonoChrome {
      B, G, R = src[0], src[0], src[0]

      if tga.hasAlpha {
        A = src[1]
      }
    } else {
      word := uint16(src[0]) | (uint16(src[1]) << 8)
      B, G, R = wordToBGR(word)

      if tga.hasAlpha && (word&(1<<15)) == 0 {
        A = 0
      }
    }

  case 1:
    if tga.isPaletted {
      index := int(src[0])

      if int(index) >= tga.paletteLength {
        return errors.New("palette index out of range")
      }

      var m int

      if tga.raw.PaletteBPP == 24 {
        m = index * 3
        B, G, R = tga.palette[m+0], tga.palette[m+1], tga.palette[m+2]
      } else if tga.raw.PaletteBPP == 32 {
        m = index * 4
        B, G, R = tga.palette[m+0], tga.palette[m+1], tga.palette[m+2]

        if tga.hasAlpha {
          A = tga.palette[m+3]
        }
      } else if tga.raw.PaletteBPP == 16 {
        m = index * 2
        word := uint16(tga.palette[m+0]) | (uint16(tga.palette[m+1]) << 8)
        B, G, R = wordToBGR(word)
      }
    } else {
      B, G, R = src[0], src[0], src[0]
    }
  }

  dst[0], dst[1], dst[2], dst[3] = R, G, B, A

  return nil
}

// wordToBGR converts 15-bit color to BGR
func wordToBGR(word uint16) (B, G, R uint8) {
  B = uint8((word >> 0) & 31)
  B = uint8((B << 3) + (B >> 2))
  G = uint8((word >> 5) & 31)
  G = uint8((G << 3) + (G >> 2))
  R = uint8((word >> 10) & 31)
  R = uint8((R << 3) + (R >> 2))
  return
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// fill-column: 70
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
