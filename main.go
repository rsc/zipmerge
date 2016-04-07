// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Zipmerge merges the content of many zip files,
// without decompressing and recompressing the data.
//
// Usage:
//
//	zipmerge [-o out.zip] a.zip b.zip ...
//
// By default, zipmerge appends the content of the second and subsequent zip files
// to the first, rewriting the first in place.
// If the -o option is given, zipmerge creates a new output file containing
// the content of all the zip files, without modifying any of the source zip files.
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"rsc.io/zipmerge/internal/zip"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: zipmerge [-o dst.zip] a.zip [b.zip...]\n")
	os.Exit(2)
}

var outputFile = flag.String("o", "", "write to `file`")

func main() {
	log.SetPrefix("zipmerge: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	var f *os.File
	var w *zip.Writer

	if *outputFile != "" {
		var err error
		f, err = os.Create(*outputFile)
		if err != nil {
			log.Fatal(err)
		}
		w = zip.NewWriter(f)
	} else {
		var err error
		f, err = os.OpenFile(args[0], os.O_RDWR, 0)
		if err != nil {
			log.Fatal(err)
		}
		size, err := f.Seek(0, 2)
		if err != nil {
			log.Fatal(err)
		}
		r, err := zip.NewReader(f, size)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.Seek(r.AppendOffset(), 0); err != nil {
			log.Fatal(err)
		}
		w = r.Append(f)
		args = args[1:]
	}

	for _, name := range args {
		rc, err := zip.OpenReader(name)
		if err != nil {
			log.Print(err)
			continue
		}
		for _, file := range rc.File {
			if err := w.Copy(file); err != nil {
				log.Printf("copying from %s (%s): %v", name, file.Name, err)
			}
		}
	}

	if err := w.Close(); err != nil {
		log.Fatal("finishing zip file: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Fatal("finishing zip file: %v", err)
	}
}
