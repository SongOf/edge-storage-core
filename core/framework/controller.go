package framework

import "github.com/SongOf/edge-storage-core/core"

type ControllerFactory interface {
	GetController(string) Controller
}

type Controller interface {
	GetDescription() ControllerDescription
	Entry(ctx *core.Context) (ControllerResult, error)
}

type ControllerDescription interface {
	Spec()
}

type ControllerResult interface {
	WithRequestId(string) ControllerResult
}

type BaseDescription struct {
	Action    string
	RequestId string
}

func (description *BaseDescription) Spec() {
	// TODO
}

type BaseResponse struct {
	RequestId string
}

func (response *BaseResponse) WithRequestId(requestId string) ControllerResult {
	response.RequestId = requestId
	return response
}
