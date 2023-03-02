package picker

import (
	"sync"
)

var valuePool = sync.Pool{
	New: func() any {
		return &Value{}
	},
}

type Value struct {
	Val       string
	IsNull    bool
	isSetNull bool
}

func NewValue(val string) *Value {
	v := valuePool.Get().(*Value)
	v.Val = val
	v.IsNull = false
	v.isSetNull = false
	return v
}

func (v *Value) Release() {
	valuePool.Put(v)
}
