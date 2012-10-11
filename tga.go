package tga

type rawHeader struct {
  IdLength      uint8
  PaletteType   uint8
  ImageType     uint8
  PaletteFirst  uint16
  PaletteLength uint16
  PaletteBPP    uint8
  OriginX       uint16
  OriginY       uint16
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

const (
  flagOriginRight   = 1 << 4
  flagOriginTop     = 1 << 5
  flagAlphaSizeMask = 0x0f
)

const (
  imageTypePaletted   = 1
  imageTypeTrueColor  = 2
  imageTypeMonoChrome = 3
  imageTypeMask       = 3
  imageTypeFlagRLE    = 1 << 3
)

const (
  tgaRawHeaderSize = 18
  tgaRawFooterSize = 26
)

const (
  extAreaAttrTypeOffset = 0x1ee
)

const (
  attrTypeAlpha              = 3
  attrTypePremultipliedAlpha = 4
)

var tgaSignature = []byte("TRUEVISION-XFILE.\x00")

func newFooter() *rawFooter {
  f := &rawFooter{}
  copy(f.Signature[:], tgaSignature)
  return f
}

func newExtArea(attrType byte) []byte {
  area := make([]byte, extAreaAttrTypeOffset+1)
  area[0], area[1] = 0xef, 0x01 // size
  area[extAreaAttrTypeOffset] = attrType
  return area
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// fill-column: 70
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
