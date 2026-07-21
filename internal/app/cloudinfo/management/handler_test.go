package management

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"logur.dev/logur"

	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/cloudinfoadapter"
)

func TestRefresh_logsErrorWhenProviderEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testLogger := &logur.TestLogger{}
	handler := &mngmntRouteHandler{log: cloudinfoadapter.NewLogger(testLogger)}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/refresh/", nil)
	handler.Refresh()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	event := findManagementLogEvent(t, testLogger, "failed to get provider")
	assert.Equal(t, logur.Error, event.Level)
	assert.Equal(t, "empty provider path param", event.Fields["reason"])
}

func findManagementLogEvent(t *testing.T, logger *logur.TestLogger, line string) logur.LogEvent {
	t.Helper()
	for _, event := range logger.Events() {
		if event.Line == line {
			return event
		}
	}
	t.Fatalf("expected log event %q not found in %#v", line, logger.Events())
	return logur.LogEvent{}
}
