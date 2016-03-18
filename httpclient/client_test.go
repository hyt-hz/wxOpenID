package myhttp

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"testing"
	"time"
)

type transport struct {
	RT func(*http.Request) (*http.Response, error)
	CR func(*http.Request)
}

func (tr *transport) RoundTrip(request *http.Request) (*http.Response, error) {
	return tr.RT(request)
}

func (tr *transport) CancelRequest(request *http.Request) {
	tr.CR(request)
}

type body struct {
	read  func(p []byte) (n int, err error)
	close func() error
}

func (b *body) Read(p []byte) (n int, err error) {
	return b.read(p)
}

func (b *body) Close() error {
	return b.close()
}

func newBody(read func(p []byte) (n int, err error), close func() error) *body {
	return &body{
		read:  read,
		close: close,
	}
}

func newResponse(statusCode int, read func(p []byte) (n int, err error), close func() error) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       newBody(read, close),
	}
}

func setTr(c *Client, rt func(*http.Request) (*http.Response, error), cr func(*http.Request)) {
	tr := &transport{
		RT: rt,
		CR: cr,
	}

	c.SetTransport(tr)
}

func newTestClient(cancelCnt *int, sleep time.Duration) *Client {
	c := DefaultClient()
	cancelRequestChan := make(chan int, 10)
	setTr(
		c,
		func(request *http.Request) (*http.Response, error) {
			select {
			case <-time.After(sleep):
				return newResponse(
					200,
					func(p []byte) (n int, err error) {
						return 0, nil
					},
					func() error {
						return nil
					},
				), nil
			case <-cancelRequestChan:
				return nil, errors.New("")
			}
		},
		func(request *http.Request) {
			*cancelCnt += 1
			cancelRequestChan <- 1
		},
	)

	return c
}

func TestDoRequest(t *testing.T) {
	// test 200 OK
	cancelCnt := 0
	c := newTestClient(&cancelCnt, 10*time.Millisecond)
	req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
	response, err := c.DoRequest(nil, req)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.Equal(t, response.StatusCode, 200) {
		return
	}
	response.Body.Close()
}

func TestDoRequestTimeout(t *testing.T) {

	{
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 10*time.Millisecond)
		c.Timeout = 15 * time.Millisecond
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		response, err := c.DoRequest(nil, req)
		if !assert.Nil(t, err) {
			return
		}
		if !assert.Equal(t, response.StatusCode, 200) {
			return
		}
		if !assert.Equal(t, cancelCnt, 0) {
			return
		}
		response.Body.Close()
	}

	{
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 1000*time.Millisecond)
		c.Timeout = 100 * time.Millisecond
		c.RetryWait = 1 * time.Microsecond
		c.Retry = 3
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		start := time.Now()
		_, err := c.DoRequest(nil, req)
		max := start.Add(480 * time.Millisecond)
		min := start.Add(400 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.Equal(t, cancelCnt, 4) {
			return
		}
	}

	{
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 1000*time.Millisecond)
		c.Timeout = 100 * time.Millisecond
		c.RetryWait = 100 * time.Millisecond
		c.Retry = 3
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		start := time.Now()
		_, err := c.DoRequest(nil, req)
		max := start.Add(780 * time.Millisecond)
		min := start.Add(700 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.Equal(t, cancelCnt, 4) {
			return
		}
	}

}

func TestDoRequestContext(t *testing.T) {

	{
		// no timeout
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 10*time.Millisecond)
		ctx := context.Background()
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		response, err := c.DoRequest(ctx, req)
		if !assert.Nil(t, err) {
			return
		}
		if !assert.Equal(t, response.StatusCode, 200) {
			return
		}
		if !assert.Equal(t, cancelCnt, 0) {
			return
		}
		response.Body.Close()
	}

	{
		// context timeout not effective
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 10*time.Millisecond)
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 1*time.Second)
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		response, err := c.DoRequest(ctx, req)
		if !assert.Nil(t, err) {
			return
		}
		if !assert.Equal(t, response.StatusCode, 200) {
			return
		}
		if !assert.Equal(t, cancelCnt, 0) {
			return
		}
		response.Body.Close()
	}

	{
		// context timeout
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 100*time.Millisecond)
		c.Timeout = 10 * time.Second
		c.RetryWait = 10 * time.Second
		c.Retry = 3
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 50*time.Millisecond)
		start := time.Now()
		_, err := c.DoRequest(ctx, req)
		max := start.Add(70 * time.Millisecond)
		min := start.Add(50 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
		if !assert.NotNil(t, err) {
			return
		}
		if !assert.Equal(t, cancelCnt, 1) {
			return
		}
	}

	{
		// several client timeout before context timeout
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 200*time.Millisecond)
		c.Timeout = 100 * time.Millisecond
		c.RetryWait = 100 * time.Millisecond
		c.Retry = 5
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 640*time.Millisecond)
		start := time.Now()
		_, err := c.DoRequest(ctx, req)
		max := start.Add(690 * time.Millisecond)
		min := start.Add(640 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
		if !assert.NotNil(t, err) {
			return
		}
	}

	{
		// several client timeout before context timeout
		cancelCnt := 0
		c := newTestClient(&cancelCnt, 200*time.Millisecond)
		c.Timeout = 100 * time.Millisecond
		c.RetryWait = 100 * time.Millisecond
		c.Retry = 5
		req, _ := http.NewRequest("GET", "http://127.0.0.1/abc", nil)
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 740*time.Millisecond)
		start := time.Now()
		_, err := c.DoRequest(ctx, req)
		max := start.Add(790 * time.Millisecond)
		min := start.Add(740 * time.Millisecond)
		if !assert.True(t, time.Now().After(min)) {
			return
		}
		if !assert.True(t, time.Now().Before(max)) {
			return
		}
		if !assert.NotNil(t, err) {
			return
		}
	}

}
