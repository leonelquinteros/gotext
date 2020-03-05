package parser

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ParseDirFunc parses one directory
type ParseDirFunc func(filePath, basePath string, data *DomainMap) error

var knownParser []ParseDirFunc

// AddParser to the known parser list
func AddParser(parser ParseDirFunc) {
	if knownParser == nil {
		knownParser = []ParseDirFunc{parser}
	} else {
		knownParser = append(knownParser, parser)
	}
}

// ParseDir calls all known parser for each directory
func ParseDir(dirPath, basePath string, data *DomainMap) error {
	dirPath, _ = filepath.Abs(dirPath)
	basePath, _ = filepath.Abs(basePath)

	for _, parser := range knownParser {
		err := parser(dirPath, basePath, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseDirRec calls all known parser for each directory
func ParseDirRec(dirPath string, exclude []string, data *DomainMap, verbose bool) error {
	dirPath, _ = filepath.Abs(dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// skip directory if in exclude list
			subDir, _ := filepath.Rel(dirPath, path)
			for _, d := range exclude {
				if strings.HasPrefix(subDir, d) {
					return nil
				}
			}
			if verbose {
				log.Print(path)
			}

			err := ParseDir(path, dirPath, data)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
