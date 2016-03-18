package myhttp

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func startDummySleepServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sleep := req.URL.Query().Get("time")
		errRes := req.URL.Query().Get("err")
		if sleep != "" {
			s, err := strconv.Atoi(sleep)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
			} else {
				time.Sleep(time.Duration(s) * time.Millisecond)
			}
		}

		if errRes != "" {
			http.Error(w, errRes, 500)
			return
		} else {
			return
		}
	}))
}

func doSleepRequest(ctx context.Context, url string, c *Client, sleep string, err string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("time", sleep)
	if err != "" {
		q.Add("err", err)
	}
	req.URL.RawQuery = q.Encode()
	return c.DoRequest(ctx, req)
}

func TestDoResponse(t *testing.T) {

	s := startDummySleepServer()
	defer s.Close()

	c := DefaultClient()
	c.Timeout = 5 * time.Second

	{
		response, err := doSleepRequest(nil, s.URL, c, "10", "")
		if !assert.Nil(t, err) {
			return
		}
		if !assert.Equal(t, response.StatusCode, 200) {
			return
		}
		response.Body.Close()
	}

	{
		_, err := doSleepRequest(nil, s.URL, c, "abc", "")
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.Equal(t, err, ErrHTTPStatusCode) {
			return
		}
	}

}

func TestDoResponseWithRetry(t *testing.T) {

	s := startDummySleepServer()
	defer s.Close()

	c := DefaultClient()
	c.RetryWait = 1 * time.Millisecond
	c.Timeout = 100 * time.Millisecond
	c.Retry = 3

	{
		start := time.Now()
		_, err := doSleepRequest(nil, s.URL, c, "2000", "")
		end := time.Now()
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.NotEqual(t, err, ErrHTTPStatusCode) {
			return
		}
		max := start.Add(583 * time.Millisecond)
		min := start.Add(403 * time.Millisecond)
		if !assert.True(t, end.After(min)) {
			return
		}
		if !assert.True(t, end.Before(max)) {
			return
		}
	}

	c.RetryWait = 100 * time.Millisecond
	c.Timeout = 100 * time.Millisecond
	c.Retry = 3

	{
		start := time.Now()
		_, err := doSleepRequest(nil, s.URL, c, "2000", "")
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.NotEqual(t, err, ErrHTTPStatusCode) {
			return
		}
		max := start.Add(780 * time.Millisecond)
		min := start.Add(700 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
	}

}

func TestWaitResponseWithContext(t *testing.T) {

	s := startDummySleepServer()
	defer s.Close()

	c := DefaultClient()
	c.Timeout = 5 * time.Second
	c.Retry = 3

	{
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 10*time.Millisecond)
		_, err := doSleepRequest(ctx, s.URL, c, "3000", "")
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.Contains(t, err.Error(), "canceled") {
			return
		}
	}

	c = DefaultClient()
	c.RetryWait = 100 * time.Millisecond
	c.Timeout = 100 * time.Millisecond
	c.Retry = 3

	{
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 220*time.Millisecond)
		start := time.Now()
		_, err := doSleepRequest(ctx, s.URL, c, "2000", "")
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.NotEqual(t, err, ErrHTTPStatusCode) {
			return
		}
		max := start.Add(290 * time.Millisecond)
		min := start.Add(220 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
	}

	{
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 320*time.Millisecond)
		start := time.Now()
		_, err := doSleepRequest(ctx, s.URL, c, "2000", "")
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.NotEqual(t, err, ErrHTTPStatusCode) {
			return
		}
		max := start.Add(390 * time.Millisecond)
		min := start.Add(320 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
	}

}
