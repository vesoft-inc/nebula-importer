package reader

import "github.com/yixinglu/nebula-importer/pkg/base"

type DataFileReader interface {
	InitFileReader(stmtChs []chan base.Stmt, doneCh chan<- bool)
	MakeStmt([][]string, int) base.Stmt
}
