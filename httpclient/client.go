package myhttp

import (
	"errors"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	ErrHTTPStatusCode = errors.New("Not 200 OK response")
)

const (
	DefaultMaxIdleConnsPerHost = 50
	DefaultRequestTimeout      = 10 * time.Second
	DefaultRequestRetryCnt     = 3
	DefaultRequestRetryWait    = 10 * time.Second
)

type roundTripperWithRequestCanceler interface {
	http.RoundTripper
	CancelRequest(*http.Request)
}

// HTTP client with following features:
// 1. supports context with timeout/cancel
// 2. supports request timeout (thanks to http.Client)
// 3. supports retry, but only retry in case of timeout/connection error, will never retry on non-200 OK response
// 4. supports HTTP keep-alive and connection pool (thanks to http.Transport)
//
type Client struct {
	http.Client
	Transport roundTripperWithRequestCanceler
	RetryWait time.Duration
	Retry     int
}

func DefaultClient() *Client {
	return NewClient(DefaultMaxIdleConnsPerHost, DefaultRequestTimeout, DefaultRequestRetryCnt, DefaultRequestRetryWait)
}

func NewClient(idleConns int, timeout time.Duration, retry int, retryWait time.Duration) *Client {

	tr := &http.Transport{
		MaxIdleConnsPerHost: idleConns,
	}
	c := &Client{
		Client: http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
		Transport: tr,
		RetryWait: retryWait,
		Retry:     retry,
	}

	return c
}

// for unit test purpose
func (c *Client) SetTransport(tr roundTripperWithRequestCanceler) {
	c.Transport = tr
	c.Client.Transport = tr
}

func (c *Client) SetPoolSize(idleConns int) {
	c.Transport = &http.Transport{
		MaxIdleConnsPerHost: idleConns,
	}
	c.Client.Transport = c.Transport
}

func (c *Client) DoRequest(ctx context.Context, req *http.Request) (response *http.Response, err error) {
	retry := c.Retry

	if ctx != nil {
		reqDoneCh := make(chan struct{})
		go func() {
		for_loop:
			for {
				response, err = c.Do(req)
				if err != nil && retry > 0 && ctx.Err() == nil {
					log.Printf("HTTP request failed, wait for retry: %s", err)
					retry -= 1
					select {
					case <-time.After(c.RetryWait):
						continue
					case <-ctx.Done():
						break for_loop
					}
					continue
				} else {
					break for_loop
				}
			}
			close(reqDoneCh)
		}()

		select {
		case <-ctx.Done():
			c.Transport.CancelRequest(req)
			<-reqDoneCh
			break
		case <-reqDoneCh:
			break
		}
	} else {
		for {
			response, err = c.Do(req)
			if err != nil && retry > 0 {
				log.Printf("HTTP request failed, wait for retry: %s", err)
				retry -= 1
				time.Sleep(c.RetryWait)
				continue
			} else {
				break
			}
		}
	}

	// check HTTP status code
	if err == nil {
		if response.StatusCode != 200 {
			if response.ContentLength != 0 {
				var errMsg []byte
				errMsg, err = ioutil.ReadAll(response.Body)
				if err != nil {
					log.Printf("Failed to read response error message body: %s", err)
				} else {
					log.Printf("HTTP request failed: %s %s: %s %s",
						response.Status,
						string(strings.TrimSpace(string(errMsg))),
						req.Method,
						req.URL.String(),
					)
					err = ErrHTTPStatusCode
				}
			} else {
				log.Printf("HTTP request failed: %s", response.Status)
				err = ErrHTTPStatusCode
			}
			response.Body.Close()
		}
	}

	return
}
