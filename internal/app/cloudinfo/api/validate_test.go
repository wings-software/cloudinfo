// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"logur.dev/logur"

	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/cloudinfoadapter"
	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/types"
)

func TestGetProviderPathParamsValidation(t *testing.T) {
	tests := []struct {
		name      string
		pathParam interface{}
		check     func(t *testing.T, err error)
	}{
		{
			name:      "getProvider path params validation should fail when provider not specified",
			pathParam: &GetProviderPathParams{},
			check: func(t *testing.T, err error) {
				assert.NotNil(t, err, "validation should fail", err)
			},
		},
		{
			name:      "getProvider path params validation should fail when provider is not supported",
			pathParam: &GetProviderPathParams{Provider: "unsupported"},
			check: func(t *testing.T, err error) {
				assert.NotNil(t, err, "validation should fail %#V", err)
			},
		},
		{
			name:      "getProvider path params validation should pass when provider is supported",
			pathParam: &GetProviderPathParams{Provider: "test-provider-1"},
			check: func(t *testing.T, err error) {
				assert.Nil(t, err, "validation should not fail")
			},
		},
	}

	// setup the validator
	v := validator.New()
	v.SetTagName("binding")
	if err := v.RegisterValidation("provider", providerValidator([]string{"test-provider-1", "test-provider-2"})); err != nil {
		t.Fatal("failed to register provider validator")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.check(t, v.Struct(test.pathParam))
		})
	}
}

func TestGetServicesPathParamsValidation(t *testing.T) {
	providers := []string{"amazon", "azure"}
	ci := &stubCloudInfo{
		services: map[string][]types.Service{
			"amazon": {
				{Service: "compute"},
				{Service: "eks"},
			},
		},
	}

	v := newTestValidator(t, providers, ci)

	tests := []struct {
		name      string
		pathParam GetServicesPathParams
		wantErr   bool
	}{
		{
			name:      "fails for unsupported provider without calling GetServices",
			pathParam: GetServicesPathParams{GetProviderPathParams: GetProviderPathParams{Provider: "INVALID_PROVIDER"}, Service: "compute"},
			wantErr:   true,
		},
		{
			name:      "fails when service is missing",
			pathParam: GetServicesPathParams{GetProviderPathParams: GetProviderPathParams{Provider: "amazon"}},
			wantErr:   true,
		},
		{
			name:      "fails when service is not supported for provider",
			pathParam: GetServicesPathParams{GetProviderPathParams: GetProviderPathParams{Provider: "amazon"}, Service: "unknown"},
			wantErr:   true,
		},
		{
			name:      "fails when GetServices returns error",
			pathParam: GetServicesPathParams{GetProviderPathParams: GetProviderPathParams{Provider: "azure"}, Service: "compute"},
			wantErr:   true,
		},
		{
			name:      "passes for supported provider and service",
			pathParam: GetServicesPathParams{GetProviderPathParams: GetProviderPathParams{Provider: "amazon"}, Service: "compute"},
			wantErr:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ci.getServicesCalls = nil

			err := v.Struct(test.pathParam)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if test.pathParam.Provider == "INVALID_PROVIDER" {
				assert.Empty(t, ci.getServicesCalls, "GetServices must not be called for unsupported providers")
			}
		})
	}
}

func TestServiceValidator_logsErrorWhenGetServicesFails(t *testing.T) {
	providers := []string{"azure"}
	ci := &stubCloudInfo{}
	testLogger := &logur.TestLogger{}
	logger := cloudinfoadapter.NewLogger(testLogger)

	v := validator.New()
	v.SetTagName("binding")
	require.NoError(t, v.RegisterValidation("provider", providerValidator(providers)))
	require.NoError(t, v.RegisterValidation("service", serviceValidator(providers, ci, logger)))

	err := v.Struct(GetServicesPathParams{
		GetProviderPathParams: GetProviderPathParams{Provider: "azure"},
		Service:               "compute",
	})
	assert.Error(t, err)

	event := findValidateLogEvent(t, testLogger, "validation failed, could not retrieve services")
	assert.Equal(t, logur.Warn, event.Level)
	assert.Equal(t, "services not yet cached", event.Fields["error"])
}

