package main

import (
	"github.com/gin-gonic/gin"
	"webok/internal/events"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
}
