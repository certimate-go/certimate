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
	return ErrWithData(e, err, nil)
}

func ErrWithData(e *core.RequestEvent, err error, data interface{}) error {
	code := http.StatusInternalServerError
	if cerr, ok := err.(*domain.Error); ok {
		code = cerr.Code
	}

	if code == 0 {
		panic("unreachable")
	}

	rs := &Response{
		Code: code,
		Msg:  err.Error(),
		Data: data,
	}
	return e.JSON(http.StatusOK, rs)
}
