package driver

import (
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type ExprCompiler struct {
	cache *sync.Map
}

func NewExprCompiler() *ExprCompiler {
	return &ExprCompiler{
		cache: new(sync.Map),
	}
}

func (c *ExprCompiler) Eval(exp string, data any) (any, error) {
	var exec *vm.Program
	var err error
	if value, ok := c.cache.Load(exp); ok {
		exec = value.(*vm.Program)
	} else {
		exec, err = expr.Compile(exp)
		if err != nil {
			return nil, err
		}
		c.cache.Store(exp, exec)
	}
	return expr.Run(exec, data)
}
