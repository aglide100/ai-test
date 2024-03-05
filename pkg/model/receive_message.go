package model

import (
	"encoding/json"
	"log"
)

type ReceiveMessageType string

var (
	ReceiveMessageTypeDoneJob   ReceiveMessageType = "DoneJob"
	ReceiveMessageTypeKeepAlive   ReceiveMessageType = "KeepAlive"
	ReceiveMessageTypeError   ReceiveMessageType = "Error"
	ReceiveMessageStillRun  ReceiveMessageType = "StillRun"
)

type ReceiveMessage struct {
	Command ReceiveMessageType `json:"Command"`
	Payload json.RawMessage `json:"Payload"`
}

type ErrorMessage string

func (e ErrorMessage) Error() string {
	return string(e)
}
 
var (
	ErrMsgInvalidType    ErrorMessage = "invalid type"
	ErrMsgInvalidPayload ErrorMessage = "invalid payload"
	ErrMsgInternalError  ErrorMessage = "internal error"
)

func NewErrorMessage(msg error) *Message {
	log.Println("error:", msg)
	return &Message{
		"error",
		[]byte(msg.Error()),
	}
}
