package importer

type DataFileReader interface {
	InitFileReader(path string, stmtChs []chan Stmt, doneCh chan<- bool)
	MakeStmt([]string) Stmt
}
