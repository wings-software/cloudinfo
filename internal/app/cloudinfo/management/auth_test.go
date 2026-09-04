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

package management

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTokenAuthRouter(token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/management/store")
	group.Use(tokenAuth(token))
	group.GET("export", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return router
}

func TestTokenAuth_RejectsMissingToken(t *testing.T) {
	router := newTokenAuthRouter("secret-token")

	req := httptest.NewRequest(http.MethodGet, "/management/store/export", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTokenAuth_RejectsWrongToken(t *testing.T) {
	router := newTokenAuthRouter("secret-token")

	req := httptest.NewRequest(http.MethodGet, "/management/store/export", nil)
	req.Header.Set("Token", "not-the-secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTokenAuth_AllowsMatchingToken(t *testing.T) {
	router := newTokenAuthRouter("secret-token")

	req := httptest.NewRequest(http.MethodGet, "/management/store/export", nil)
	req.Header.Set("Token", "secret-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}
