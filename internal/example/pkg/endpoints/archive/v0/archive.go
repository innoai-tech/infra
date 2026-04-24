package v0

import "github.com/octohelm/courier/pkg/courierhttp"

// ArchiveZip 定义下载 zip 压缩包接口。
type ArchiveZip struct {
	courierhttp.MethodGet `path:"/archive/zip"`
}
