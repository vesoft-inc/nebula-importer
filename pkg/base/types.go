package base

type Stmt struct {
	Stmt string
	Data [][]interface{}
}

type Record []string

type OpType int

const (
	DONE   OpType = 0
	INSERT OpType = 1
	DELETE OpType = 2
	HEADER OpType = 100
)

func (op OpType) String() string {
	switch op {
	case 0:
		return "DONE"
	case 1:
		return "INSERT"
	case 2:
		return "DELETE"
	case 100:
		return "HEADER"
	default:
		return "UNKNOWN"
	}
}

type Data struct {
	Type   OpType
	Record Record
}

func InsertData(record Record) Data {
	return Data{
		Type:   INSERT,
		Record: record,
	}
}

func DeleteData(record Record) Data {
	return Data{
		Type:   DELETE,
		Record: record,
	}
}

func HeaderData(record Record) Data {
	return Data{
		Type:   HEADER,
		Record: record,
	}
}

func FinishData() Data {
	return Data{
		Type:   DONE,
		Record: nil,
	}
}

type ErrData struct {
	Error error
	Data  Data
}
