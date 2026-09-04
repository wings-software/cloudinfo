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
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/banzaicloud/cloudinfo/internal/app/cloudinfo/problems"
	"github.com/banzaicloud/cloudinfo/internal/cloudinfo"
)

// Responder marks responders
type Responder interface {
	// Respond implements the responding logic / it's intended to be self-contained
	Respond(ginCtx *gin.Context, err error)
}

// errorResponder struct in charge for assembling classified error responses
type errorResponder struct {
	errClassifier Classifier
	logger        cloudinfo.Logger
}

// Respond assembles the error response corresponding to the passed in error
func (er *errorResponder) Respond(ginCtx *gin.Context, err error) {
	if responseData, e := er.errClassifier.Classify(err); e == nil {
		er.respond(ginCtx, responseData)
		return
	}

	// The concrete error is masked from the client to avoid leaking internal
	// details, so it must be logged here to remain diagnosable.
	if er.logger != nil {
		er.logger.Error("unhandled internal error", map[string]interface{}{"error": err.Error()})
	}

	er.respond(ginCtx, problems.NewUnknownProblem(err))
}

// respond sets the response in the gin context
func (er *errorResponder) respond(ginCtx *gin.Context, d interface{}) {
	if problems.IsDefaultProblem(d) {
		ginCtx.AbortWithStatusJSON(problems.ProblemStatus(d), d)
		return
	}

	ginCtx.AbortWithStatusJSON(http.StatusInternalServerError, problems.NewUnknownProblem(d))
}

// NewErrorResponder configures a new error responder
func NewErrorResponder(logger cloudinfo.Logger) Responder {
	return &errorResponder{
		errClassifier: NewErrorClassifier(),
		logger:        logger,
	}
}
