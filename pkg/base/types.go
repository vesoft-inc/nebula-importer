package base

type Stmt struct {
	Stmt string
	Data [][]interface{}
}

type ErrData struct {
	Error error
	Data  Record
	Done  bool
}

type Record []string

func DoneRecord() Record {
	return Record{"DONE"}
}
