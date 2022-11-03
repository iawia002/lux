package gojq

import "fmt"

// CompilerOption is a compiler option.
type CompilerOption func(*compiler)

// WithModuleLoader is a compiler option for module loader.
// If you want to load modules from the filesystem, use NewModuleLoader.
func WithModuleLoader(moduleLoader ModuleLoader) CompilerOption {
	return func(c *compiler) {
		c.moduleLoader = moduleLoader
	}
}

// WithEnvironLoader is a compiler option for environment variables loader.
// The OS environment variables are not accessible by default due to security
// reason. You can pass os.Environ if you allow to access it.
func WithEnvironLoader(environLoader func() []string) CompilerOption {
	return func(c *compiler) {
		c.environLoader = environLoader
	}
}

// WithVariables is a compiler option for variable names. The variables can be
// used in the query. You have to give the values to code.Run in the same order.
func WithVariables(variables []string) CompilerOption {
	return func(c *compiler) {
		c.variables = variables
	}
}

// WithFunction is a compiler option for adding a custom internal function.
// Specify the minimum and maximum count of the function arguments. These
// values should satisfy 0 <= minarity <= maxarity <= 30, otherwise panics.
// On handling numbers, you should take account to int, float64 and *big.Int.
// These are the number types you are allowed to return, so do not return int64.
// Refer to ValueError to return a value error just like built-in error function.
// If you want to emit multiple values, call the empty function, accept a filter
// for its argument, or call another built-in function, then use LoadInitModules
// of the module loader.
func WithFunction(name string, minarity, maxarity int,
	f func(interface{}, []interface{}) interface{}) CompilerOption {
	return withFunction(name, minarity, maxarity, false, f)
}

// WithIterFunction is a compiler option for adding a custom iterator function.
// This is like the WithFunction option, but you can add a function which
// returns an Iter to emit multiple values. You cannot define both iterator and
// non-iterator functions of the same name (with possibly different arities).
// See also NewIter, which can be used to convert values or an error to an Iter.
func WithIterFunction(name string, minarity, maxarity int,
	f func(interface{}, []interface{}) Iter) CompilerOption {
	return withFunction(name, minarity, maxarity, true,
		func(v interface{}, args []interface{}) interface{} {
			return f(v, args)
		},
	)
}

func withFunction(name string, minarity, maxarity int, iter bool,
	f func(interface{}, []interface{}) interface{}) CompilerOption {
	if !(0 <= minarity && minarity <= maxarity && maxarity <= 30) {
		panic(fmt.Sprintf("invalid arity for %q: %d, %d", name, minarity, maxarity))
	}
	argcount := 1<<(maxarity+1) - 1<<minarity
	return func(c *compiler) {
		if c.customFuncs == nil {
			c.customFuncs = make(map[string]function)
		}
		if fn, ok := c.customFuncs[name]; ok {
			if fn.iter != iter {
				panic(fmt.Sprintf("cannot define both iterator and non-iterator functions for %q", name))
			}
			c.customFuncs[name] = function{
				argcount | fn.argcount, iter,
				func(x interface{}, xs []interface{}) interface{} {
					if argcount&(1<<len(xs)) != 0 {
						return f(x, xs)
					}
					return fn.callback(x, xs)
				},
			}
		} else {
			c.customFuncs[name] = function{argcount, iter, f}
		}
	}
}

// WithInputIter is a compiler option for input iterator used by input(s)/0.
// Note that input and inputs functions are not allowed by default. We have
// to distinguish the query input and the values for input(s) functions. For
// example, consider using inputs with --null-input. If you want to allow
// input(s) functions, create an Iter and use WithInputIter option.
func WithInputIter(inputIter Iter) CompilerOption {
	return func(c *compiler) {
		c.inputIter = inputIter
	}
}
