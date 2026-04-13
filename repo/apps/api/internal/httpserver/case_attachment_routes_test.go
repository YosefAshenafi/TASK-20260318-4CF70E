package httpserver

import (
	"strings"
	"testing"
)

func TestCaseAttachmentRoutesRegistered(t *testing.T) {
	required := []string{
		`GET("/cases/:id/attachments"`,
		`POST("/cases/:id/attachments"`,
		`DELETE("/cases/:id/attachments/:fileId"`,
	}
	for _, sub := range required {
		if !strings.Contains(serverGoSource, sub) {
			t.Fatalf("missing case attachment route %q", sub)
		}
	}
}
