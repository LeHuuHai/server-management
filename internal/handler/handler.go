package handler

type Handler struct {
	*ServerHandler
	*AuthHandler
}

func NewHandler(serverHandler *ServerHandler, authHandler *AuthHandler) *Handler {
	return &Handler{
		ServerHandler: serverHandler,
		AuthHandler:   authHandler,
	}
}
