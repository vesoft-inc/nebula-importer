package importer

type Stmt struct {
	Stmt string
	Data [][]interface{}
}

type ErrorConfig struct {
	ErrorLogPath  string
	ErrorDataPath string
}

type ErrData struct {
	Error error
	Data  [][]interface{}
	Done  bool
}

type Stats struct {
	Latency uint64
	ReqTime float64
}
