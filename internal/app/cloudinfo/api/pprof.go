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
	"crypto/subtle"
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

type pprofRoute struct {
	pattern string
	handler http.Handler
}

func pprofRoutes() []pprofRoute {
	return []pprofRoute{
		{pattern: "/debug/pprof/", handler: http.HandlerFunc(pprof.Index)},
		{pattern: "/debug/pprof/cmdline", handler: http.HandlerFunc(pprof.Cmdline)},
		{pattern: "/debug/pprof/profile", handler: http.HandlerFunc(pprof.Profile)},
		{pattern: "/debug/pprof/symbol", handler: http.HandlerFunc(pprof.Symbol)},
		{pattern: "/debug/pprof/trace", handler: http.HandlerFunc(pprof.Trace)},
		{pattern: "/debug/pprof/allocs", handler: pprof.Handler("allocs")},
		{pattern: "/debug/pprof/block", handler: pprof.Handler("block")},
		{pattern: "/debug/pprof/goroutine", handler: pprof.Handler("goroutine")},
		{pattern: "/debug/pprof/heap", handler: pprof.Handler("heap")},
		{pattern: "/debug/pprof/mutex", handler: pprof.Handler("mutex")},
		{pattern: "/debug/pprof/threadcreate", handler: pprof.Handler("threadcreate")},
	}
}

func pprofAuth(secretToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := c.GetHeader("Token")
		if !tokenEquals(provided, secretToken) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

func attachPprof(router gin.IRouter, secretToken string) {
	if secretToken == "" {
		return
	}
	auth := pprofAuth(secretToken)
	for _, route := range pprofRoutes() {
		router.Any(route.pattern, auth, gin.WrapH(route.handler))
	}
}

func tokenEquals(provided, expected string) bool {
	if len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
