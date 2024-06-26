package profiler

import (
	"fmt"
	"strings"

	"cloud.google.com/go/profiler"
	"github.com/sirupsen/logrus"
)

func StartProfiling(serviceName, projectID string) error {
	cfg := profiler.Config{
		Service:        strings.ToLower(serviceName), // needs to be in lowercase
		ServiceVersion: "1.0.0",
	}

	if len(projectID) > 0 {
		cfg.ProjectID = projectID
	}

	if err := profiler.Start(cfg); err != nil {
		logrus.WithError(err).Errorf("Unable to start profiler")
		return fmt.Errorf("Unable to start profiling. %s", err)
	}
	return nil
}
