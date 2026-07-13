package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type trawlRequest struct {
	Cmd        string `json:"cmd"`
	URL        string `json:"url"`
	MaxTimeout int    `json:"maxTimeout"`
}

type trawlResponse struct {
	Status   string         `json:"status"`
	Message  string         `json:"message"`
	Solution trawlSolution  `json:"solution"`
}

type trawlSolution struct {
	URL      string            `json:"url"`
	Status   int               `json:"status"`
	Headers  map[string]string `json:"headers"`
	Response string            `json:"response"`
}

type trawlTransport struct {
	base     http.RoundTripper
	endpoint string
}

func newTrawlTransport(base http.RoundTripper, endpoint string) *trawlTransport {
	return &trawlTransport{base: base, endpoint: strings.TrimRight(endpoint, "/")}
}

// RoundTrip converts the original GET request into a POST to TRAWL's /v1
// endpoint, parses the FlareSolverr-compatible JSON response, and returns a
// synthetic *http.Response whose body is the solved page HTML.
func (t *trawlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	payload := trawlRequest{
		Cmd:        "request.get",
		URL:        req.URL.String(),
		MaxTimeout: 60000,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("trawl: marshal request: %w", err)
	}

	trawlReq, err := http.NewRequest(http.MethodPost, t.endpoint+"/v1", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("trawl: build request: %w", err)
	}
	trawlReq.Header.Set("Content-Type", "application/json")

	resp, err := t.base.RoundTrip(trawlReq)
	if err != nil {
		return nil, fmt.Errorf("trawl: send: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("trawl: read response: %w", err)
	}

	var tr trawlResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, fmt.Errorf("trawl: parse response: %w", err)
	}
	if tr.Status != "ok" {
		return nil, fmt.Errorf("trawl: %s", tr.Message)
	}
	if tr.Solution.Status < 200 || tr.Solution.Status >= 300 {
		return nil, &StatusError{StatusCode: tr.Solution.Status, URL: req.URL.String()}
	}

	synth := &http.Response{
		Status:        http.StatusText(http.StatusOK),
		StatusCode:    http.StatusOK,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          io.NopCloser(strings.NewReader(tr.Solution.Response)),
		ContentLength: int64(len(tr.Solution.Response)),
		Request:       req,
	}
	synth.Header.Set("Content-Type", "text/html; charset=utf-8")
	return synth, nil
}
