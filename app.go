package react_fullstack_go_server

import (
	"encoding/json"
	gosocketio "github.com/graarh/golang-socketio"
)

func App(transport *gosocketio.Server, rootComponent func(params *ComponentParams)) AppInstance {
	isAppRunning := true
	var shareableViewData []*ShareableViewData
	transport.On("request_views_tree", func(sender *gosocketio.Channel) {
		sender.Emit("update_views_tree", TransportUpdateTree{Views: shareableViewData})
	})
	transport.On(gosocketio.OnConnection, func(sender *gosocketio.Channel) {
		sender.Emit("update_views_tree", TransportUpdateTree{Views: shareableViewData})
	})
	var rootComponentCancelListeners []func()
	go rootComponent(createComponentParams(&ComponentFactoryParams{
		IsRoot:     true,
		ParentUuid: "",
		UpdateViewData: func(viewData *ShareableViewData) {
			for index, currentViewData := range shareableViewData {
				if currentViewData.Uuid == viewData.Uuid {
					shareableViewData[index] = viewData
				}
			}
			transport.BroadcastToAll("update_view", TransportViewUpdate{
				View: viewData,
			})
		},
		CreateViewData: func(viewData *ShareableViewData) {
			shareableViewData = append(shareableViewData, viewData)
			if isAppRunning {
				transport.BroadcastToAll("update_view", TransportViewUpdate{
					View: viewData,
				})
			}
		},
		RemoveViewData: func(uuidString string) {
			var index int
			for currentIndex, currentView := range shareableViewData {
				if currentView.Uuid == uuidString {
					index = currentIndex
				}
			}
			shareableViewData = append(shareableViewData[:index], shareableViewData[index+1:]...)
			if isAppRunning {
				transport.BroadcastToAll("delete_view", TransportViewDelete{ViewUuid: uuidString})
			}
		},
		ListenToFunctionProps: func(propUuid string, handler func(data [][]byte) interface{}) {
			transport.On("request_event", func(sender *gosocketio.Channel, event TransportEventRequest) {
				if !isAppRunning {
					return
				}
				if event.EventUuid == propUuid {
					go func() {
						handlerResult := handler(event.EventArguments)
						result, _ := json.Marshal(TransportEventResponse{
							Data:      handlerResult,
							Uuid:      event.Uuid,
							EventUuid: propUuid,
						})
						sender.Emit("respond_to_event", result)
					}()
				}
			})
		},
		ListenToComponentCancel: func(onCancel func()) {
			rootComponentCancelListeners = append(rootComponentCancelListeners, onCancel)
		},
	}))
	return AppInstance{
		IsAppRunning: &isAppRunning,
		Stop: func() {
			isAppRunning = false
		},
		Continue: func() {
			isAppRunning = true
			transport.BroadcastToAll("update_views_tree", TransportUpdateTree{Views: shareableViewData})
		},
		Cancel: func() {
			for _, onCancel := range rootComponentCancelListeners {
				onCancel()
			}
		},
	}
}
