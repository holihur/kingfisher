package com.kingfisher.security;

import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.rbac.domain.Permission;
import com.kingfisher.modules.rbac.service.PermissionService;
import com.kingfisher.modules.rbac.service.RoleService;
import jakarta.servlet.http.HttpServletRequest;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import java.util.List;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

/**
 * 权限校验切面：对 {@link RequirePerm} 注解进行双重校验
 * 1. 节点有效性校验（格式 + 是否在 permissions 表中定义）
 * 2. 用户是否拥有该权限（原有逻辑）
 *
 * 节点不存在时抛出 {@link InvalidPermissionException}，由全局异常处理器转为 500（ErrInternal），
 * 避免将错误配置静默为 403 而掩盖拼写错误/配置遗漏。
 */
@Aspect
@Component
public class RequirePermAspect {

    private static final Logger log = LoggerFactory.getLogger(RequirePermAspect.class);

    /**
     * 权限码格式：resource:action，仅允许小写字母、数字、下划线，中间以冒号分隔
     * 与 Go 版本 database.SeedData 中的 Code 风格保持一致（如 user:list、worktask:create）
     */
    private static final Pattern PERM_PATTERN = Pattern.compile("^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$");

    private final RoleService roleService;
    private final PermissionService permissionService;

    /** 已定义权限码缓存（permissions 表全量），惰性加载后常驻内存 */
    private volatile Set<String> validPermsCache;
    private final Object cacheLock = new Object();

    public RequirePermAspect(RoleService roleService, PermissionService permissionService) {
        this.roleService = roleService;
        this.permissionService = permissionService;
    }

    @Around("@annotation(requirePerm)")
    public Object checkPerm(ProceedingJoinPoint pjp, RequirePerm requirePerm) throws Throwable {
        String need = requirePerm.value();

        // 1. 节点存在性/格式校验（先于用户权限检查，避免非法节点被静默判为 403）
        validatePermNode(need);

        // 2. 用户权限校验（原有逻辑）
        ServletRequestAttributes attrs = (ServletRequestAttributes) RequestContextHolder.getRequestAttributes();
        if (attrs == null) return pjp.proceed();
        HttpServletRequest request = attrs.getRequest();
        Long userId = (Long) request.getAttribute("user_id");
        if (userId == null) {
            Object claims = request.getAttribute("claims");
            if (claims instanceof io.jsonwebtoken.Claims c) userId = c.get("user_id", Long.class);
        }
        if (userId == null) {
            throw new ForbiddenException(need);
        }
        List<String> perms = roleService.getUserPermissions(userId);
        if (perms == null || !perms.contains(need)) {
            throw new ForbiddenException(need);
        }
        return pjp.proceed();
    }

    /**
     * 校验权限节点是否合法：非空、格式正确、且在系统已定义的权限集合中。
     * 已定义集合来自 permissions 表（通过 PermissionService.list() 加载并缓存）。
     * 校验失败抛出 {@link InvalidPermissionException}，全局异常处理器会转为 500。
     */
    private void validatePermNode(String code) {
        if (code == null || code.isBlank()) {
            log.error("RequirePerm 权限节点为空（注解值缺失）");
            throw new InvalidPermissionException("权限节点不能为空");
        }
        if (!PERM_PATTERN.matcher(code).matches()) {
            log.error("RequirePerm 非法权限码格式: {}", code);
            throw new InvalidPermissionException("非法权限节点格式: " + code + "，期望为 resource:action（如 user:list）");
        }
        Set<String> valid = getValidPerms();
        if (!valid.contains(code)) {
            log.error("RequirePerm 权限节点未在 permissions 表中定义: {}，已知节点数: {}", code, valid.size());
            throw new InvalidPermissionException("权限节点未定义: " + code);
        }
    }

    /**
     * 获取全部合法权限码集合（带缓存）。首次调用时从 DB 加载，后续直接命中内存。
     * 若 DB 暂不可用或表未初始化，则返回空集并记录警告，调用方会按“未定义”抛出 500 以暴露配置问题。
     * 提供 {@link #refresh()} 供外部在权限种子变更后刷新缓存。
     */
    private Set<String> getValidPerms() {
        Set<String> cached = validPermsCache;
        if (cached != null) {
            return cached;
        }
        synchronized (cacheLock) {
            if (validPermsCache != null) {
                return validPermsCache;
            }
            try {
                List<Permission> all = permissionService.list();
                Set<String> codes = all.stream()
                        .map(Permission::getCode)
                        .filter(c -> c != null && !c.isBlank())
                        .collect(Collectors.toSet());
                // 防御：若 DB 为空（未播种），记录警告；此时任何 RequirePerm 都会 500，避免静默通过
                if (codes.isEmpty()) {
                    log.warn("RequirePerm 权限注册表为空：permissions 表无数据，请检查是否已执行 SeedData/迁移");
                } else {
                    log.info("RequirePerm 权限注册表已加载: {} 个节点", codes.size());
                }
                // 使用不可变集合 + ConcurrentHashMap 包装以保证线程安全
                validPermsCache = ConcurrentHashMap.newKeySet();
                validPermsCache.addAll(codes);
                return validPermsCache;
            } catch (Exception e) {
                log.error("RequirePerm 加载权限注册表失败", e);
                // 加载失败时不缓存空结果，允许下次重试；此处直接抛 500
                throw new InvalidPermissionException("权限注册表加载失败: " + e.getMessage());
            }
        }
    }

    /**
     * 刷新权限缓存（权限增删后调用）。暂未自动监听 DB 变更，需手动或定时刷新。
     */
    public void refresh() {
        synchronized (cacheLock) {
            validPermsCache = null;
        }
        log.info("RequirePerm 权限注册表已刷新（下次请求将重新加载）");
    }

    /**
     * 用于测试或初始化时直接注入已知集合，避免 DB 依赖。
     */
    public void setValidPermsForTest(Set<String> codes) {
        synchronized (cacheLock) {
            validPermsCache = ConcurrentHashMap.newKeySet();
            validPermsCache.addAll(codes);
        }
    }

    public static class ForbiddenException extends RuntimeException {
        private final String perm;
        public ForbiddenException(String perm) { super("无权限: " + perm); this.perm = perm; }
        public String getPerm() { return perm; }
    }

    /**
     * 权限节点非法/未定义时抛出，映射为 500 内部错误以暴露配置错误。
     */
    public static class InvalidPermissionException extends RuntimeException {
        public InvalidPermissionException(String message) { super(message); }
    }
}
