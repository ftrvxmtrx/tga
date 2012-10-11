package tga_test

import (
  "bufio"
  _ "github.com/ftrvxmtrx/tga" // should be the first one, because TGA doesn't have any constant "header"
  "image"
  "image/color"
  "os"
  "testing"

  _ "image/png"
)

type tgaTest struct {
  golden string
  source string
}

var tgaTests = []tgaTest{
  {"bw.png", "cbw8.tga"},
  {"bw.png", "ubw8.tga"},
  {"color.png", "ctc32.tga"},
  {"color.png", "ctc24.tga"},
  {"color.png", "ctc16.tga"},
  {"color.png", "ccm8.tga"},
  {"color.png", "ucm8.tga"},
  {"color.png", "utc32.tga"},
  {"color.png", "utc24.tga"},
  {"color.png", "utc16.tga"},
  {"monochrome16.png", "monochrome16_top_left_rle.tga"},
  {"monochrome16.png", "monochrome16_top_left.tga"},
  {"monochrome8.png", "monochrome8_bottom_left_rle.tga"},
  {"monochrome8.png", "monochrome8_bottom_left.tga"},
  {"rgb24.0.png", "rgb24_bottom_left_rle.tga"},
  {"rgb24.1.png", "rgb24_top_left_colormap.tga"},
  {"rgb24.0.png", "rgb24_top_left.tga"},
  {"rgb32.0.png", "rgb32_bottom_left.tga"},
  {"rgb32.1.png", "rgb32_top_left_rle_colormap.tga"},
  {"rgb32.0.png", "rgb32_top_left_rle.tga"},
}

func decode(filename string) (image.Image, string, error) {
  f, err := os.Open("testdata/" + filename)

  if err != nil {
    return nil, "", err
  }

  defer f.Close()

  return image.Decode(bufio.NewReader(f))
}

func delta(a, b uint32) int {
  if a < b {
    return int(b) - int(a)
  }

  return int(a) - int(b)
}

func equal(c0, c1 color.Color) bool {
  r0, g0, b0, a0 := c0.RGBA()
  r1, g1, b1, a1 := c1.RGBA()

  if a0 == 0 && a1 == 0 {
    return true
  }

  return r0 == r1 && g0 == g1 && b0 == b1 && a0 == a1
}

func TestDecode(t *testing.T) {
loop:

  for _, test := range tgaTests {
    if golden, _, err := decode(test.golden); err != nil {
      t.Errorf("%s: %v", test.golden, err)
    } else if source, format, err := decode(test.source); err != nil {
      t.Errorf("%s: %v", test.source, err)
    } else if format != "tga" {
      t.Errorf("%s: expected tga, got %v", test.source, format)
    } else if !golden.Bounds().Eq(source.Bounds()) {
      t.Errorf("%s: expected bounds %v, got %v", test.source, golden.Bounds(), source.Bounds())
    } else {
      gb := golden.Bounds()

      for x := gb.Min.X; x < gb.Max.X; x++ {
        for y := gb.Min.Y; y < gb.Max.Y; y++ {
          if !equal(golden.At(x, y), source.At(x, y)) {
            t.Errorf("%s: (%d, %d) -- expected %v, got %v", test.source, x, y, golden.At(x, y), source.At(x, y))
            continue loop
          }
        }
      }
    }
  }
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// fill-column: 70
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
