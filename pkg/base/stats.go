package base

type StatType int

const (
	SUCCESS  StatType = 0
	FAILURE  StatType = 1
	FILEDONE StatType = 2
	OUTPUT   StatType = 3
)

const STAT_FILEDONE string = "FILEDONE"

type Stats struct {
	Type          StatType
	Latency       int64
	ReqTime       int64
	BatchSize     int
	ImportedBytes int64
	Filename      string
}

func NewSuccessStats(latency int64, reqTime int64, batchSize int, importedBytes int64) Stats {
	return Stats{
		Type:          SUCCESS,
		Latency:       latency,
		ReqTime:       reqTime,
		BatchSize:     batchSize,
		ImportedBytes: importedBytes,
	}
}

func NewFailureStats(batchSize int, importedBytes int64) Stats {
	return Stats{
		Type:          FAILURE,
		BatchSize:     batchSize,
		ImportedBytes: importedBytes,
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

func NewOutputStats() Stats {
	return Stats{
		Type: OUTPUT,
	}
}
