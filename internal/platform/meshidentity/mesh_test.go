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

package meshidentity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "git0.harness.io/l7B_kbSEQD2wjrM7PShm5w/PROD/Harness_Commons/go-sdk/mesh"
	"github.com/stretchr/testify/require"
)

func TestBootstrapStageZero(t *testing.T) {
	t.Setenv("MESH_IDENTITY_INBOUND_ENABLED", "false")
	t.Setenv("MESH_IDENTITY_OUTBOUND_ENABLED", "false")
	t.Setenv("MESH_IDENTITY_AUDIENCE", "")

	holder, err := Bootstrap(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, holder.Close())
	})

	require.True(t, holder.IsNoop())
	require.False(t, holder.InboundEnabled())
	require.False(t, holder.OutboundEnabled())
	require.Equal(t, ServiceIdentity, holder.Config().Audience)
	require.Equal(t, "CloudInfo", ServiceIdentity)
	require.True(t, sdk.IsKnownServiceID(ServiceIdentity))
}

func TestWrapHandlerStageZeroPassesThrough(t *testing.T) {
	holder, err := sdk.Bootstrap(context.Background(), sdk.Config{}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, holder.Close()) })

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})
	handler := WrapHandler(holder, next)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/status", nil))
	require.True(t, called)
	require.Equal(t, http.StatusTeapot, rr.Code)
}

func TestWrapHandlerNilNext(t *testing.T) {
	require.Nil(t, WrapHandler(nil, nil))
}

func TestWrapHandlerDefersToExistingAuthorization(t *testing.T) {
	holder, err := sdk.Bootstrap(context.Background(), sdk.Config{
		InboundEnabled:  true,
		FallbackEnabled: true,
		Audience:        ServiceIdentity,
	}, &sdk.BootstrapOptions{
		Source:          &sdk.StaticSource{},
		DisableMetrics:  true,
		ExtraServiceIDs: []string{ServiceIdentity},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, holder.Close())
	})

	existingAuth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := WrapHandler(holder, existingAuth)

	t.Run("legacy authorization remains accepted after mesh validation fails", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set(sdk.IdentityHeader, "invalid-mesh-token")
		request.Header.Set("Authorization", "Bearer existing-jwt")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("existing authorization still rejects unauthenticated requests", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set(sdk.IdentityHeader, "invalid-mesh-token")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusUnauthorized, response.Code)
	})
}

func TestWrapHTTPClientNoopWhenOutboundDisabled(t *testing.T) {
	holder, err := sdk.Bootstrap(context.Background(), sdk.Config{}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, holder.Close()) })

	base := &http.Transport{}
	client := &http.Client{Transport: base}
	WrapHTTPClient(holder, client, sdk.ServiceNextGenManager)
	require.Equal(t, base, client.Transport)
}

func TestWrapHTTPClientNilSafe(t *testing.T) {
	WrapHTTPClient(nil, nil, sdk.ServiceNextGenManager)
	WrapHTTPClient(nil, &http.Client{}, sdk.ServiceNextGenManager)
}
