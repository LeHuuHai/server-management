package handler

type Handler struct {
	serverHandler *ServerHandler
	authHandler   *AuthHandler
}

func NewHandler(serverHandler *ServerHandler, authHandler *AuthHandler) *Handler {
	return &Handler{
		serverHandler: serverHandler,
		authHandler:   authHandler,
	}
}
