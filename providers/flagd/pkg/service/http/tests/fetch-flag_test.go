package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	models "github.com/open-feature/flagd/pkg/model"
	service "github.com/open-feature/golang-sdk-contrib/providers/flagd/pkg/service/http"
	mocks "github.com/open-feature/golang-sdk-contrib/providers/flagd/pkg/service/http/tests/mocks"
	of "github.com/open-feature/golang-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	schemav1 "go.buf.build/grpc/go/open-feature/flagd/schema/v1"
)

type TestFetchFlagArgs struct {
	name                          string
	serviceClientMockRequestSetup mocks.ServiceClientMockRequestSetup
	body                          interface{}
	url                           string
	ctx                           of.EvaluationContext
	statusCode                    int
	err                           error
}

func TestFetchFlag(t *testing.T) {
	tests := []TestFetchFlagArgs{
		{
			name: "happy path",
			serviceClientMockRequestSetup: mocks.ServiceClientMockRequestSetup{
				InMethod: "POST",
				InUrl:    "GET/MY/FLAG",
				OutRes:   &http.Response{},
				OutErr:   nil,
			},
			body: map[string]interface{}{
				"food": "bars",
			},
			url: "GET/MY/FLAG",
			ctx: map[string]interface{}{
				"con": "text",
			},
			statusCode: 200,
			err:        nil,
		},
		{
			name: "200 response cannot unmarshal",
			serviceClientMockRequestSetup: mocks.ServiceClientMockRequestSetup{
				InMethod: "POST",
				InUrl:    "GET/MY/FLAG",
				OutRes:   &http.Response{},
				OutErr:   nil,
			},
			body: "string",
			url:  "GET/MY/FLAG",
			ctx: map[string]interface{}{
				"con": "text",
			},
			statusCode: 200,
			err:        errors.New(models.ParseErrorCode),
		},
		{
			name: "non 200 response cannot unmarshal",
			serviceClientMockRequestSetup: mocks.ServiceClientMockRequestSetup{
				InMethod: "POST",
				InUrl:    "GET/MY/FLAG",
				OutRes:   &http.Response{},
				OutErr:   nil,
			},
			body: "string",
			url:  "GET/MY/FLAG",
			ctx: map[string]interface{}{
				"con": "text",
			},
			statusCode: 400,
			err:        errors.New(models.ParseErrorCode),
		},
		{
			name: "non 200 response",
			serviceClientMockRequestSetup: mocks.ServiceClientMockRequestSetup{
				InMethod: "POST",
				InUrl:    "GET/MY/FLAG",
				OutRes:   &http.Response{},
				OutErr:   nil,
			},
			body: schemav1.ErrorResponse{
				ErrorCode: models.FlagNotFoundErrorCode,
			},
			url: "GET/MY/FLAG",
			ctx: map[string]interface{}{
				"con": "text",
			},
			statusCode: 404,
			err:        errors.New(models.FlagNotFoundErrorCode),
		},
		{
			name: "500 response",
			serviceClientMockRequestSetup: mocks.ServiceClientMockRequestSetup{
				InMethod: "POST",
				InUrl:    "GET/MY/FLAG",
				OutRes:   &http.Response{},
				OutErr:   nil,
			},
			url: "GET/MY/FLAG",
			ctx: map[string]interface{}{
				"con": "text",
			},
			statusCode: 500,
			err:        errors.New(models.GeneralErrorCode),
		},
		{
			name: "ctx cannot be parsed",
			serviceClientMockRequestSetup: mocks.ServiceClientMockRequestSetup{
				InMethod: "POST",
				InUrl:    "GET/MY/FLAG",
				OutRes:   &http.Response{},
				OutErr:   nil,
			},
			url:        "GET/MY/FLAG",
			ctx:        make(chan error, 5),
			statusCode: 500,
			err:        errors.New(models.ParseErrorCode),
		},
		{
			name: "fall through",
			serviceClientMockRequestSetup: mocks.ServiceClientMockRequestSetup{
				InMethod: "POST",
				InUrl:    "GET/MY/FLAG",
				OutRes:   &http.Response{},
				OutErr:   nil,
			},
			body: schemav1.ErrorResponse{
				ErrorCode: "",
			},
			url:        "GET/MY/FLAG",
			statusCode: 400,
			err:        errors.New(models.GeneralErrorCode),
		},
	}
	for _, test := range tests {
		bodyM, err := json.Marshal(test.body)
		if err != nil {
			t.Error(err)
		}
		test.serviceClientMockRequestSetup.OutRes = &http.Response{
			StatusCode: test.statusCode,
			Body:       io.NopCloser(bytes.NewReader(bodyM)),
		}
		svc := service.HTTPService{
			Client: &mocks.ServiceClient{
				RequestSetup: test.serviceClientMockRequestSetup,
				Testing:      t,
			},
		}
		target := map[string]interface{}{}
		err = svc.FetchFlag(test.url, test.ctx, &target)

		if test.err != nil && !assert.EqualError(t, err, test.err.Error()) {
			t.Errorf("%s: unexpected value for error expected %v recieved %v", test.name, test.err, err)
		}
	}
}
