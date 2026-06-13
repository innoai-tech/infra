package metrichttp

// DurationHistogramBoundaries 定义 HTTP 请求耗时直方图的桶边界（秒）。
var DurationHistogramBoundaries = []float64{
	0,
	0.005,
	0.01,
	0.025,
	0.05,
	0.075,
	0.1,
	0.25,
	0.5,
	0.75,
	1,
	2.5,
	5,
	7.5,
	10,
}

const (
	// B 表示 1 字节。
	B = 1
	// KiB 表示 1024 字节。
	KiB = 1024 * B
	// MiB 表示 1024 * 1024 字节。
	MiB = 1024 * KiB
)

// SizeHistogramBoundaries 定义 HTTP 请求/响应体大小的直方图桶边界（字节）。
var SizeHistogramBoundaries = []float64{
	512 * B,
	1 * KiB,
	64 * KiB,
	128 * KiB,
	256 * KiB,
	512 * KiB,
	1 * MiB,
	2 * MiB,
	5 * MiB,
	10 * MiB,
	20 * MiB,
	50 * MiB,
}