func TestGetRegionPathParamsValidation(t *testing.T) {
	providers := []string{"amazon"}
	ci := &stubCloudInfo{
		services: map[string][]types.Service{
			"amazon": {{Service: "compute"}},
		},
		regions: map[string]map[string]string{
			"amazon/compute": {
				"us-west-1": "US West",
				"us-east-1": "US East",
			},
		},
	}

	v := newTestValidator(t, providers, ci)

	tests := []struct {
		name      string
		pathParam GetRegionPathParams
		wantErr   bool
	}{
		{
			name: "fails for unsupported provider without calling GetRegions",
			pathParam: GetRegionPathParams{
				GetServicesPathParams: GetServicesPathParams{
					GetProviderPathParams: GetProviderPathParams{Provider: "INVALID_PROVIDER"},
					Service:               "compute",
				},
				Region: "us-west-1",
			},
			wantErr: true,
		},
		{
			name: "fails when region is not supported",
			pathParam: GetRegionPathParams{
				GetServicesPathParams: GetServicesPathParams{
					GetProviderPathParams: GetProviderPathParams{Provider: "amazon"},
					Service:               "compute",
				},
				Region: "eu-central-1",
			},
			wantErr: true,
		},
		{
			name: "passes for supported provider, service and region",
			pathParam: GetRegionPathParams{
				GetServicesPathParams: GetServicesPathParams{
					GetProviderPathParams: GetProviderPathParams{Provider: "amazon"},
					Service:               "compute",
				},
				Region: "us-west-1",
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ci.getRegionsCalls = nil

			err := v.Struct(test.pathParam)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if test.pathParam.Provider == "INVALID_PROVIDER" {
				assert.Empty(t, ci.getRegionsCalls, "GetRegions must not be called for unsupported providers")
			}
		})
	}
}

func TestRegionValidator_logsErrorWhenGetRegionsFails(t *testing.T) {
	providers := []string{"amazon"}
	ci := &stubCloudInfo{
		services: map[string][]types.Service{
			"amazon": {{Service: "compute"}},
		},
	}
	testLogger := &logur.TestLogger{}
	logger := cloudinfoadapter.NewLogger(testLogger)

	v := validator.New()
	v.SetTagName("binding")
	require.NoError(t, v.RegisterValidation("provider", providerValidator(providers)))
	require.NoError(t, v.RegisterValidation("service", serviceValidator(providers, ci, logger)))
	require.NoError(t, v.RegisterValidation("region", regionValidator(providers, ci, logger)))

	err := v.Struct(GetRegionPathParams{
		GetServicesPathParams: GetServicesPathParams{
			GetProviderPathParams: GetProviderPathParams{Provider: "amazon"},
			Service:               "compute",
		},
		Region: "us-west-1",
	})
	assert.Error(t, err)

	event := findValidateLogEvent(t, testLogger, "validation failed, could not retrieve regions")
	assert.Equal(t, logur.Warn, event.Level)
	assert.Equal(t, "regions not yet cached", event.Fields["error"])
}

func TestIsSupportedProvider(t *testing.T) {
	providers := []string{"amazon", "azure"}

	assert.True(t, isSupportedProvider(providers, "amazon"))
	assert.False(t, isSupportedProvider(providers, "INVALID_PROVIDER"))
	assert.False(t, isSupportedProvider(providers, ""))
}

func findValidateLogEvent(t *testing.T, logger *logur.TestLogger, line string) logur.LogEvent {
	t.Helper()
	for _, event := range logger.Events() {
		if event.Line == line {
			return event
		}
	}
	t.Fatalf("expected log event %q not found in %#v", line, logger.Events())
	return logur.LogEvent{}
}

func newTestValidator(t *testing.T, providers []string, ci types.CloudInfo) *validator.Validate {
	t.Helper()

	v := validator.New()
	v.SetTagName("binding")

	logger := cloudinfoadapter.NewNoopLogger()
	require.NoError(t, v.RegisterValidation("provider", providerValidator(providers)))
	require.NoError(t, v.RegisterValidation("service", serviceValidator(providers, ci, logger)))
	require.NoError(t, v.RegisterValidation("region", regionValidator(providers, ci, logger)))
	require.NoError(t, v.RegisterValidation("product", productValidator(ci, logger)))

	return v
}

type stubCloudInfo struct {
	services         map[string][]types.Service
	regions          map[string]map[string]string
	getServicesCalls []string
	getRegionsCalls  []string
}

func (s *stubCloudInfo) GetServices(provider string) ([]types.Service, error) {
	s.getServicesCalls = append(s.getServicesCalls, provider)
	services, ok := s.services[provider]
	if !ok {
		return nil, errors.New("services not yet cached")
	}
	return services, nil
}

func (s *stubCloudInfo) GetRegions(provider, service string) (map[string]string, error) {
	s.getRegionsCalls = append(s.getRegionsCalls, provider+"/"+service)
	regions, ok := s.regions[provider+"/"+service]
	if !ok {
		return nil, errors.New("regions not yet cached")
	}
	return regions, nil
}

func (s *stubCloudInfo) GetProviders() ([]types.Provider, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetProvider(string) (types.Provider, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetZones(string, string, string) ([]string, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetStatus(string) (string, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetProductDetails(string, string, string) ([]types.ProductDetails, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetProductDetail(string, string, string, string) (types.ProductDetails, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetSeriesDetails(string, string, string) (map[string]map[string][]string, []types.SeriesDetails, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetServiceImages(string, string, string) ([]types.Image, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetVersions(string, string, string) ([]types.LocationVersion, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetContinentsData(string, string) (map[string][]types.Region, error) {
	panic("unexpected call")
}

func (s *stubCloudInfo) GetContinents() []string {
	panic("unexpected call")
}
