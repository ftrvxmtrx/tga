package tga

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

const (
  extAreaAttrTypeOffset = 0x1ee
)

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// fill-column: 70
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
