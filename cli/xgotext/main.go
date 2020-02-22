package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

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

	data, err := parser.ParseDirRec(*dirName)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(*outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create output dir: %s", err)
	}

	for name, domain := range data {
		outFile := filepath.Join(*outputDir, name+".po")
		file, err := os.Create(outFile)
		if err != nil {
			log.Fatalf("failed to save po file for %s: %s", name, err)
		}

		file.WriteString(`msgid ""
msgstr ""
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: \n"
"X-Generator: xgotext\n"

`)
		file.WriteString(domain.Dump())
		file.Close()
	}
}
