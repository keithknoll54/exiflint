# exiflint

A JPEG carries its metadata (camera make, date taken, orientation, and so
on) inside an APP1 segment whose payload is a small TIFF file glued onto the
front of the image. Most tools that read this either trust it blindly or
paper over corruption with recovered/best-guess values. Files downloaded
from the wild, re-saved by random editors, or hand-crafted for testing
regularly have offsets that point past the end of the buffer, IFD entry
counts that don't match the remaining bytes, or type/length combinations
that don't make sense. A parser that isn't checking those things is one
`panic: runtime error: index out of range` away from a bad day.

exiflint does two things:

- parses the EXIF block with bounds and structure checks at every offset,
  failing with a specific error instead of reading out of range
- pretty-prints whatever it found in a plain, aligned list

## Usage

```
go build -o exiflint .
./exiflint photo.jpg
```

Example output:

```
photo.jpg
  Make            Canon
  Model           Canon EOS 80D
  Orientation     1
  DateTime        2024:03:11 14:22:07
  XResolution     72/1
  YResolution     72/1
  ResolutionUnit  2
  Software        GIMP 2.10
```

On a malformed file:

```
$ ./exiflint broken.jpg
broken.jpg: invalid EXIF data: tag Model: value offset 91742 out of range
```

## Current scope

Only IFD0 is read right now - the handful of tags most JPEGs set directly
(Make, Model, Orientation, DateTime, resolution, Software, Artist,
Copyright). The EXIF sub-IFD (exposure time, ISO, lens info) and the GPS
IFD are not walked yet; unknown tags are skipped rather than guessed at.

## Layout

- `jpeg.go` - walks JPEG marker segments to find the APP1 Exif payload
- `exif.go` - validates the TIFF header and decodes IFD0 entries
- `print.go` - formats decoded tags for output
- `main.go` - CLI glue
