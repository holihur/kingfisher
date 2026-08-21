package com.kingfisher.security;

import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.rbac.domain.Permission;
import com.kingfisher.modules.rbac.service.PermissionService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Component;

import java.lang.reflect.Method;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

/**
 * 启动时校验：扫描所有 {@link RequirePerm} 注解的权限码，提前暴露拼写错误或未入库的权限节点。
 * <p>
 * 仅在应用启动完成后执行一次；校验失败会记录 ERROR 并抛出异常以阻止应用以错误配置启动，
 * 避免运行时将错误节点静默判为 403 而掩盖问题。
 * </p>
 */
@Component
public class PermNodeStartupValidator implements ApplicationRunner {

    private static final Logger log = LoggerFactory.getLogger(PermNodeStartupValidator.class);
    private static final Pattern PERM_PATTERN = Pattern.compile("^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$");

    private final ApplicationContext ctx;
    private final PermissionService permissionService;

    public PermNodeStartupValidator(ApplicationContext ctx, PermissionService permissionService) {
        this.ctx = ctx;
        this.permissionService = permissionService;
    }

    @Override
    public void run(ApplicationArguments args) {
        Set<String> validPerms;
        try {
            List<Permission> all = permissionService.list();
            validPerms = all.stream()
                    .map(Permission::getCode)
                    .filter(c -> c != null && !c.isBlank())
                    .collect(Collectors.toSet());
        } catch (Exception e) {
            log.warn("启动校验：无法加载 permissions 表，跳过权限节点校验（可能为首次建库）: {}", e.getMessage());
            return;
        }

        if (validPerms.isEmpty()) {
            log.warn("启动校验：permissions 表为空，跳过 @RequirePerm 校验");
            return;
        }

        Set<String> usedCodes = new HashSet<>();
        // 扫描所有 Bean 的方法与类上的 @RequirePerm
        for (String beanName : ctx.getBeanDefinitionNames()) {
            Object bean = ctx.getBean(beanName);
            Class<?> targetClass = bean.getClass();
            // 类级别注解
            RequirePerm classAnn = targetClass.getAnnotation(RequirePerm.class);
            if (classAnn != null) {
                usedCodes.add(classAnn.value());
            }
            for (Method m : targetClass.getMethods()) {
                RequirePerm ann = m.getAnnotation(RequirePerm.class);
                if (ann != null) {
                    usedCodes.add(ann.value());
                }
            }
            // 也检查原始类（代理场景）
            Class<?> userClass = org.springframework.aop.support.AopUtils.getTargetClass(bean);
            if (userClass != targetClass) {
                RequirePerm ca = userClass.getAnnotation(RequirePerm.class);
                if (ca != null) usedCodes.add(ca.value());
                for (Method m : userClass.getMethods()) {
                    RequirePerm ann = m.getAnnotation(RequirePerm.class);
                    if (ann != null) usedCodes.add(ann.value());
                }
            }
        }

        if (usedCodes.isEmpty()) {
            log.info("启动校验：未发现任何 @RequirePerm 注解，已定义权限数: {}", validPerms.size());
            return;
        }

        boolean hasError = false;
        for (String code : usedCodes) {
            if (code == null || code.isBlank() || !PERM_PATTERN.matcher(code).matches()) {
                log.error("启动校验失败：@RequirePerm 权限码格式非法: \"{}\"，期望为 resource:action", code);
                hasError = true;
                continue;
            }
            if (!validPerms.contains(code)) {
                log.error("启动校验失败：@RequirePerm 引用了未定义的权限节点: \"{}\"，已知节点: {}", code, validPerms);
                hasError = true;
            }
        }

        if (hasError) {
            // 抛出异常以阻止应用带错误配置启动；如需仅告警可改为 return
            throw new IllegalStateException("存在非法的 @RequirePerm 权限节点，请检查日志并修正后重启");
        }
        log.info("启动校验通过：{} 个 @RequirePerm 节点均已在 permissions 表中定义", usedCodes.size());
    }
}
