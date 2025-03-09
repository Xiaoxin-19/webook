package web

import "github.com/gin-gonic/gin"

type Handler interface {
	RegisterRoutes(server *gin.Engine)
}

type Page struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}
