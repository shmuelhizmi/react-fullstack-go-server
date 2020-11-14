package react_fullstack_go_server

import (
	"encoding/json"
	"errors"
	gosocketio "github.com/graarh/golang-socketio"
	"go/types"
	"reflect"
	"sync"
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
	functionPropsHandlers := sync.Map{} // to avoid concurrent map read and map write (crash)
	transport.On("request_event", func(sender *gosocketio.Channel, data map[string]interface{}) {
		event := TransportEventRequest{
			EventArguments: data["eventArguments"].([]interface{}),
			Uuid:           data["uid"].(string),
			EventUuid:      data["eventUid"].(string),
		}
		if !isAppRunning {
			return
		}
		handler, ok := functionPropsHandlers.Load(event.EventUuid)
		if ok {
			go func() {
				// Parsing JSON into function arguments
				parameters := make([]reflect.Value, 0)
				funcHandler := reflect.ValueOf(handler)
				funcHandlerType := reflect.TypeOf(handler)
				handlerParameters := make([]reflect.Type, 0)
				handlerParametersLen := funcHandlerType.NumIn()
				for i := 0; i < handlerParametersLen; i++ {
					handlerParameters = append(handlerParameters, funcHandlerType.In(i))
				}
				for index, handlerParameter := range handlerParameters {
					emptyTypeInstance := reflect.New(handlerParameter).Interface()
					if event.EventArguments[index] == nil {
						panic(errors.New("missing argument"))
					}
					toJson, _ := json.Marshal(event.EventArguments[index])
					_ = json.Unmarshal(toJson, &emptyTypeInstance)
					parameters = append(parameters, reflect.Indirect(reflect.ValueOf(emptyTypeInstance)))
				}
				handlerResults := funcHandler.Call(parameters) // call function with parsed arguments
				eventResult := TransportEventResponse{
					Uuid:      event.Uuid,
					EventUuid: event.EventUuid,
				}
				if len(handlerResults) != 0 {
					if len(handlerResults) == 1 {
						eventResult.Data = handlerResults[0].Interface()
					} else {
						parsedHandlerResults := make([]interface{}, 0)
						for _, handlerResult := range handlerResults {
							parsedHandlerResults = append(parsedHandlerResults, handlerResult.Interface())
						}
						eventResult.Data = parsedHandlerResults
					}
				}
				_ = sender.Emit("respond_to_event", eventResult) // emit function result
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
		ListenToFunctionProps: func(propUuid string, handler interface{}) {
			functionPropsHandlers.Store(propUuid, handler)
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
