package model


type MessageType string

var (
	MessageTypeNotifyClientID MessageType = "notify-client-id"
	MessageTypeLeaveClient    MessageType = "leave-client"
	MessageTypeError          MessageType = "error"
	MessageTypeKeepAlive	  MessageType = "keep-alive"
	MessageTypeGiveJob	  	  MessageType = "give-job"
	MessageTypeStillRun	  	  MessageType = "still-run"
)

type LeaveClientPayload struct {
	ClientID string `json:"client_id"`
}

type NewClientPayload struct {
	ClientID string `json:"client_id"`
}

type Message struct {
	MessageType `json:"Command"`
	Payload		`json:"Payload"`
}

type Payload interface {
}
