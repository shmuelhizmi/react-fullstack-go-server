package react_fullstack_go_server

func createComponentParams(params *ComponentFactoryParams) *ComponentParams {
	return &ComponentParams{
		View: func(layer uint16, view string, viewParent *View) View {
			viewIsOn := false
			dataProps := make(map[string]interface{})
			functionProps := make(map[string]func(arguments [][]byte) interface{})
			uuidString := StringUuid()
			makeProps := func() []*ShareableViewDataProps {
				props := make([]*ShareableViewDataProps, 0, len(dataProps)+len(functionProps))
				for name, prop := range dataProps {
					props = append(props, &ShareableViewDataProps{
						Name: name,
						Type: "data",
						Uuid: "",
						Data: prop,
					})
				}
				for name, propHandler := range functionProps {
					handlerUuid := StringUuid()
					params.ListenToFunctionProps(handlerUuid, propHandler)
					props = append(props, &ShareableViewDataProps{
						Name: name,
						Type: "data",
						Uuid: handlerUuid,
						Data: nil,
					})
				}
				return props
			}
			createViewData := func() *ShareableViewData {
				viewParentUuid := params.ParentUuid
				if viewParent != nil {
					viewParentUuid = viewParent.Uuid
				}
				return &ShareableViewData{
					Uuid:       uuidString,
					Name:       view,
					ParentUuid: viewParentUuid,
					ChildIndex: layer,
					IsRoot:     params.IsRoot,
					Props:      makeProps(),
				}
			}
			return View{
				Params: dataProps,
				On: func(eventName string, handler func([][]byte) interface{}) {
					functionProps[eventName] = handler
				},
				Update: func() {
					if viewIsOn {
						params.UpdateViewData(createViewData())
					}
				},
				Start: func() {
					viewIsOn = true
					params.CreateViewData(createViewData())
				},
				Stop: func() {
					viewIsOn = false
					params.RemoveViewData(uuidString)
				},
				Uuid: uuidString,
			}
		},
		Run: func(component Component, viewParent View) (stop func()) {
			parentUuid := ""
			var cancelListeners []func()
			if &viewParent != nil {
				parentUuid = viewParent.Uuid
			}
			component(createComponentParams(&ComponentFactoryParams{
				IsRoot:                false,
				ParentUuid:            parentUuid,
				UpdateViewData:        params.UpdateViewData,
				CreateViewData:        params.CreateViewData,
				RemoveViewData:        params.RemoveViewData,
				ListenToFunctionProps: params.ListenToFunctionProps,
				ListenToComponentCancel: func(onCancel func()) {
					cancelListeners = append(cancelListeners, onCancel)
				},
			}))
			return func() {
				for _, onCancel := range cancelListeners {
					onCancel()
				}
			}
		},
	}
}
