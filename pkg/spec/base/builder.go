//go:generate mockgen -source=builder.go -destination builder_mock.go -package specbase StatementBuilder
package specbase

type (
	// StatementBuilder is the interface to build statement
	StatementBuilder interface {
		Build(records ...Record) (string, error)
	}

	StatementBuilderFunc func(records ...Record) (string, error)
)

func (f StatementBuilderFunc) Build(records ...Record) (string, error) {
	return f(records...)
}
