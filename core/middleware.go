package core

// Middleware ...
type Middleware interface {
	Run(ctx *Context) error
}
