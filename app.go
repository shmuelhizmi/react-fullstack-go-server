package react_fullstack_go_server

import (
	"encoding/json"
	gosocketio "github.com/graarh/golang-socketio"
	"go/types"
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
	cancelChannel := make(chan *types.Nil)
	functionPropsHandlers := make(map[string]func(data [][]byte) interface{})
	transport.On("request_event", func(sender *gosocketio.Channel, data map[string]interface{}) {
		event := TransportEventRequest{
			EventArguments: data["eventArguments"].([]interface{}),
			Uuid:           data["uid"].(string),
			EventUuid:      data["eventUid"].(string),
		}
		if !isAppRunning {
			return
		}
		handler, ok := functionPropsHandlers[event.EventUuid]
		if ok {
			go func() {
				var arguments [][]byte
				for _, data := range event.EventArguments {
					argument, _ := json.Marshal(data)
					arguments = append(arguments, argument)
				}
				handlerResult := handler(arguments)
				result, _ := json.Marshal(TransportEventResponse{
					Data:      handlerResult,
					Uuid:      event.Uuid,
					EventUuid: event.EventUuid,
				})
				sender.Emit("respond_to_event", result)
			}()
		}
	})
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
			functionPropsHandlers[propUuid] = handler
		},
		CancelChan: cancelChannel,
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
			cancelChannel <- nil
		},
	}
}
