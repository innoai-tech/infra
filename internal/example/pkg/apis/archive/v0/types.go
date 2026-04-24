package v0

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// File 表示一个待写入压缩包的文件。
type File struct {
	Name     string
	Contents io.ReadCloser
}

// ZipFile 表示一个可升级为 zip 响应的文件集。
type ZipFile struct {
	FileName string
	Files    []File
}

// NewTextFile 使用文本内容创建一个 zip 文件项。
func NewTextFile(name string, content string) File {
	return File{
		Name:     name,
		Contents: io.NopCloser(bytes.NewBufferString(content)),
	}
}

// Upgrade 将 ZipFile 写入 HTTP 响应。
func (z *ZipFile) Upgrade(w http.ResponseWriter, r *http.Request) (err error) {
	h := w.Header()
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, z.FileName))
	w.WriteHeader(http.StatusOK)

	zw := zip.NewWriter(w)
	defer zw.Close()

	writeFile := func(f *File) error {
		defer f.Contents.Close()

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
