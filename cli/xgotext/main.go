package main

import (
	"flag"
	"log"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

var (
	dirName   = flag.String("in", "", "input dir: /path/to/go/pkg")
	outputDir = flag.String("out", "", "output dir: /path/to/i18n/files")
)

func main() {
	flag.Parse()

	// Init logger
	log.SetFlags(0)

	if *dirName == "" {
		log.Fatal("No input directory given")
	}
	if *outputDir == "" {
		log.Fatal("No output directory given")
	}

	data, err := parser.ParseDirRec(*dirName)
	if err != nil {
		log.Fatal(err)
	}

	err = data.Save(*outputDir)
	if err != nil {
		log.Fatal(err)
	}
}
