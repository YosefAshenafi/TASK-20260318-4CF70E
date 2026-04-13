package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOK_EnvelopeShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("requestId", "req_test")
	OK(c, gin.H{"x": 1})

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}
