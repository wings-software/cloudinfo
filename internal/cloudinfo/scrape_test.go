// Copyright © 2019 Banzai Cloud
package cloudinfo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/banzaicloud/cloudinfo/internal/app/cloudinfo/messaging"
	"github.com/banzaicloud/cloudinfo/internal/app/cloudinfo/tracing"
	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/metrics"
	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/types"
)

type stubMetricsReporter struct{}

func (stubMetricsReporter) ReportScrapeProviderCompleted(string, time.Time)              {}
func (stubMetricsReporter) ReportScrapeRegionCompleted(string, string, string, time.Time) {}
func (stubMetricsReporter) ReportScrapeFailure(string, string, string)                   {}
func (stubMetricsReporter) ReportScrapeProviderShortLivedCompleted(string, time.Time)    {}
func (stubMetricsReporter) ReportScrapeRegionShortLivedCompleted(string, string, time.Time) {
}
func (stubMetricsReporter) ReportScrapeShortLivedFailure(string, string) {}

type stubEventBus struct{}

func (stubEventBus) PublishScrapingComplete(string)                {}
func (stubEventBus) SubscribeScrapingComplete(string, interface{}) {}

type stubErrorHandler struct{}

func (stubErrorHandler) Handle(error) {}

type stubInfoer struct {
	regionsErr error
}

func (s *stubInfoer) Initialize() (map[string]map[string]types.Price, error) {
	return nil, nil
}
func (s *stubInfoer) GetVirtualMachines(string) ([]types.VMInfo, error) { return nil, nil }
func (s *stubInfoer) GetProducts([]types.VMInfo, string, string) ([]types.VMInfo, error) {
	return nil, nil
}
func (s *stubInfoer) GetZones(string) ([]string, error) { return nil, nil }
func (s *stubInfoer) GetRegions(string) (map[string]string, error) {
	if s.regionsErr != nil {
		return nil, s.regionsErr
	}
	return map[string]string{"us-west-1": "US West"}, nil
}
func (s *stubInfoer) HasShortLivedPriceInfo() bool { return false }
func (s *stubInfoer) GetCurrentPrices(string) (map[string]types.Price, error) {
	return nil, nil
}
func (s *stubInfoer) HasImages() bool { return false }
func (s *stubInfoer) GetServiceImages(string, string) ([]types.Image, error) {
	return nil, nil
}
func (s *stubInfoer) GetVersions(string, string) ([]types.LocationVersion, error) {
	return nil, nil
}
func (s *stubInfoer) GetServiceProducts(string, string) ([]types.ProductDetails, error) {
	return nil, nil
}

type logEvent struct {
	level  string
	line   string
	fields map[string]interface{}
}

type recordingLogger struct {
	events *[]logEvent
	fields map[string]interface{}
}

func newRecordingLogger() *recordingLogger {
	events := make([]logEvent, 0)
	return &recordingLogger{events: &events}
}

func (l *recordingLogger) record(level, msg string, fields ...map[string]interface{}) {
	merged := map[string]interface{}{}
	for k, v := range l.fields {
		merged[k] = v
	}
	if len(fields) > 0 {
		for k, v := range fields[0] {
			merged[k] = v
		}
	}
	*l.events = append(*l.events, logEvent{level: level, line: msg, fields: merged})
}

func (l *recordingLogger) Trace(msg string, fields ...map[string]interface{}) {
	l.record("trace", msg, fields...)
}
func (l *recordingLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.record("debug", msg, fields...)
}
func (l *recordingLogger) Info(msg string, fields ...map[string]interface{}) {
	l.record("info", msg, fields...)
}
func (l *recordingLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.record("warn", msg, fields...)
}
func (l *recordingLogger) Error(msg string, fields ...map[string]interface{}) {
	l.record("error", msg, fields...)
}
func (l *recordingLogger) WithFields(fields map[string]interface{}) Logger {
	merged := map[string]interface{}{}
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &recordingLogger{events: l.events, fields: merged}
}
func (l *recordingLogger) WithContext(ctx context.Context) Logger { return l }

func findScrapeLogEvent(t *testing.T, logger *recordingLogger, line string) logEvent {
	t.Helper()
	for _, event := range *logger.events {
		if event.line == line {
			return event
		}
	}
	t.Fatalf("expected log event %q not found in %#v", line, *logger.events)
	return logEvent{}
}

func newTestScrapingManager(provider string, infoer CloudInfoer, store CloudInfoStore, logger Logger) *scrapingManager {
	return NewScrapingManager(
		provider,
		infoer,
		store,
		logger,
		stubMetricsReporter{},
		tracing.NewNoOpTracer(),
		stubEventBus{},
		stubErrorHandler{},
	)
}

func TestScrapingManager_scrapeServiceInformation_logsWhenServicesMissing(t *testing.T) {
	logger := newRecordingLogger()
	sm := newTestScrapingManager(
		"azure",
		&stubInfoer{},
		&DummyCloudInfoStore{TcId: notCached},
		logger,
	)

	sm.scrapeServiceInformation(context.Background())

	event := findScrapeLogEvent(t, logger, "failed to retrieve services")
	assert.Equal(t, "error", event.level)
	assert.Equal(t, "azure", event.fields["provider"])
}

func TestScrapingManager_scrapeServiceInformation_logsWhenRegionScrapeFails(t *testing.T) {
	logger := newRecordingLogger()
	regionErr := errors.New("regions unavailable")
	sm := newTestScrapingManager(
		"azure",
		&stubInfoer{regionsErr: regionErr},
		&DummyCloudInfoStore{},
		logger,
	)

	sm.scrapeServiceInformation(context.Background())

	event := findScrapeLogEvent(t, logger, "failed to scrape service region information")
	assert.Equal(t, "error", event.level)
	assert.Equal(t, "azure", event.fields["provider"])
	assert.Contains(t, event.fields["error"], regionErr.Error())
}

var _ metrics.Reporter = stubMetricsReporter{}
var _ messaging.EventBus = stubEventBus{}
