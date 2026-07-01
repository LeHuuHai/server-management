package model

type ServerEventType string
type ServerActionType string

const (
	ServerCreateEvent ServerEventType = "ServerCreate"
	ServerUpdateEvent ServerEventType = "ServerUpdate"
	ServerDeleteEvent ServerEventType = "ServerDelete"
)

const (
	ServerCreateAction = "Create"
	ServerUpdateAction = "Update"
	ServerDeleteAction = "Delete"
)
