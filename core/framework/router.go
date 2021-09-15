package framework

type Router struct {
	factory ControllerFactory
}

func NewRouter(factory ControllerFactory) *Router {
	r := &Router{factory: factory}
	return r
}

func (r *Router) Dispatch(action string) Controller {
	return r.factory.GetController(action)
}
