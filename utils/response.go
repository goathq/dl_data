package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func RespondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "success",
		Data:    data,
	})
}

func RespondWithError(c *gin.Context, statusCode int, errorMessage string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   errorMessage,
	})
}
