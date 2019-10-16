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
