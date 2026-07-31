// Package response implements response logic.

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"kingfisher/core/errcode"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type PageData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func OK(data any) *Response {
	return &Response{Code: 0, Message: "success", Data: data}
}

func Page(items any, total int64, page, pageSize int) *Response {
	return &Response{Code: 0, Message: "success", Data: PageData{
		Items: items, Total: total, Page: page, PageSize: pageSize,
	}}
}

func Error(code int) *Response {
	return &Response{Code: code, Message: errcode.Msg(code)}
}

func ErrorWithMsg(code int, msg string) *Response {
	return &Response{Code: code, Message: msg}
}

// JSON helpers — write response to gin.Context
func JSON(c *gin.Context, resp *Response) {
	c.JSON(errcode.HTTPStatus(resp.Code), resp)
}

func AbortJSON(c *gin.Context, resp *Response) {
	c.AbortWithStatusJSON(errcode.HTTPStatus(resp.Code), resp)
}

func OKJSON(c *gin.Context, data any) {
	JSON(c, OK(data))
}

func PageJSON(c *gin.Context, items any, total int64, page, pageSize int) {
	JSON(c, Page(items, total, page, pageSize))
}

func ErrorJSON(c *gin.Context, code int) {
	AbortJSON(c, Error(code))
}

func BadRequest(c *gin.Context, msg string) {
	AbortJSON(c, ErrorWithMsg(errcode.ErrInvalidParam, msg))
}

func Unauthorized(c *gin.Context) {
	AbortJSON(c, Error(errcode.ErrUnauthorized))
}

func Forbidden(c *gin.Context) {
	AbortJSON(c, Error(errcode.ErrForbidden))
}

func NotFound(c *gin.Context) {
	AbortJSON(c, Error(errcode.ErrNotFound))
}

func InternalError(c *gin.Context) {
	AbortJSON(c, Error(errcode.ErrInternal))
}

func TooManyRequest(c *gin.Context) {
	AbortJSON(c, Error(errcode.ErrTooManyRequest))
}

// Ensure http import is used
var _ = http.StatusOK
