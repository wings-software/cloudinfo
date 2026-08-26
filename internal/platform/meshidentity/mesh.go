// Copyright 2026 Harness Inc.
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

// Package meshidentity wires Harness SPIRE mesh identity into cloud-info.
package meshidentity

import (
	"context"
	"net/http"

	sdk "git0.harness.io/l7B_kbSEQD2wjrM7PShm5w/PROD/Harness_Commons/go-sdk/mesh"
)

// ServiceIdentity is the cloud-info SPIFFE / MESH_IDENTITY_AUDIENCE value.
// It is not yet in AuthorizationServiceHeader or go-sdk KnownServiceIDs.
const ServiceIdentity = "CloudInfo"

// Holder is the process mesh handle.
type Holder = sdk.Holder

// Bootstrap loads MESH_IDENTITY_* / SPIFFE_ENDPOINT_SOCKET and returns a holder.
// Stage 0 (both flags false) returns a noop holder with no Workload API connection.
func Bootstrap(ctx context.Context) (*sdk.Holder, error) {
	cfg := sdk.ConfigFromEnv()
	if cfg.Audience == "" {
		cfg.Audience = ServiceIdentity
	}
	sdk.RegisterServiceIDs(ServiceIdentity)
	return sdk.Bootstrap(ctx, cfg, &sdk.BootstrapOptions{
		ExtraServiceIDs: []string{ServiceIdentity},
	})
}

// WrapHandler applies inbound mesh middleware around the HTTP handler.
// Fallback succeeds empty so existing handlers remain authoritative through Stage 0–2.
func WrapHandler(holder *sdk.Holder, next http.Handler) http.Handler {
	if next == nil {
		return next
	}
	fallback := sdk.FallbackAuthFunc(func(_ *http.Request) (sdk.Principal, error) {
		return sdk.Principal{}, nil
	})
	return sdk.Middleware(holder, fallback)(next)
}

// WrapHTTPClient stamps X-Harness-Identity on outbound calls when outbound mesh is enabled.
// targetServiceID must match the callee AuthorizationServiceHeader value.
func WrapHTTPClient(holder *sdk.Holder, client *http.Client, targetServiceID string) {
	if client == nil || holder == nil || !holder.OutboundEnabled() {
		return
	}
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	client.Transport = sdk.NewRoundTripper(holder, sdk.OutboundConfig{
		TargetServiceID: targetServiceID,
	}, base)
}
