package com.kingfisher.common;

import com.kingfisher.modules.agent.service.AgentService;
import com.kingfisher.modules.user.service.UserService;
import com.kingfisher.security.RequirePermAspect;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

/**
 * 全局异常处理，统一转为 ApiResponse。
 */
@RestControllerAdvice
public class GlobalExceptionHandler {

    private static final Logger log = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ApiResponse<Void>> handleValidation(MethodArgumentNotValidException e) {
        String msg = e.getBindingResult().getFieldErrors().stream()
                .map(fe -> fe.getDefaultMessage())
                .findFirst().orElse("参数错误");
        ApiResponse<Void> body = ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, msg);
        return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_INVALID_PARAM)).body(body);
    }

    @ExceptionHandler(UserService.BizException.class)
    public ResponseEntity<ApiResponse<Void>> handleBiz(UserService.BizException e) {
        int code = e.getCode();
        ApiResponse<Void> body = ApiResponse.error(code);
        return ResponseEntity.status(ErrorCode.httpStatus(code)).body(body);
    }

    @ExceptionHandler(AgentService.BizException.class)
    public ResponseEntity<ApiResponse<Void>> handleAgentBiz(AgentService.BizException e) {
        int code = e.getCode();
        ApiResponse<Void> body = ApiResponse.error(code, e.getMessage());
        return ResponseEntity.status(ErrorCode.httpStatus(code)).body(body);
    }

    @ExceptionHandler(RequirePermAspect.ForbiddenException.class)
    public ResponseEntity<ApiResponse<Void>> handleForbidden(RequirePermAspect.ForbiddenException e) {
        log.warn("权限不足: {}", e.getPerm());
        ApiResponse<Void> body = ApiResponse.error(ErrorCode.ERR_FORBIDDEN);
        return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_FORBIDDEN)).body(body);
    }

    @ExceptionHandler(RequirePermAspect.InvalidPermissionException.class)
    public ResponseEntity<ApiResponse<Void>> handleInvalidPerm(RequirePermAspect.InvalidPermissionException e) {
        log.error("权限节点校验失败: {}", e.getMessage());
        ApiResponse<Void> body = ApiResponse.error(ErrorCode.ERR_INTERNAL, e.getMessage());
        return ResponseEntity.status(500).body(body);
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<ApiResponse<Void>> handleIllegalArg(IllegalArgumentException e) {
        ApiResponse<Void> body = ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage());
        return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_INVALID_PARAM)).body(body);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ApiResponse<Void>> handleGeneric(Exception e) {
        log.error("未处理异常", e);
        ApiResponse<Void> body = ApiResponse.error(ErrorCode.ERR_INTERNAL);
        return ResponseEntity.status(500).body(body);
    }
}
