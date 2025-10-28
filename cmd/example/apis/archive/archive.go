package archive

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/octohelm/courier/pkg/courierhttp"
)

type ArchiveZip struct {
	courierhttp.MethodGet `path:"/archive/zip"`
}

func (a ArchiveZip) Output(ctx context.Context) (any, error) {
	return &ZipFile{
		FileName: "x.zip",
		Files: []file{
			{Name: "readme.md", Contents: io.NopCloser(bytes.NewBufferString("123"))},
			{Name: "1.txt", Contents: io.NopCloser(bytes.NewBufferString("1"))},
		},
	}, nil
}

type file struct {
	Name     string
	Contents io.ReadCloser
}

type ZipFile struct {
	FileName string
	Files    []file
}

func (z *ZipFile) Upgrade(w http.ResponseWriter, r *http.Request) (err error) {
	h := w.Header()
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, z.FileName))
	w.WriteHeader(200)

	zw := zip.NewWriter(w)
	defer zw.Close()

	writeFile := func(f *file) error {
		defer f.Contents.Close()

		// already failed, should skip
		// but must close reader to avoid oom
		if err != nil {
			return nil
		}

		zff, err := zw.Create(f.Name)
		if err != nil {
			return err
		}
		if _, err = io.Copy(zff, f.Contents); err != nil {
			return err
		}
		return nil
	}

	for i := range z.Files {
		if e := writeFile(&z.Files[i]); e != nil {
			err = e
			continue
		}
	}

	return err
}
