package chessex

import (
	"net/http"
	"testing"
)

type TestService struct {
	s          *Service
	HTTPClient TestHTTPClient
}

type TestHTTPClient struct {
	httpClient http.Client
}

func StartTestService(t *testing.T) *TestService {
	chessex := NewService("", false, false)
	chessex.Cfg = NewDefaultServiceCfg()

	client := TestHTTPClient{}

	ts := TestService{
		s:          chessex,
		HTTPClient: client,
	}

	ts.s.Start()

	return &ts
}

func (ts *TestService) Stop() {
	ts.s.Stop()
}
