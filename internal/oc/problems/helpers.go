package problems

import (
	"archive/zip"
	"fmt"
	"io"
)

func unzip(data io.ReaderAt, size int64) error {
    r, err := zip.NewReader(data, size)
    if err != nil {
        return err
    }

    for _, f := range r.File {
        fmt.Println(f.Name)
    }

    return nil
}