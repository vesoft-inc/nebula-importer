package base

type StatType int

const (
	SUCCESS  StatType = 0
	FAILURE  StatType = 1
	FILEDONE StatType = 2
)

const STAT_FILEDONE string = "FILEDONE"

type Stats struct {
	Type      StatType
	Latency   int64
	ReqTime   int64
	BatchSize int
	Filename  string
}

func NewSuccessStats(latency int64, reqTime int64, batchSize int) Stats {
	return Stats{
		Type:      SUCCESS,
		Latency:   latency,
		ReqTime:   reqTime,
		BatchSize: batchSize,
	}
}

func NewFailureStats(batchSize int) Stats {
	return Stats{
		Type:      FAILURE,
		BatchSize: batchSize,
	}
}

func NewFileDoneStats(filename string) Stats {
	// When goto this step, we have finished configure file validation
	// and it's safe to ignore following error
	fpath, _ := FormatFilePath(filename)
	return Stats{
		Type:     FILEDONE,
		Filename: fpath,
	}
}
