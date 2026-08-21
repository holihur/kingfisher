package com.kingfisher.common;

import java.util.Map;

/**
 * 错误码定义，与 Go core/errcode 保持一致。
 * 前端 request.ts 强依赖 10104/10003/10105 的语义，切勿随意改动。
 */
public final class ErrorCode {

    private ErrorCode() {}

    public static final int SUCCESS = 0;

    // 通用 10000-10099
    public static final int ERR_INVALID_PARAM = 10001;
    public static final int ERR_UNAUTHORIZED = 10003;
    public static final int ERR_FORBIDDEN = 10004;
    public static final int ERR_NOT_FOUND = 10005;
    public static final int ERR_INTERNAL = 10006;
    public static final int ERR_TOO_MANY_REQUEST = 10007;

    // 用户 10100-10199
    public static final int ERR_USER_EXISTS = 10101;
    public static final int ERR_USER_NOT_FOUND = 10102;
    public static final int ERR_PASSWORD_WRONG = 10103;
    public static final int ERR_TOKEN_EXPIRED = 10104;
    public static final int ERR_TOKEN_INVALID = 10105;
    public static final int ERR_USER_DISABLED = 10106;
    public static final int ERR_LOGIN_FAILED = 10107;
    public static final int ERR_PASSWORD_TOO_SHORT = 10108;
    public static final int ERR_PASSWORD_TOO_LONG = 10109;
    public static final int ERR_PASSWORD_WEAK = 10110;

    // Menu 10200-10299
    public static final int ERR_MENU_EXISTS = 10201;
    public static final int ERR_MENU_NOT_FOUND = 10202;
    public static final int ERR_MENU_HAS_CHILDREN = 10203;

    // Role 10300-10399
    public static final int ERR_ROLE_EXISTS = 10301;
    public static final int ERR_ROLE_NOT_FOUND = 10302;
    public static final int ERR_ROLE_IN_USE = 10303;

    // Config 10400-10499
    public static final int ERR_CONFIG_NOT_FOUND = 10401;
    public static final int ERR_CONFIG_KEY_EXISTS = 10402;

    // Dict 10500-10599
    public static final int ERR_DICT_TYPE_NOT_FOUND = 10501;
    public static final int ERR_DICT_TYPE_CODE_EXISTS = 10502;
    public static final int ERR_DICT_ENTRY_NOT_FOUND = 10503;
    public static final int ERR_DICT_TYPE_HAS_ENTRIES = 10504;
    public static final int ERR_DICT_TYPE_NOT_PUBLIC = 10505;

    // Template 10600-10699
    public static final int ERR_TEMPLATE_NOT_FOUND = 10601;
    public static final int ERR_TEMPLATE_CODE_EXISTS = 10602;

    // Task 10700-10799
    public static final int ERR_TASK_NOT_FOUND = 10701;

    // Doc 10800-10899
    public static final int ERR_DOC_DIR_NOT_FOUND = 10801;
    public static final int ERR_DOC_DIR_HAS_CHILDREN = 10802;
    public static final int ERR_DOC_DIR_HAS_DOCUMENTS = 10803;
    public static final int ERR_DOC_DIR_NOT_VISIBLE = 10804;
    public static final int ERR_DOC_NOT_FOUND = 10805;
    public static final int ERR_DOC_FORBIDDEN = 10806;
    public static final int ERR_DOC_VERSION_NOT_FOUND = 10807;
    public static final int ERR_DOC_VERSION_CONFLICT = 10808;
    public static final int ERR_DOC_CONTENT_INVALID = 10809;

    // Department 10900-10999
    public static final int ERR_DEPT_NOT_FOUND = 10901;
    public static final int ERR_DEPT_HAS_CHILDREN = 10902;

    // Agent 11000-11099
    public static final int ERR_AGENT_DISABLED = 11001;
    public static final int ERR_AGENT_CONVERSATION_NF = 11002;
    public static final int ERR_AGENT_LLM_ERROR = 11003;
    public static final int ERR_AGENT_TOOL_FORBIDDEN = 11004;
    public static final int ERR_AGENT_NO_API_KEY = 11005;

    public static final int ERR_METHOD_NOT_ALLOWED = 10008;
    public static final int ERR_SERVICE_UNAVAILABLE = 10009;
    public static final int ERR_REGISTRATION_DISABLED = 10111;

