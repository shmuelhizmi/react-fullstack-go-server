package react_fullstack_go_server

import (
	"github.com/google/uuid"
)

func StringUuid() string {
	uuidByte, _ := uuid.NewUUID()
	uuidString := ""
	for _, currentByte := range uuidByte {
		uuidString += string(currentByte)
	}
	return uuidString
}