package base

type Stmt struct {
	Stmt string
	Data [][]interface{}
}

type ErrData struct {
	Error error
	Data  [][]interface{}
	Done  bool
}
