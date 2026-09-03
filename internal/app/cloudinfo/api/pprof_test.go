// Copyright © 2018 Banzai Cloud
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestPprofAuth_RejectsMissingAndInvalidToken(t *testing.T) {
	router := gin.New()
	router.GET("/debug/pprof/heap", pprofAuth("secret-token"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("wrong token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
		req.Header.Set("Token", "not-the-secret")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestPprofAuth_AllowsMatchingToken(t *testing.T) {
	router := gin.New()
	router.GET("/debug/pprof/heap", pprofAuth("secret-token"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
	req.Header.Set("Token", "secret-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestAttachPprof_ProtectsHeapEndpoint(t *testing.T) {
	router := gin.New()
	attachPprof(router, "secret-token")

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("authenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
		req.Header.Set("Token", "secret-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Body.Bytes())
	})
}

func TestAttachPprof_SkipsHandlersWhenTokenUnset(t *testing.T) {
	router := gin.New()
	attachPprof(router, "")

	for _, route := range router.Routes() {
		assert.NotContains(t, route.Path, "/debug/pprof")
	}

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
	req.Header.Set("Token", "anything")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPprofRoutes(t *testing.T) {
	routes := pprofRoutes()
	require.NotEmpty(t, routes)

	seen := make(map[string]bool, len(routes))
	for _, route := range routes {
		assert.NotEmpty(t, route.pattern)
		assert.NotNil(t, route.handler)
		assert.False(t, seen[route.pattern], "duplicate pprof route %s", route.pattern)
		seen[route.pattern] = true
	}

	assert.True(t, seen["/debug/pprof/heap"])
	assert.True(t, seen["/debug/pprof/goroutine"])
}

func TestTokenEquals(t *testing.T) {
	assert.True(t, tokenEquals("abc", "abc"))
	assert.False(t, tokenEquals("abc", "abd"))
	assert.False(t, tokenEquals("abc", "abcd"))
	assert.False(t, tokenEquals("", "a"))
}
