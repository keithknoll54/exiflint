package main

import "fmt"

// JPEG segment markers we care about. Everything else is skipped by length.
const (
	markerSOI  = 0xD8
	markerEOI  = 0xD9
	markerSOS  = 0xDA // start of scan: metadata segments never appear after this
	markerAPP1 = 0xE1
)

// findEXIF walks the JPEG marker segments and returns the raw TIFF-structured
// payload of the first Exif APP1 segment it finds. It validates the segment
// framing as it goes so that a truncated or corrupt file is rejected here
// rather than causing an out-of-range read later in the TIFF parser.
func findEXIF(data []byte) ([]byte, error) {
	if len(data) < 4 || data[0] != 0xFF || data[1] != markerSOI {
		return nil, fmt.Errorf("missing SOI marker")
	}

	pos := 2
	for pos+2 <= len(data) {
		if data[pos] != 0xFF {
			return nil, fmt.Errorf("expected marker at offset %d, got 0x%02X", pos, data[pos])
		}
		marker := data[pos+1]
		pos += 2

		if marker == markerEOI || marker == markerSOS {
			break
		}
		// Standalone markers (RST0-RST7 and a couple of others) carry no
		// length field and no payload.
		if marker >= 0xD0 && marker <= 0xD7 {
			continue
		}

		if pos+2 > len(data) {
			return nil, fmt.Errorf("truncated segment header at offset %d", pos)
		}
		segLen := int(data[pos])<<8 | int(data[pos+1])
		if segLen < 2 || pos+segLen > len(data) {
			return nil, fmt.Errorf("invalid segment length at offset %d", pos)
		}
		payload := data[pos+2 : pos+segLen]

		if marker == markerAPP1 && len(payload) > 6 && string(payload[:6]) == "Exif\x00\x00" {
			return payload[6:], nil
		}

		pos += segLen
	}

	return nil, fmt.Errorf("no EXIF (APP1) segment found")
}
