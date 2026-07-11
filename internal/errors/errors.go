package errors

import (
	"log"

	"github.com/gin-gonic/gin"
)

type ErrorHandler interface {
	error
	Handle()
}

type ContextError struct {
	Code    int
	Message string
	Context *gin.Context
}

func (e *ContextError) Error() string {
	return e.Message
}

func (e *ContextError) Handle() {
	e.Context.JSON(e.Code, gin.H{"error": e.Message})
	e.Context.AbortWithStatus(e.Code)
}

type LoadConfigError struct {
	Message string
}

func (e *LoadConfigError) Error() string {
	return e.Message
}

func (e *LoadConfigError) Handle() {
	log.Fatalf("[ERROR] Could not load config: %s", e.Message)
}