    private static final Map<Integer, String> MESSAGES = Map.ofEntries(
            Map.entry(SUCCESS, "success"),
            Map.entry(ERR_INVALID_PARAM, "参数错误"),
            Map.entry(ERR_UNAUTHORIZED, "未认证"),
            Map.entry(ERR_FORBIDDEN, "无权限"),
            Map.entry(ERR_NOT_FOUND, "资源不存在"),
            Map.entry(ERR_INTERNAL, "服务器内部错误"),
            Map.entry(ERR_TOO_MANY_REQUEST, "请求过于频繁"),
            Map.entry(ERR_USER_EXISTS, "用户已存在"),
            Map.entry(ERR_USER_NOT_FOUND, "用户不存在"),
            Map.entry(ERR_PASSWORD_WRONG, "密码错误"),
            Map.entry(ERR_TOKEN_EXPIRED, "Token 过期"),
            Map.entry(ERR_TOKEN_INVALID, "Token 无效"),
            Map.entry(ERR_USER_DISABLED, "用户已禁用"),
            Map.entry(ERR_LOGIN_FAILED, "登录失败次数过多"),
            Map.entry(ERR_PASSWORD_TOO_SHORT, "密码过短"),
            Map.entry(ERR_PASSWORD_TOO_LONG, "密码过长"),
            Map.entry(ERR_PASSWORD_WEAK, "密码强度不足"),
            Map.entry(ERR_CONFIG_NOT_FOUND, "配置不存在"),
            Map.entry(ERR_CONFIG_KEY_EXISTS, "配置键已存在"),
            Map.entry(ERR_DICT_TYPE_NOT_FOUND, "字典类型不存在"),
            Map.entry(ERR_DICT_TYPE_CODE_EXISTS, "字典类型编码已存在"),
            Map.entry(ERR_DICT_ENTRY_NOT_FOUND, "字典条目不存在"),
            Map.entry(ERR_DICT_TYPE_HAS_ENTRIES, "字典类型下存在条目，不可删除"),
            Map.entry(ERR_DICT_TYPE_NOT_PUBLIC, "字典类型未公开"),
            Map.entry(ERR_DEPT_NOT_FOUND, "部门不存在"),
            Map.entry(ERR_DEPT_HAS_CHILDREN, "部门下有子部门，不可删除"),
            Map.entry(ERR_AGENT_DISABLED, "Agent 聊天未启用"),
            Map.entry(ERR_AGENT_CONVERSATION_NF, "会话不存在"),
            Map.entry(ERR_AGENT_LLM_ERROR, "LLM 调用失败"),
            Map.entry(ERR_AGENT_TOOL_FORBIDDEN, "无权调用该接口"),
            Map.entry(ERR_AGENT_NO_API_KEY, "未配置 LLM API Key"),
            Map.entry(ERR_MENU_EXISTS, "菜单已存在"),
            Map.entry(ERR_MENU_NOT_FOUND, "菜单不存在"),
            Map.entry(ERR_MENU_HAS_CHILDREN, "菜单有子节点，不可删除"),
            Map.entry(ERR_ROLE_EXISTS, "角色已存在"),
            Map.entry(ERR_ROLE_NOT_FOUND, "角色不存在"),
            Map.entry(ERR_ROLE_IN_USE, "角色被用户使用中"),
            Map.entry(ERR_TEMPLATE_NOT_FOUND, "模版不存在"),
            Map.entry(ERR_TEMPLATE_CODE_EXISTS, "模版编码已存在"),
            Map.entry(ERR_TASK_NOT_FOUND, "周期任务不存在"),
            Map.entry(ERR_DOC_DIR_NOT_FOUND, "文档目录不存在"),
            Map.entry(ERR_DOC_DIR_HAS_CHILDREN, "目录下有子目录，不可删除"),
            Map.entry(ERR_DOC_DIR_HAS_DOCUMENTS, "目录下存在文档，不可删除"),
            Map.entry(ERR_DOC_DIR_NOT_VISIBLE, "目录不可见或无权访问"),
            Map.entry(ERR_DOC_NOT_FOUND, "文档不存在"),
            Map.entry(ERR_DOC_FORBIDDEN, "无权操作该文档"),
            Map.entry(ERR_DOC_VERSION_NOT_FOUND, "文档版本不存在"),
            Map.entry(ERR_DOC_VERSION_CONFLICT, "文档已被他人修改，请刷新后重试"),
            Map.entry(ERR_DOC_CONTENT_INVALID, "文档内容格式不合法"),
            Map.entry(ERR_METHOD_NOT_ALLOWED, "方法不允许"),
            Map.entry(ERR_SERVICE_UNAVAILABLE, "服务暂时不可用"),
            Map.entry(ERR_REGISTRATION_DISABLED, "注册未开放")
    );

    public static String message(int code) {
        return MESSAGES.getOrDefault(code, "未知错误");
    }

    /**
     * 与 Go core/errcode.HTTPStatus 对齐
     */
    public static int httpStatus(int code) {
        if (code == SUCCESS) return 200;
        if (code == ERR_TOO_MANY_REQUEST || code == ERR_LOGIN_FAILED) return 429;
        if (code == ERR_UNAUTHORIZED || code == ERR_TOKEN_EXPIRED || code == ERR_TOKEN_INVALID) return 401;
        if (code == ERR_FORBIDDEN) return 403;
        if (code == ERR_NOT_FOUND) return 404;
        if (code == ERR_INVALID_PARAM || (code >= 10100 && code < 11100)) return 400;
        return 500;
    }
}
