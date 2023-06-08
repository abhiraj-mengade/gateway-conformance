package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ipfs/gateway-conformance/tooling/specs"
)

type SugarTest struct {
	Name      string
	Hint      string
	Request   RequestBuilder
	Requests  []RequestBuilder
	Response  ExpectBuilder
	Responses ExpectsBuilder
}

type SugarTests []SugarTest

func RunIfSpecsAreEnabled(
	t *testing.T,
	tests SugarTests,
	required ...specs.Spec,
) {
	t.Helper()

	missing := []specs.Spec{}
	for _, spec := range required {
		if !spec.IsEnabled() {
			missing = append(missing, spec)
		}
	}

	if len(missing) > 0 {
		t.Skipf("skipping tests, missing specs: %v", missing)
		return
	}

	Run(t, tests)
}

func Run(t *testing.T, tests SugarTests) {
	t.Helper()

	for _, test := range tests {
		timeout, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if len(test.Requests) > 0 {
			t.Run(test.Name, func(t *testing.T) {
				responses := make([]*http.Response, 0, len(test.Requests))

				for _, req := range test.Requests {
					_, res, localReport := runRequest(timeout, t, test, req)
					validateResponse(t, test.Response, res, localReport)
					responses = append(responses, res)
				}

				validateResponses(t, test.Responses, responses)
			})
		} else {
			t.Run(test.Name, func(t *testing.T) {
				_, res, localReport := runRequest(timeout, t, test, test.Request)
				validateResponse(t, test.Response, res, localReport)
			})
		}
	}
}
