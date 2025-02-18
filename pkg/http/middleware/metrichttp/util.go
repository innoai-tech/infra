package metrichttp

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
	B   = 1
	KiB = 1024 * B
	MiB = 1024 * KiB
)

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
