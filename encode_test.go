package tga_test

import (
  "github.com/ftrvxmtrx/tga"
  "image"
  "image/draw"
  "os"
  "testing"
)

func encode(m image.Image) (filename string, err error) {
  filename = "encode_test.tga"
  var f *os.File

  if f, err = os.Create("testdata/" + filename); err != nil {
    return
  }

  defer f.Close()
  err = tga.Encode(f, m)

  return
}

func testImagesEqual(filename string, expected, got image.Image, t *testing.T) bool {
  gb := expected.Bounds()

  for x := gb.Min.X; x < gb.Max.X; x++ {
    for y := gb.Min.Y; y < gb.Max.Y; y++ {
      if !equal(expected.At(x, y), got.At(x, y)) {
        t.Errorf("%s: (%d, %d) -- expected %v, got %v", filename, x, y, expected.At(x, y), got.At(x, y))
        return false
      }
    }
  }

  return true
}

func TestEncode(t *testing.T) {
  var dst string

loop:

  for _, test := range tgaTests {
    typ := "normal"

    if source, _, err := decode(test.source); err != nil {
      t.Errorf("%s: %v", test.source, err)
    } else if dst, err = encode(source); err != nil {
      t.Errorf("%s: encode failed (%v)", test.source, err)
    } else if second, _, err := decode(dst); err != nil {
      t.Errorf("%s: %v", dst, err)
    } else if testImagesEqual(test.source, source, second, t) {
      // test monochrome
      b := second.Bounds()
      gray := image.NewGray(b)
      draw.Draw(gray, b, second, b.Min, draw.Src)
      typ = "gray"

      if dst, err = encode(gray); err != nil {
        t.Errorf("%s: encode to gray failed (%v)", test.source, err)
      } else if third, _, err := decode(dst); err != nil {
        t.Errorf("%s: gray decode failed (%v)", dst, err)
      } else if testImagesEqual(test.source, gray, third, t) {
        premultiplied := image.NewRGBA(b)
        draw.Draw(premultiplied, b, second, b.Min, draw.Src)
        typ = "premultiplied"

        if dst, err = encode(premultiplied); err != nil {
          t.Errorf("%s: encode to premultiplied failed (%v)", test.source, err)
        } else if fourth, _, err := decode(dst); err != nil {
          t.Errorf("%s: premultiplied decode failed (%v)", dst, err)
        } else if testImagesEqual(test.source, premultiplied, fourth, t) {
          continue loop
        } else {
          t.Errorf("%T %T", premultiplied, fourth)
        }
      }
    }

    t.Errorf("%s: encoded %s image (%s) is different", test.source, typ, dst)
  }

  os.Remove("testdata/" + dst)
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// fill-column: 70
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
