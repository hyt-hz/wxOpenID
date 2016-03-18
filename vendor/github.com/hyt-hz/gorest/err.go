package gorest

import "fmt"

var ECNonRestErr = 0

type RestErr struct {
	ErrCode int    `json:"errcode"`
	ErrMSG  string `json:"errmsg"`
}

func (err *RestErr) Error() string {
	if err.ErrCode == ECNonRestErr {
		return err.ErrMSG
	} else {
		return fmt.Sprintf("REST error %d %s", err.ErrCode, err.ErrMSG)
	}
}
