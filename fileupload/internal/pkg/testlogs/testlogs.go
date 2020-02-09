package testlogs

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type ZipTestoutput struct {
	ZipWriter *zip.Writer
}

func (zto *ZipTestoutput) ZipLogFiles() {

	_ = filepath.Walk("/test-output", func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		if err != nil {
			return err
		}

		fileInfoHeader, _ := zip.FileInfoHeader(info)
		writer, err := zto.ZipWriter.CreateHeader(fileInfoHeader)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
		return nil
	})

}

func handleZip(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("main.go")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// write straight to the http.ResponseWriter
	zw := zip.NewWriter(w)
	cf, err := zw.Create(f.Name())
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", f.Name()))

	// copy the file contents to the zip Writer
	_, err = io.Copy(cf, f)
	if err != nil {
		log.Fatal(err)
	}

	// close the zip Writer to flush the contents to the ResponseWriter
	err = zw.Close()
	if err != nil {
		log.Fatal(err)
	}
}
