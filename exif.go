package main

import (
	"encoding/binary"
	"fmt"
)

// Tag is one decoded IFD entry, already formatted for display.
type Tag struct {
	Name  string
	Type  string
	Value string
}

// Only IFD0's most common tags are named for now; anything else is skipped
// rather than guessed at. GPS and EXIF sub-IFDs are not walked yet.
var tagNames = map[uint16]string{
	0x010F: "Make",
	0x0110: "Model",
	0x0112: "Orientation",
	0x011A: "XResolution",
	0x011B: "YResolution",
	0x0128: "ResolutionUnit",
	0x0131: "Software",
	0x0132: "DateTime",
	0x013B: "Artist",
	0x8298: "Copyright",
}

const (
	typeByte     = 1
	typeASCII    = 2
	typeShort    = 3
	typeLong     = 4
	typeRational = 5
)

var typeSizes = map[uint16]int{
	typeByte:     1,
	typeASCII:    1,
	typeShort:    2,
	typeLong:     4,
	typeRational: 8,
}

// parseEXIF validates a TIFF header and reads IFD0. Any offset or length
// that would run past the end of data is treated as a parse error, not a
// panic - this data usually arrives from files of unknown provenance.
func parseEXIF(data []byte) ([]Tag, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("TIFF header too short")
	}

	var order binary.ByteOrder
	switch string(data[0:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return nil, fmt.Errorf("unrecognized byte-order mark %q", data[0:2])
	}

	if order.Uint16(data[2:4]) != 42 {
		return nil, fmt.Errorf("bad TIFF magic number")
	}

	ifdOffset := int(order.Uint32(data[4:8]))
	return readIFD(data, order, ifdOffset)
}

func readIFD(data []byte, order binary.ByteOrder, offset int) ([]Tag, error) {
	if offset < 0 || offset+2 > len(data) {
		return nil, fmt.Errorf("IFD offset %d out of range", offset)
	}
	count := int(order.Uint16(data[offset : offset+2]))
	entriesStart := offset + 2
	entriesEnd := entriesStart + count*12
	if entriesEnd > len(data) {
		return nil, fmt.Errorf("IFD claims %d entries but data is too short", count)
	}

	var tags []Tag
	for i := 0; i < count; i++ {
		entry := data[entriesStart+i*12 : entriesStart+(i+1)*12]
		tagID := order.Uint16(entry[0:2])
		fieldType := order.Uint16(entry[2:4])
		fieldCount := order.Uint32(entry[4:8])
		valueBytes := entry[8:12]

		name, known := tagNames[tagID]
		if !known {
			continue
		}

		size, ok := typeSizes[fieldType]
		if !ok {
			return nil, fmt.Errorf("tag %s: unsupported field type %d", name, fieldType)
		}
		total := size * int(fieldCount)
		if total <= 0 {
			return nil, fmt.Errorf("tag %s: zero-length value", name)
		}

		var raw []byte
		if total <= 4 {
			raw = valueBytes[:total]
		} else {
			off := int(order.Uint32(valueBytes))
			if off < 0 || off+total > len(data) {
				return nil, fmt.Errorf("tag %s: value offset %d out of range", name, off)
			}
			raw = data[off : off+total]
		}

		value, typeName, err := decodeValue(fieldType, raw, order)
		if err != nil {
			return nil, fmt.Errorf("tag %s: %w", name, err)
		}
		tags = append(tags, Tag{Name: name, Type: typeName, Value: value})
	}
	return tags, nil
}

func decodeValue(fieldType uint16, raw []byte, order binary.ByteOrder) (string, string, error) {
	switch fieldType {
	case typeASCII:
		s := string(raw)
		for i, c := range s {
			if c == 0 {
				s = s[:i]
				break
			}
		}
		return s, "ASCII", nil
	case typeByte:
		return fmt.Sprintf("%d", raw[0]), "BYTE", nil
	case typeShort:
		return fmt.Sprintf("%d", order.Uint16(raw[:2])), "SHORT", nil
	case typeLong:
		return fmt.Sprintf("%d", order.Uint32(raw[:4])), "LONG", nil
	case typeRational:
		if len(raw) < 8 {
			return "", "", fmt.Errorf("truncated RATIONAL value")
		}
		num := order.Uint32(raw[0:4])
		den := order.Uint32(raw[4:8])
		if den == 0 {
			return "", "", fmt.Errorf("RATIONAL with zero denominator")
		}
		return fmt.Sprintf("%d/%d", num, den), "RATIONAL", nil
	default:
		return "", "", fmt.Errorf("unsupported field type %d", fieldType)
	}
}
