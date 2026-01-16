// Copyright © 2021 Banzai Cloud
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

package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"logur.dev/logur"

	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/cloudinfoadapter"
)

func TestGceInfoer_priceFromSku(t *testing.T) {
	gceInfoer := GceInfoer{log: cloudinfoadapter.NewLogger(&logur.TestLogger{})}

	tests := []struct {
		name           string
		price          map[string]map[string]map[string]float64
		region         string
		device         string
		priceType      string
		priceInUsd     float64
		expectedPrice  float64
		expectedExists bool
		description    string
	}{
		{
			name:           "set initial non-zero price",
			price:          make(map[string]map[string]map[string]float64),
			region:         "us-east4",
			device:         "m4-memory",
			priceType:      "OnDemand",
			priceInUsd:     0.00457,
			expectedPrice:  0.00457,
			expectedExists: true,
			description:    "should set price when map is empty",
		},
		{
			name: "update existing price with new non-zero value",
			price: map[string]map[string]map[string]float64{
				"us-east4": {
					"m4-memory": {
						"OnDemand": 0.00457,
					},
				},
			},
			region:         "us-east4",
			device:         "m4-memory",
			priceType:      "OnDemand",
			priceInUsd:     0.00567588,
			expectedPrice:  0.00567588,
			expectedExists: true,
			description:    "should overwrite existing price with new non-zero value",
		},
		{
			name: "do not overwrite non-zero price with zero",
			price: map[string]map[string]map[string]float64{
				"us-east4": {
					"m4-memory": {
						"OnDemand": 0.00567588,
					},
				},
			},
			region:         "us-east4",
			device:         "m4-memory",
			priceType:      "OnDemand",
			priceInUsd:     0.0,
			expectedPrice:  0.00567588,
			expectedExists: true,
			description:    "should not overwrite non-zero price with zero",
		},
		{
			name:           "set zero price when no price exists",
			price:          make(map[string]map[string]map[string]float64),
			region:         "us-east4",
			device:         "m4-memory",
			priceType:      "OnDemand",
			priceInUsd:     0.0,
			expectedPrice:  0.0,
			expectedExists: true,
			description:    "should allow setting zero price when no price exists",
		},
		{
			name: "set different price type independently",
			price: map[string]map[string]map[string]float64{
				"us-east4": {
					"m4-memory": {
						"OnDemand": 0.00457,
					},
				},
			},
			region:         "us-east4",
			device:         "m4-memory",
			priceType:      "Preemptible",
			priceInUsd:     0.001828,
			expectedPrice:  0.001828,
			expectedExists: true,
			description:    "should set Preemptible price independently of OnDemand",
		},
		{
			name: "preserve other price types when updating one",
			price: map[string]map[string]map[string]float64{
				"us-east4": {
					"m4-memory": {
						"OnDemand":    0.00457,
						"Preemptible": 0.001828,
					},
				},
			},
			region:         "us-east4",
			device:         "m4-memory",
			priceType:      "OnDemand",
			priceInUsd:     0.00567588,
			expectedPrice:  0.00567588,
			expectedExists: true,
			description:    "should preserve other price types when updating one",
		},
		{
			name: "handle different regions independently",
			price: map[string]map[string]map[string]float64{
				"us-east4": {
					"m4-memory": {
						"OnDemand": 0.00457,
					},
				},
			},
			region:         "us-west1",
			device:         "m4-memory",
			priceType:      "OnDemand",
			priceInUsd:     0.006,
			expectedPrice:  0.006,
			expectedExists: true,
			description:    "should handle different regions independently",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := gceInfoer.priceFromSku(test.price, test.region, test.device, test.priceType, test.priceInUsd)

			// Verify the returned map
			assert.NotNil(t, result, "result should not be nil")
			price, exists := result[test.priceType]
			assert.Equal(t, test.expectedExists, exists, "price existence should match")
			if test.expectedExists {
				assert.Equal(t, test.expectedPrice, price, test.description)
			}

			// Verify the price map was updated correctly
			if test.price[test.region] != nil && test.price[test.region][test.device] != nil {
				storedPrice, storedExists := test.price[test.region][test.device][test.priceType]
				assert.Equal(t, test.expectedExists, storedExists, "stored price existence should match")
				if test.expectedExists {
					assert.Equal(t, test.expectedPrice, storedPrice, "stored price should match")
				}
			}
		})
	}
}
