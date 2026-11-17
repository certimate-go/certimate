package resp

import (
	"net/http"

	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/domain"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Ok(e *core.RequestEvent, data interface{}) error {
	rs := &Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
	return e.JSON(http.StatusOK, rs)
}

func Err(e *core.RequestEvent, err error) error {
	code := 500

	xerr, ok := err.(*domain.Error)
	if ok {
		code = xerr.Code
	}

	rs := &Response{
		Code: code,
		Msg:  err.Error(),
		Data: nil,
	}
	return e.JSON(http.StatusOK, rs)
}
