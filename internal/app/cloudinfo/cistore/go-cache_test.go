package cistore

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"logur.dev/logur"

	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/cloudinfoadapter"
)

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestCacheProductStore_Export_logsErrorDetails(t *testing.T) {
	testLogger := &logur.TestLogger{}
	store := NewCacheProductStore(time.Minute, time.Minute, cloudinfoadapter.NewLogger(testLogger))

	err := store.Export(errWriter{})
	assert.Error(t, err)

	event := findCistoreLogEvent(t, testLogger, "failed to export the store")
	assert.Equal(t, logur.Error, event.Level)
	assert.Equal(t, "write failed", event.Fields["error"])
}

func TestCacheProductStore_Import_logsErrorDetails(t *testing.T) {
	testLogger := &logur.TestLogger{}
	store := NewCacheProductStore(time.Minute, time.Minute, cloudinfoadapter.NewLogger(testLogger))

	err := store.Import(errReader{})
	assert.Error(t, err)

	event := findCistoreLogEvent(t, testLogger, "failed to load store data")
	assert.Equal(t, logur.Error, event.Level)
	assert.Equal(t, "read failed", event.Fields["error"])
}

func TestCacheProductStore_ExportImportRoundTrip(t *testing.T) {
	testLogger := &logur.TestLogger{}
	store := NewCacheProductStore(time.Minute, time.Minute, cloudinfoadapter.NewLogger(testLogger))
	store.StoreStatus("amazon", "ok")

	var buf bytes.Buffer
	require.NoError(t, store.Export(&buf))

	imported := NewCacheProductStore(time.Minute, time.Minute, cloudinfoadapter.NewLogger(&logur.TestLogger{}))
	require.NoError(t, imported.Import(&buf))

	status, ok := imported.GetStatus("amazon")
	assert.True(t, ok)
	assert.Equal(t, "ok", status)
}

func findCistoreLogEvent(t *testing.T, logger *logur.TestLogger, line string) logur.LogEvent {
	t.Helper()
	for _, event := range logger.Events() {
		if event.Line == line {
			return event
		}
	}
	t.Fatalf("expected log event %q not found in %#v", line, logger.Events())
	return logur.LogEvent{}
}
