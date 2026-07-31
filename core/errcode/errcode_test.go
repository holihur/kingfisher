package errcode

import "testing"

func TestHTTPStatus(t *testing.T) {
	tests := []struct{ code, want int }{
		{0, 200}, {10001, 400}, {10007, 429}, {10107, 429}, {10009, 503},
		{10003, 401}, {10104, 401}, {10105, 401},
		{10004, 403}, {10005, 404}, {10008, 405}, {99999, 500},
	}
	for _, tt := range tests {
		if got := HTTPStatus(tt.code); got != tt.want {
			t.Errorf("HTTPStatus(%d)=%d, want %d", tt.code, got, tt.want)
		}
	}
}

func TestMsg(t *testing.T) {
	if Msg(0) != "success" { t.Error("Msg(0) should be success") }
	if Msg(10101) != "用户已存在" { t.Error("Msg(10101) wrong") }
	if Msg(99999) != "未知错误" { t.Error("Msg(99999) should be unknown") }
}
