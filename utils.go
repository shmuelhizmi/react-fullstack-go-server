package react_fullstack_go_server

import "github.com/google/uuid"

func stringUuid() string  {
	uuidByte, _ := uuid.NewUUID()
	uuidString := ""
	for _, currentByte := range uuidByte {
		uuidString+= string(currentByte)
	}
	return uuidString
}
