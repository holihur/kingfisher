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

	// Doc 10800-10899
	ErrDocDirNotFound     = 10801
	ErrDocDirHasChildren  = 10802
	ErrDocDirHasDocuments = 10803
	ErrDocDirNotVisible   = 10804
	ErrDocNotFound        = 10805
	ErrDocForbidden       = 10806
	ErrDocVersionNotFound = 10807
	ErrDocVersionConflict = 10808
	ErrDocContentInvalid  = 10809

	// Department 10900-10999
	ErrDeptNotFound    = 10901
	ErrDeptHasChildren = 10902

	// Agent 11000-11099
	ErrAgentDisabled       = 11001
	ErrAgentConversationNF = 11002
	ErrAgentLLMError       = 11003
	ErrAgentToolForbidden  = 11004
	ErrAgentNoAPIKey       = 11005

	// SubAccount 11100-11199
	ErrSubAccountNotFound    = 11101
	ErrSubAccountLimit       = 11102
	ErrSubAccountNoPerm      = 11103
	ErrSubAccountIsSub       = 11104
	ErrSubAccountParentNotFound = 11105
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
	ErrDocDirNotFound:       "文档目录不存在",
	ErrDocDirHasChildren:    "目录下有子目录，不可删除",
	ErrDocDirHasDocuments:   "目录下存在文档，不可删除",
	ErrDocDirNotVisible:     "目录不可见或无权访问",
	ErrDocNotFound:          "文档不存在",
	ErrDocForbidden:         "无权操作该文档",
	ErrDocVersionNotFound:   "文档版本不存在",
	ErrDeptNotFound:         "部门不存在",
	ErrDeptHasChildren:      "部门下有子部门，不可删除",
	ErrDocVersionConflict:   "文档已被他人修改，请刷新后重试",
	ErrDocContentInvalid:    "文档内容格式不合法",
	ErrAgentDisabled:        "Agent 聊天未启用",
	ErrAgentConversationNF:  "会话不存在",
	ErrAgentLLMError:        "LLM 调用失败",
	ErrAgentToolForbidden:   "无权调用该接口",
	ErrAgentNoAPIKey:        "未配置 LLM API Key",
	ErrSubAccountNotFound:       "子账户不存在",
	ErrSubAccountLimit:          "子账户数量已达上限",
	ErrSubAccountNoPerm:         "子账户权限超出父账户范围",
	ErrSubAccountIsSub:          "子账户不能再创建子账户",
	ErrSubAccountParentNotFound: "父账户不存在",
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
	case code == 10004 || code == ErrDocForbidden:
		return 403
	case code == 10005 || code == ErrDocNotFound || code == ErrDocDirNotFound || code == ErrDocDirNotVisible || code == ErrDocVersionNotFound:
		return 404
	case code == 10008:
		return 405
	case code == 10001 || (code >= 10100 && code < 11100):
		return 400
	default:
		return 500
	}
}
