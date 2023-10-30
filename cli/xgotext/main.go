package main

import (
	"flag"
	"log"
	"strings"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
	"github.com/leonelquinteros/gotext/cli/xgotext/parser/dir"
	pkg_tree "github.com/leonelquinteros/gotext/cli/xgotext/parser/pkg-tree"
)

var (
	pkgTree       = flag.String("pkg-tree", "", "Main path: /path/to/go/pkg")
	dirName       = flag.String("in", "", "Input dir: /path/to/go/pkg")
	outputDir     = flag.String("out", "", "Output dir: /path/to/i18n/files")
	defaultDomain = flag.String("default", "default", "Name of default domain")
	excludeDirs   = flag.String("exclude", ".git", "Comma separated list of directories to exclude")
	noLocations   = flag.Bool("no-locations", false, "Do not write translation locations")
	verbose       = flag.Bool("v", false, "Print currently handled directory")
)

func main() {
	flag.Parse()

	// Init logger
	log.SetFlags(0)

	if *pkgTree == "" && *dirName == "" {
		log.Fatal("No input directory given")
	}
	if *pkgTree != "" && *dirName != "" {
		log.Fatal("Specify either main or in")
	}
	if *outputDir == "" {
		log.Fatal("No output directory given")
	}

	data := &parser.DomainMap{
		Default: *defaultDomain,
	}

	if *pkgTree != "" {
		err := pkg_tree.ParsePkgTree(*pkgTree, data, *verbose)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := dir.ParseDirRec(*dirName, strings.Split(*excludeDirs, ","), data, *verbose)
		if err != nil {
			log.Fatal(err)
		}
	}

	err := data.Save(*outputDir, !*noLocations)
	if err != nil {
		log.Fatal(err)
	}
}
