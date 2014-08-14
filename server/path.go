package server

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var NoFileRegisteredErr = fmt.Errorf("No File Registered")
var FileNotFound = fmt.Errorf("File not found")
var FileToLarge = fmt.Errorf("File to large")

func (flux *FluxServer) getInMemReader() (TftpReader, error) {
	flux.RLock()
	defer flux.RUnlock()
	if flux.mem == nil {
		return nil, NoFileRegisteredErr
	}
	return bytes.NewBuffer(flux.mem.Bytes()), nil
}

const MAX_SIZE int64 = 1000 * 1024 * 1024

func (flux *FluxServer) getFileReader(filename string) (TftpReader, error) {
	searchPath := flux.Conf.SearchPath
	info, err := os.Stat(searchPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("Search path is not a directory")
	}

	// sanitize filename
	filename = strings.Replace(filename, "/", "", -1)

	file, err := os.Open(filepath.Join(searchPath, filename))
	if err != nil {
		return nil, FileNotFound
	}
	defer file.Close()
	fInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fInfo.Size() >= MAX_SIZE {
		return nil, FileToLarge
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// This function does the actual magic. We look for the asked file in the
// search path and if it doesn't exist we just return a reader to the most
// recent file added
func (flux *FluxServer) GetFile(path string) (TftpReader, error) {
	if flux.Conf.SearchPath == "" {
		return flux.getInMemReader()
	}

	// look if file
	reader, err := flux.getFileReader(path)
	if err == FileNotFound {
		return flux.getInMemReader()
	} else if err != nil {
		return nil, err
	}

	return reader, nil
}
