package base

type StatType int

const (
	SUCCESS  StatType = 0
	FAILURE  StatType = 1
	FILEDONE StatType = 2
)

type Stats struct {
	Type      StatType
	Latency   uint64
	ReqTime   float64
	BatchSize int
	Filename  string
}

func NewSuccessStats(latency uint64, reqTime float64) Stats {
	return Stats{
		Type:    SUCCESS,
		Latency: latency,
		ReqTime: reqTime,
	}
}

func NewFailureStats(batchSize int) Stats {
	return Stats{
		Type:      FAILURE,
		BatchSize: batchSize,
	}
}

func NewFileDoneStats(filename string) Stats {
	return Stats{
		Type:     FILEDONE,
		Filename: filename,
	}
}
