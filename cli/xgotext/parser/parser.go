package parser

import (
	"os"
	"path/filepath"
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
func ParseDirRec(dirPath string, data *DomainMap) error {
	dirPath, _ = filepath.Abs(dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			err := ParseDir(path, dirPath, data)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
