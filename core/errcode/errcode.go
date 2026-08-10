// Package errcode implements errcode logic.

package errcode

const (
	CodeSuccess = 0

	// Generic 10000-10099
	ErrInvalidParam       = 10001
	ErrUnauthorized       = 10003
	ErrForbidden          = 10004
	ErrNotFound           = 10005
	ErrInternal           = 10006
	ErrTooManyRequest     = 10007
	ErrMethodNotAllowed   = 10008
	ErrServiceUnavailable = 10009

	// User 10100-10199
	ErrUserExists           = 10101
	ErrUserNotFound         = 10102
	ErrPasswordWrong        = 10103
	ErrTokenExpired         = 10104
	ErrTokenInvalid         = 10105
	ErrUserDisabled         = 10106
	ErrLoginFailed          = 10107
	ErrPasswordTooShort     = 10108
	ErrPasswordTooLong      = 10109
	ErrPasswordWeak         = 10110
	ErrRegistrationDisabled = 10111

	// Menu 10200-10299
	ErrMenuExists      = 10201
	ErrMenuNotFound    = 10202
	ErrMenuHasChildren = 10203

	// Role 10300-10399
	ErrRoleExists   = 10301
	ErrRoleNotFound = 10302
	ErrRoleInUse    = 10303

	// Config 10400-10499
	ErrConfigNotFound  = 10401
	ErrConfigKeyExists = 10402

	// Dict 10500-10599
	ErrDictTypeNotFound   = 10501
	ErrDictTypeCodeExists = 10502
	ErrDictEntryNotFound  = 10503
	ErrDictTypeHasEntries = 10504
	ErrDictTypeNotPublic  = 10505

	// Template 10600-10699
	ErrTemplateNotFound   = 10601
	ErrTemplateCodeExists = 10602

	// Task 10700-10799
	ErrTaskNotFound = 10701
)

//nolint:gosec // false positive — Chinese error messages
var errMsg = map[int]string{
	CodeSuccess:             "success",
	ErrInvalidParam:         "参数错误",
	ErrUnauthorized:         "未认证",
	ErrForbidden:            "无权限",
	ErrNotFound:             "资源不存在",
	ErrInternal:             "服务器内部错误",
	ErrTooManyRequest:       "请求过于频繁",
	ErrMethodNotAllowed:     "方法不允许",
	ErrServiceUnavailable:   "服务暂时不可用",
	ErrUserExists:           "用户已存在",
	ErrUserNotFound:         "用户不存在",
	ErrPasswordWrong:        "密码错误",
	ErrTokenExpired:         "Token 过期",
	ErrTokenInvalid:         "Token 无效",
	ErrUserDisabled:         "用户已禁用",
	ErrLoginFailed:          "登录失败次数过多",
	ErrPasswordTooShort:     "密码过短",
	ErrPasswordTooLong:      "密码过长",
	ErrPasswordWeak:         "密码强度不足",
	ErrRegistrationDisabled: "注册未开放",
	ErrMenuExists:           "菜单已存在",
	ErrMenuNotFound:         "菜单不存在",
	ErrMenuHasChildren:      "菜单有子节点，不可删除",
	ErrRoleExists:           "角色已存在",
	ErrRoleNotFound:         "角色不存在",
	ErrRoleInUse:            "角色被用户使用中",
	ErrConfigNotFound:       "配置不存在",
	ErrConfigKeyExists:      "配置键已存在",
	ErrDictTypeNotFound:     "字典类型不存在",
	ErrDictTypeCodeExists:   "字典类型编码已存在",
	ErrDictEntryNotFound:    "字典条目不存在",
	ErrDictTypeHasEntries:   "字典类型下存在条目，不可删除",
	ErrDictTypeNotPublic:    "字典类型未公开",
	ErrTemplateNotFound:     "模版不存在",
	ErrTemplateCodeExists:   "模版编码已存在",
	ErrTaskNotFound:         "周期任务不存在",
}

func Msg(code int) string {
	if m, ok := errMsg[code]; ok {
		return m
	}
	return "未知错误"
}

func HTTPStatus(code int) int {
	switch {
	case code == 0:
		return 200
	case code == 10007 || code == 10107:
		return 429
	case code == 10009:
		return 503
	case code == 10003 || code == 10104 || code == 10105:
		return 401
	case code == 10004:
		return 403
	case code == 10005:
		return 404
	case code == 10008:
		return 405
	case code == 10001 || (code >= 10100 && code < 10700):
		return 400
	default:
		return 500
	}
}
