package react_fullstack_go_server


func createComponentParams(params *ComponentFactoryParams) *ComponentParams {
	return &ComponentParams{
		View: func(layer uint16, view string) View {
			viewIsOn := false
			dataProps := make(map[string]interface{})
			functionProps := make(map[string]func(arguments [][]byte) interface{})
			uuidString := stringUuid()
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
					handlerUuid := stringUuid()
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
				return &ShareableViewData{
					Uuid:       uuidString,
					Name:       view,
					ParentUuid: params.ParentUuid,
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
		Run: func(component func(params *ComponentParams), viewParent View) {
			parentUuid := ""
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
			}))
		},
	}
}
