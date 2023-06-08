//go:generate mockgen -source=builder.go -destination builder_mock.go -package specbase StatementBuilder
package specbase

type (
	// StatementBuilder is the interface to build statement
	StatementBuilder interface {
		Build(records ...Record) (statement string, nRecord int, err error)
	}

	StatementBuilderFunc func(records ...Record) (statement string, nRecord int, err error)
)

func (f StatementBuilderFunc) Build(records ...Record) (statement string, nRecord int, err error) {
	return f(records...)
}
