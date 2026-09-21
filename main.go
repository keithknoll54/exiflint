package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: exiflint <path-to-jpeg>")
		os.Exit(2)
	}

	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}

	exifData, err := findEXIF(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: invalid JPEG: %v\n", path, err)
		os.Exit(1)
	}

	tags, err := parseEXIF(exifData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: invalid EXIF data: %v\n", path, err)
		os.Exit(1)
	}

	printTags(os.Stdout, path, tags)
}
