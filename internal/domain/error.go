package domain

var (
	ErrInvalidParams    = NewError(400, "invalid params")
	ErrRecordNotFound   = NewError(404, "record not found")
	ErrRevisionConflict = NewError(409, "revision conflict")
)

type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewError(code int, msg string) *Error {
	if code == 0 {
		code = -1
	}

	return &Error{code, msg}
}

func (e *Error) Error() string {
	return e.Msg
}

func IsRecordNotFoundError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == ErrRecordNotFound.Code
	}
	return false
}
