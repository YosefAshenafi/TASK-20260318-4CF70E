package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const HeaderRequestID = "X-Request-Id"

type Envelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
	Data      any    `json:"data,omitempty"`
}

func JSON(c *gin.Context, status int, env Envelope) {
	if env.RequestID == "" {
		env.RequestID = c.GetString("requestId")
	}
	c.JSON(status, env)
}

func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, Envelope{
		Code:      "OK",
		Message:   "success",
		RequestID: c.GetString("requestId"),
		Data:      data,
	})
}

func Error(c *gin.Context, status int, code, message string) {
	JSON(c, status, Envelope{
		Code:      code,
		Message:   message,
		RequestID: c.GetString("requestId"),
	})
}
