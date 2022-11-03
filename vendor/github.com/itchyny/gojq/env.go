package gojq

import "context"

type env struct {
	pc        int
	stack     *stack
	paths     *stack
	scopes    *scopeStack
	values    []interface{}
	codes     []*code
	codeinfos []codeinfo
	forks     []fork
	backtrack bool
	offset    int
	expdepth  int
	label     int
	args      [32]interface{} // len(env.args) > maxarity
	ctx       context.Context
}

func newEnv(ctx context.Context) *env {
	return &env{
		stack:  newStack(),
		paths:  newStack(),
		scopes: newScopeStack(),
		ctx:    ctx,
	}
}

type scope struct {
	id         int
	offset     int
	pc         int
	saveindex  int
	outerindex int
}

type fork struct {
	pc         int
	stackindex int
	stacklimit int
	scopeindex int
	scopelimit int
	pathindex  int
	pathlimit  int
	expdepth   int
}
