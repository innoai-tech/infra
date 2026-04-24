package archive

import (
	"context"

	apiv0 "github.com/innoai-tech/infra/internal/example/pkg/apis/archive/v0"
)

type Service struct{}

func (Service) ArchiveZip(ctx context.Context) (*apiv0.ZipFile, error) {
	return &apiv0.ZipFile{
		FileName: "x.zip",
		Files: []apiv0.File{
			apiv0.NewTextFile("readme.md", "123"),
			apiv0.NewTextFile("1.txt", "1"),
		},
	}, nil
}
