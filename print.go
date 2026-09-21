package main

import (
	"fmt"
	"io"
)

// printTags renders decoded tags as a simple aligned list. Kept separate
// from the parser so the two can be tested and changed independently.
func printTags(w io.Writer, path string, tags []Tag) {
	fmt.Fprintln(w, path)

	if len(tags) == 0 {
		fmt.Fprintln(w, "  (no recognized EXIF tags)")
		return
	}

	width := 0
	for _, t := range tags {
		if len(t.Name) > width {
			width = len(t.Name)
		}
	}

	for _, t := range tags {
		fmt.Fprintf(w, "  %-*s  %s\n", width, t.Name, t.Value)
	}
}
