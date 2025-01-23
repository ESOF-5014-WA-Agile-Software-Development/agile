package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mylakehead/agile/code"
	"github.com/mylakehead/agile/runtime"
)

type Payload struct {
	Code    code.Code     `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`
}

type Error struct {
	Status  int
	Payload *Payload
}

type Handler func(rt *runtime.Runtime, gCtx *gin.Context) (interface{}, *Error)

type DataType int

const (
	DataTypeJson     DataType = iota // exactly JSON and already JSON header set
	DataTypeJsonStr                  // exactly JSON but with no JSON header set
	DataTypePlainStr                 // plain text
)

type WrapConfig struct {
	RespDataType DataType
}

func WithDataType(t DataType) func(config *WrapConfig) {
	return func(w *WrapConfig) {
		w.RespDataType = t
	}
}

// Wrap
// success:               status - 200
// bad request:           status - 400
// unauthorized:          status - 401
// forbidden:             status - 403
// internal server error: status - 500
func Wrap(h Handler, rt *runtime.Runtime, opts ...func(*WrapConfig)) gin.HandlerFunc {
	w := &WrapConfig{
		RespDataType: DataTypeJson,
	}

	for _, opt := range opts {
		opt(w)
	}

	return func(gCtx *gin.Context) {
		data, err := h(rt, gCtx)

		if err != nil {
			if err.Status == http.StatusInternalServerError {
				gCtx.String(http.StatusInternalServerError, err.Payload.Message)
			} else if err.Status == http.StatusSeeOther {
				gCtx.Redirect(http.StatusSeeOther, data.(string))
			} else {
				gCtx.JSON(err.Status, err.Payload)
			}
		} else {
			switch w.RespDataType {
			case DataTypeJson:
				gCtx.JSON(http.StatusOK, data)
			case DataTypeJsonStr:
				gCtx.Header("Content-Type", "application/json; charset=utf-8")
				gCtx.String(http.StatusOK, data.(string))
			case DataTypePlainStr:
				gCtx.String(http.StatusOK, data.(string))
			default:
				gCtx.JSON(http.StatusOK, data)
			}
		}
	}
}

func Redirect(status int) *Error {
	return &Error{
		Status: status,
	}
}

func InvalidArgument(details []interface{}, messages ...string) *Error {
	if len(messages) > 0 {
		return &Error{
			Status: http.StatusBadRequest,
			Payload: &Payload{
				Code:    code.InvalidArgument,
				Message: messages[0],
				Details: details,
			},
		}
	}

	return &Error{
		Status: http.StatusBadRequest,
		Payload: &Payload{
			Code:    code.InvalidArgument,
			Message: code.InvalidArgument.String(),
			Details: details,
		},
	}
}

func InternalServerError(messages ...string) *Error {
	if len(messages) > 0 {
		return &Error{
			Status: http.StatusInternalServerError,
			Payload: &Payload{
				Message: messages[0],
			},
		}
	}

	return &Error{
		Status: http.StatusInternalServerError,
		Payload: &Payload{
			Message: "500 Internal Server Error",
		},
	}
}

func NotFoundError(messages ...string) *Error {
	if len(messages) > 0 {
		return &Error{
			Status: http.StatusBadRequest,
			Payload: &Payload{
				Code:    code.NotFoundError,
				Message: messages[0],
			},
		}
	}

	return &Error{
		Status: http.StatusBadRequest,
		Payload: &Payload{
			Code:    code.NotFoundError,
			Message: code.NotFoundError.String(),
		},
	}
}
