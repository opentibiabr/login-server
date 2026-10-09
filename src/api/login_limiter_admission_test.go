package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type unreadableLoginBody struct{ reads int }

func (body *unreadableLoginBody) Read([]byte) (int, error) {
	body.reads++
	return 0, io.ErrUnexpectedEOF
}
func (*unreadableLoginBody) Close() error { return nil }

func TestLoginRateLimitRejectsExhaustedBucketsBeforeReadingBody(t *testing.T) {
	for _, endpoint := range []string{"/", "/login", "/login.php"} {
		t.Run(endpoint, func(t *testing.T) {
			router, _ := newLimitedLoginRouter(t)
			ip := "203.0.113.9"
			for attempt := 0; attempt < loginTestBurst; attempt++ {
				for _, action := range []string{"login", "news"} {
					assertLimitedLoginStatus(t, limitedLoginRequest(router, endpoint, ip, `{"type":"`+action+`"}`), http.StatusOK)
				}
			}
			unreadableBody := &unreadableLoginBody{}
			largeBody := &countedLoginBody{reader: strings.NewReader(`{"type":"login","email":"` + strings.Repeat("a", 256<<10) + `"}`)}
			for _, test := range []struct {
				name  string
				body  io.ReadCloser
				reads *int
			}{
				{name: "unreadable", body: unreadableBody, reads: &unreadableBody.reads},
				{name: "large", body: largeBody, reads: &largeBody.reads},
			} {
				t.Run(test.name, func(t *testing.T) {
					request := httptest.NewRequest(http.MethodPost, endpoint, nil)
					request.RemoteAddr = ip + ":43210"
					request.Body = test.body
					response := httptest.NewRecorder()
					router.ServeHTTP(response, request)
					assertLimitedLoginStatus(t, response, http.StatusTooManyRequests)
					if *test.reads != 0 {
						t.Fatalf("exhausted request read the body %d times", *test.reads)
					}
				})
			}
		})
	}
}
