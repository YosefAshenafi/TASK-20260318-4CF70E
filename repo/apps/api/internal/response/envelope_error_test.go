package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestError_envelope_shape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/fail", func(c *gin.Context) {
		c.Set("requestId", "req-err-1")
		Error(c, http.StatusForbidden, "FORBIDDEN_PERMISSION", "not allowed")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	var env Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Code != "FORBIDDEN_PERMISSION" {
		t.Errorf("code: got %q", env.Code)
	}
	if env.Message != "not allowed" {
		t.Errorf("message: got %q", env.Message)
	}
	if env.RequestID != "req-err-1" {
		t.Errorf("requestId: got %q", env.RequestID)
	}
}

func TestError_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/bad", func(c *gin.Context) {
		c.Set("requestId", "req-400")
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "bad input")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var env Envelope
	json.Unmarshal(w.Body.Bytes(), &env)
	if env.Code != "VALIDATION_ERROR" {
		t.Errorf("code: %q", env.Code)
	}
}

func TestError_500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/err", func(c *gin.Context) {
		c.Set("requestId", "req-500")
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "server error")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
