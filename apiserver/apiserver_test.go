package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rusq/vhoster/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_gateway_handleHealth(t *testing.T) {
	r := httptest.NewRecorder()
	g := &gateway{}
	g.handleHealth(r, nil)
	assert.Equal(t, "OK\n", r.Body.String(), "response mismatch")
}

func TestHandleRemove(t *testing.T) {
	testCases := []struct {
		name       string
		vhost      string
		mockFn     func(mc *mocks.MockHostManager)
		statusCode int
	}{
		{
			name:  "success",
			vhost: "test",
			mockFn: func(mc *mocks.MockHostManager) {
				mc.EXPECT().Exists("test").Return(false)
				mc.EXPECT().Exists("test.example.com").Return(true)
				mc.EXPECT().Remove("test.example.com").Return(nil)
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "bad request",
			vhost:      "",
			mockFn:     func(mc *mocks.MockHostManager) {},
			statusCode: http.StatusBadRequest,
		},
		{
			name:  "not found",
			vhost: "test",
			mockFn: func(mc *mocks.MockHostManager) {
				mc.EXPECT().Exists("test").Return(false)
				mc.EXPECT().Exists("test.example.com").Return(false)
			},
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Create a new HTTP request
			req, err := http.NewRequest(http.MethodDelete, "/vhost/"+tc.vhost, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a new HTTP response recorder
			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			mc := mocks.NewMockHostManager(ctrl)
			tc.mockFn(mc)
			// Create a new gateway with the mock VHoster
			g := &gateway{
				vg:   mc,
				addr: "example.com",
			}

			// Call the handler function
			handler := http.HandlerFunc(g.handleRemove)
			handler.ServeHTTP(rr, req)

			// Check the response status code
			if rr.Code != tc.statusCode {
				t.Errorf("unexpected status code: %d", rr.Code)
			}
		})
	}
}
