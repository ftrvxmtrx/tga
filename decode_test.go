package tga_test

import (
  "bufio"
  "image"
  "image/color"
  "os"
  "testing"
  _ "tga" // should be the first one, because TGA doesn't have any constant "header"

  _ "image/png"
)

type tgaTest struct {
  golden   string
  source   string
  maxDelta int
}

var tgaTests = []tgaTest{
  {"bw.png", "cbw8.tga", 0},
  {"bw.png", "ubw8.tga", 0},
  {"color.png", "ctc32.tga", 0},
  {"color.png", "ctc24.tga", 0},
  {"color.png", "ctc16.tga", 0},
  {"color.png", "ccm8.tga", 0},
  {"color.png", "ucm8.tga", 0},
  {"color.png", "utc32.tga", 0},
  {"color.png", "utc24.tga", 0},
  {"color.png", "utc16.tga", 0},
  {"monochrome16.png", "monochrome16_top_left_rle.tga", 0},
  {"monochrome16.png", "monochrome16_top_left.tga", 0},
  {"monochrome8.png", "monochrome8_bottom_left_rle.tga", 0},
  {"monochrome8.png", "monochrome8_bottom_left.tga", 0},
  {"rgb24.png", "rgb24_bottom_left_rle.tga", 2<<8 + 2},
  {"rgb24.png", "rgb24_top_left_colormap.tga", 0},
  {"rgb24.png", "rgb24_top_left.tga", 2<<8 + 2},
  {"rgb32.0.png", "rgb32_bottom_left.tga", 0},
  {"rgb32.1.png", "rgb32_top_left_rle_colormap.tga", 0},
  {"rgb32.0.png", "rgb32_top_left_rle.tga", 0},
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

func equal(c0, c1 color.Color, maxDelta int) bool {
  r0, g0, b0, a0 := c0.RGBA()
  r1, g1, b1, a1 := c1.RGBA()

  if a0 == 0 && a1 == 0 {
    return true
  }

  d0 := delta(r0, r1)
  d1 := delta(g0, g1)
  d2 := delta(b0, b1)
  d3 := delta(a0, a1)

  return d0 <= maxDelta && d1 <= maxDelta && d2 <= maxDelta && d3 <= maxDelta
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
          if !equal(golden.At(x, y), source.At(x, y), test.maxDelta) {
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
