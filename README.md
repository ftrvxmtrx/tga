# tga

tga is a [Go](http://golang.org/) package for decoding and encoding TARGA image
format.

It supports RLE and raw TARGA images with 8/15/16/24/32 bits per pixel,
monochrome, truecolor and colormapped images. It also correctly handles origins,
attribute type in extensions area and successfully passes TGA 2.0 conformance
suite (http://googlesites.inequation.org/tgautilities).

Encoding an image doesn't involve conversion if it's `image.Gray`, `image.RGBA`
or `image.NRGBA`. Other types are converted to `image.NRGBA` prior to encoding.

## Build status

<a href="http://goci.me/project/github.com/ftrvxmtrx/tga">
<img src="http://goci.me/project/image/github.com/ftrvxmtrx/tga" />
</a>

## Installation

    $ go get github.com/ftrvxmtrx/tga

## Documentation and examples

[tga on go.pkgdoc.org](http://go.pkgdoc.org/github.com/ftrvxmtrx/tga)

## License

Code is licensed under the MIT license (see `LICENSE.MIT`).

Several sample image files in `testdata` directory are copyright to TrueVision,
Inc. and are freely available, free of charge and under no licensing terms at
http://googlesites.inequation.org/tgautilities

These sample images (and those which were converted from them) are:
```
bw.png
cbw8.tga
ccm8.tga
color.png
ctc16.tga
ctc24.tga
ctc32.tga
ubw8.tga
ucm8.tga
utc16.tga
utc24.tga
utc32.tga
```
