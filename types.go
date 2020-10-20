package react_fullstack_go_server

import "go/types"

type View struct {
	Params map[string]interface{}
	On     func(eventName string, handler func([][]byte) interface{})
	Update func()
	Start  func()
	Stop   func()
	Uuid   string
}

type ShareableViewDataProps struct {
	Name string      `json:"name"`
	Type string      `json:"type"`
	Uuid string      `json:"uid"`
	Data interface{} `json:"data"`
}

type ShareableViewData struct {
	Uuid       string                    `json:"uid"`
	Name       string                    `json:"name"`
	ParentUuid string                    `json:"parentUid"`
	ChildIndex uint16                    `json:"childIndex"`
	IsRoot     bool                      `json:"isRoot"`
	Props      []*ShareableViewDataProps `json:"props"`
}

type Component func(params *ComponentParams)

type CreateView func(layer uint16, view string, viewParent *View) View

type ComponentParams struct {
	View   CreateView
	Run    func(component Component, viewParent View) (stop func())
	Cancel <-chan *types.Nil
}

type AppInstance struct {
	IsAppRunning *bool
	Stop         func()
	Continue     func()
	Cancel       func()
}

type ComponentFactoryParams struct {
	IsRoot                bool
	ParentUuid            string
	UpdateViewData        func(viewData *ShareableViewData)
	CreateViewData        func(viewData *ShareableViewData)
	RemoveViewData        func(uuidString string)
	ListenToFunctionProps func(propUuid string, handler func(data [][]byte) interface{})
	CancelChan            <-chan *types.Nil
}

type TransportEventRequest struct {
	EventArguments []interface{} `json:"eventArguments"`
	Uuid           string        `json:"uid"`
	EventUuid      string        `json:"eventUid"`
}

type TransportEventResponse struct {
	Data      interface{} `json:"data"`
	Uuid      string      `json:"uid"`
	EventUuid string      `json:"eventUid"`
}

type TransportViewUpdate struct {
	View *ShareableViewData `json:"view"`
}

type TransportViewDelete struct {
	ViewUuid string `json:"viewUid"`
}

type TransportUpdateTree struct {
	Views []*ShareableViewData `json:"views"`
}
