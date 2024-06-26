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

package amazon

import (
	"strings"
)

// mapSeries get instance series associated with the instanceType
func (e *Ec2Infoer) mapSeries(instanceType string) string {
	seriesParts := strings.Split(instanceType, ".")

	if len(seriesParts) == 2 {
		return seriesParts[0]
	}

	e.log.Warn("error parsing instance series from instanceType", map[string]interface{}{"instanceType": instanceType})

	// return instanceType itself when there is a parsing error, to speedup debugging
	return instanceType
}
