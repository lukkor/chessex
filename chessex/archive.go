package chessex

import (
	"bufio"
	"compress/bzip2"
	"io"
	"os"
	"regexp"
)

type Archive struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewArchive(archive string) (*Archive, error) {
	// Open the archive
	file, err := os.Open(archive)
	if err != nil {
		return nil, err
	}

	// Create an unzip stream reader
	r := bzip2.NewReader(file)

	// Create the custom scanner
	s := bufio.NewScanner(r)
	s.Split(splitFunc)

	return &Archive{
		file:    file,
		scanner: s,
	}, nil
}

func (a *Archive) Scan() bool {
	return a.scanner.Scan()
}

func (a *Archive) Text() string {
	return a.scanner.Text()
}

func (a *Archive) Close() {
	a.file.Close()
}

func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	endOfGame := regexp.MustCompile(`[\r\n]{2}\[`)
	if len(data) == 0 {
		return 0, nil, nil
	}

	if loc := endOfGame.FindIndex(data); loc != nil && loc[0] >= 0 {
		return loc[1] - 1, data[0:loc[0]], nil
	}

	if atEOF {
		return len(data), data, io.EOF
	}

	return 0, nil, nil

}
