package v0

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestNewTextFile(t *testing.T) {
	t.Parallel()

	file := NewTextFile("readme.md", "hello")
	content := MustValue(t, func() ([]byte, error) {
		defer file.Contents.Close()
		return io.ReadAll(file.Contents)
	})

	Then(t, "创建带文本内容的文件项",
		Expect(file.Name, Equal("readme.md")),
		Expect(string(content), Equal("hello")),
	)
}

func TestZipFileUpgrade(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	z := &ZipFile{
		FileName: "bundle.zip",
		Files: []File{
			NewTextFile("a.txt", "A"),
			NewTextFile("b.txt", "B"),
		},
	}

	Must(t, func() error {
		return z.Upgrade(recorder, httptest.NewRequest("GET", "/archive/zip", nil))
	})

	reader := MustValue(t, func() (*zip.Reader, error) {
		return zip.NewReader(bytes.NewReader(recorder.Body.Bytes()), int64(recorder.Body.Len()))
	})

	names := make([]string, 0, len(reader.File))
	contents := map[string]string{}
	for _, f := range reader.File {
		names = append(names, f.Name)
		rc := MustValue(t, f.Open)
		raw := MustValue(t, func() ([]byte, error) {
			defer rc.Close()
			return io.ReadAll(rc)
		})
		contents[f.Name] = string(raw)
	}

	Then(t, "写出 zip 响应头和压缩内容",
		Expect(recorder.Code, Equal(200)),
		Expect(recorder.Header().Get("Content-Type"), Equal("application/zip")),
		Expect(recorder.Header().Get("Content-Disposition"), Equal(`attachment; filename="bundle.zip"`)),
		Expect(names, Be(cmp.Len[[]string](2))),
		Expect(contents, Equal(map[string]string{
			"a.txt": "A",
			"b.txt": "B",
		})),
	)
}
