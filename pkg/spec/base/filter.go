package specbase

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Filter struct {
	Expr    string `yaml:"expr,omitempty"`
	program *vm.Program
}

func (f *Filter) Build() error {
	env := map[string]any{
		"Record": Record{},
	}
	program, err := expr.Compile(f.Expr, expr.Env(env), expr.AsBool())
	if err != nil {
		return err
	}
	f.program = program
	return nil
}

func (f *Filter) Filter(record Record) (bool, error) {
	env := map[string]any{
		"Record": record,
	}
	out, err := expr.Run(f.program, env)
	if err != nil {
		return false, err
	}
	return out.(bool), nil
}
