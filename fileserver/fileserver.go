// Package fileserver implements a gemini handler function to serve files with optional auto
// indexing for directory listings.
package fileserver

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"mime"
	"path"
	"strings"

	"github.com/karelbilek/website/gemini"
)

var (
	ErrDirWithoutIndexFile = errors.New("path without index.gmi not allowed")
	ErrUnsupportedFileType = errors.New("disabled/unsupported file type")
)

func Serve(root fs.FS) func(w gemini.ResponseWriter, r *gemini.Request) {
	return func(w gemini.ResponseWriter, r *gemini.Request) {
		fl, fullpath, redirTo, err := fullPath(root, r.URL.Path)
		if err != nil {
			log.Println("error 1 - ", r.URL.Path, err)
			w.WriteHeader(gemini.StatusNotFound, "oopsie woopsie!! UwU")
			return
		}

		if redirTo != "" {
			w.WriteHeader(gemini.StatusRedirectPermanent, redirTo)
			return
		}

		body, mimeType, err := readFile(fullpath, fl)
		if err != nil {
			log.Println("error 2 -  ", r.URL.Path, fullpath, err)

			w.WriteHeader(gemini.StatusNotFound, "oopsie woopsie!! UwU")
			return
		}

		w.WriteHeader(gemini.StatusSuccess, mimeType)
		w.Write(body)
	}
}

func fullPath(root fs.FS, requestPath string) (fs.File, string, string, error) {
	fPath := strings.TrimPrefix(requestPath, "/")
	fPath = strings.TrimSuffix(fPath, "/")
	if fPath == "" {
		fPath = gemini.IndexFile
	}
	pathInfo, err := root.Open(fPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("path: %w", err)
	}

	stat, err := pathInfo.Stat()
	if err != nil {
		return nil, "", "", fmt.Errorf("path stat: %w", err)
	}

	if stat.IsDir() {
		if !strings.HasSuffix(requestPath, "/") {
			return nil, "", requestPath + "/", nil
		}

		log.Println("it is dir")
		//fmt.Println(requestPath)
		subDirIndex := path.Join(fPath, gemini.IndexFile)
		subPathInfo, err := root.Open(subDirIndex)
		log.Println("spi", subDirIndex)

		if err != nil {
			log.Println("error getting")

			return nil, "", "", ErrDirWithoutIndexFile
		}
		log.Println("got index")

		return subPathInfo, subDirIndex, "", nil
	}

	return pathInfo, fPath, "", nil
}

func readFile(fpath string, file fs.File) ([]byte, string, error) {
	mimeType := getMimeType(fpath)
	if mimeType == "" {
		return nil, "", ErrUnsupportedFileType
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("read: %w", err)
	}
	return data, mimeType, nil
}

func getMimeType(fpath string) string {
	if ext := path.Ext(fpath); ext != ".gmi" {
		return mime.TypeByExtension(ext)
	}
	if strings.HasSuffix(fpath, ".cs.gmi") {

	}
	return gemini.MimeType
}
