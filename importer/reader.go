package importer

type DataFileReader interface {
	InitFileReader(stmtChs []chan Stmt, doneCh chan<- bool)
	MakeStmt([][]string, int) Stmt
}
