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

var fileDoneStats = Stats{Type: FILEDONE}

func NewFileDoneStats() Stats {
	return fileDoneStats
}
