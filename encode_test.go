package tga_test

import (
  "github.com/ftrvxmtrx/tga"
  "image"
  "image/draw"
  "os"
  "testing"
)

const dstFilename = "encode_test.tga"

func encode(m image.Image) (err error) {
  var f *os.File

  if f, err = os.Create("testdata/" + dstFilename); err != nil {
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
  defer os.Remove("testdata/" + dstFilename)

loop:

  for _, test := range tgaTests {
    typ := "normal"

    if m, _, err := decode(test.source); err != nil {
      t.Errorf("%s: %v", test.source, err)
    } else if err = encode(m); err != nil {
      t.Errorf("%s: encode failed (%v)", test.source, err)
    } else if second, _, err := decode(dstFilename); err != nil {
      t.Errorf("%s: %v", dstFilename, err)
    } else if testImagesEqual(test.source, m, second, t) {
      // test monochrome
      b := second.Bounds()
      gray := image.NewGray(b)
      draw.Draw(gray, b, second, b.Min, draw.Src)
      typ = "gray"

      if err = encode(gray); err != nil {
        t.Errorf("%s: encode to gray failed (%v)", test.source, err)
      } else if third, _, err := decode(dstFilename); err != nil {
        t.Errorf("%s: gray decode failed (%v)", dstFilename, err)
      } else if testImagesEqual(test.source, gray, third, t) {
        premultiplied := image.NewRGBA(b)
        draw.Draw(premultiplied, b, second, b.Min, draw.Src)
        typ = "premultiplied"

        if err = encode(premultiplied); err != nil {
          t.Errorf("%s: encode to premultiplied failed (%v)", test.source, err)
        } else if fourth, _, err := decode(dstFilename); err != nil {
          t.Errorf("%s: premultiplied decode failed (%v)", dstFilename, err)
        } else if testImagesEqual(test.source, premultiplied, fourth, t) {
          continue loop
        } else {
          t.Errorf("%T %T", premultiplied, fourth)
        }
      }
    }

    t.Errorf("%s: encoded %s image (%s) is different", test.source, typ, dstFilename)
  }
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// fill-column: 70
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
