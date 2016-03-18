package gorest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	RTCS_HTTP_ERROR_CODE = 551
)

func WriteJsonResponse(w http.ResponseWriter, response interface{}) {

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %+v", err)
	}
}

func ErrHTTP(w http.ResponseWriter, r *http.Request, msg string, a ...interface{}) {

	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	http.Error(w, msg, RTCS_HTTP_ERROR_CODE)
	log.Printf("%s %s %s", r.Method, r.URL, msg)
}

func ErrHTTPErrJson(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	rtcsErr, ok := err.(*RestErr)
	if !ok {
		rtcsErr = &RestErr{ErrCode: ECNonRestErr, ErrMSG: err.Error()}
	}
	log.Printf("%s %s: %s", r.Method, r.URL, rtcsErr.Error())
	if err := json.NewEncoder(w).Encode(rtcsErr); err != nil {
		log.Printf("Failed to encode response: %+v", err)
	}
}

func ErrHTTPStrJson(w http.ResponseWriter, r *http.Request, msg string, a ...interface{}) {
	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	rtcsErr := &RestErr{ErrCode: ECNonRestErr, ErrMSG: msg}
	log.Printf("%s %s: %s", r.Method, r.URL, rtcsErr.Error())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rtcsErr); err != nil {
		log.Printf("Failed to encode RTCS error response: %+v", err)
	}
}

func ErrHTTPRTCSErrJson(w http.ResponseWriter, r *http.Request, rtcsErr *RestErr) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("%s %s: %s", r.Method, r.URL, rtcsErr.Error())
	if err := json.NewEncoder(w).Encode(rtcsErr); err != nil {
		log.Printf("Failed to encode RTCS error response: %+v", err)
	}
}
