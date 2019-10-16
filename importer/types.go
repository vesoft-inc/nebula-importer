package nebula_importer

type Prop struct {
	Name string
	Type string
}

type Schema struct {
	Type  string
	Name  string
	Props []Prop
}

type Stmt struct {
	Stmt string
	Data []interface{}
}

type ErrorConfig struct {
	ErrorLogPath  string
	ErrorDataPath string
}
