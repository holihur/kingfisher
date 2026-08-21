package com.kingfisher.security;

import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.rbac.service.RoleService;
import org.aspectj.lang.ProceedingJoinPoint;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import java.util.List;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class RequirePermAspectTest {

    @Autowired RoleService roleService;
    @Autowired com.kingfisher.modules.rbac.service.PermissionService permissionService;
    private RequirePermAspect aspect;

    @BeforeEach
    void setUp() {
        aspect = new RequirePermAspect(roleService, permissionService);
        aspect.setValidPermsForTest(Set.of("user:list", "user:create", "role:list", "menu:list", "config:list", "audit:list", "worktask:list"));
    }

    private com.kingfisher.modules.rbac.domain.Permission perm(String code) {
        var p = new com.kingfisher.modules.rbac.domain.Permission();
        p.setCode(code);
        p.setName(code);
        return p;
    }

    private RequirePerm ann(String value) {
        return new RequirePerm() {
            @Override public Class<? extends java.lang.annotation.Annotation> annotationType() { return RequirePerm.class; }
            @Override public String value() { return value; }
        };
    }

    @Test
    void validPerm_withUserPermission_shouldProceed() throws Throwable {
        MockHttpServletRequest req = new MockHttpServletRequest();
        req.setAttribute("user_id", 1L);
        RequestContextHolder.setRequestAttributes(new ServletRequestAttributes(req));

        ProceedingJoinPoint pjp = mock(ProceedingJoinPoint.class);
        when(pjp.proceed()).thenReturn("ok");

        Object result = aspect.checkPerm(pjp, ann("user:list"));
        assertEquals("ok", result);
        verify(pjp).proceed();
    }

    @Test
    void validPerm_withoutUserPermission_shouldThrowForbidden() {
        MockHttpServletRequest req = new MockHttpServletRequest();
        req.setAttribute("user_id", 3L);
        RequestContextHolder.setRequestAttributes(new ServletRequestAttributes(req));

        ProceedingJoinPoint pjp = mock(ProceedingJoinPoint.class);
        assertThrows(RequirePermAspect.ForbiddenException.class, () -> aspect.checkPerm(pjp, ann("user:create")));
    }

    @Test
    void blankPerm_shouldThrowInvalid() {
        ProceedingJoinPoint pjp = mock(ProceedingJoinPoint.class);
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("")));
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("   ")));
    }

    @Test
    void badFormat_shouldThrowInvalid() {
        ProceedingJoinPoint pjp = mock(ProceedingJoinPoint.class);
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("badformat")));
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("User:list")));
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("user:")));
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann(":list")));
    }

    @Test
    void unknownPerm_shouldThrowInvalid() {
        ProceedingJoinPoint pjp = mock(ProceedingJoinPoint.class);
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("unknown:perm")));
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("user:delete")));
    }

    @Test
    void noRequestAttributes_shouldProceedWithoutUserCheck_butStillValidateNode() throws Throwable {
        RequestContextHolder.resetRequestAttributes();
        ProceedingJoinPoint pjp = mock(ProceedingJoinPoint.class);
        when(pjp.proceed()).thenReturn("ok");
        Object result = aspect.checkPerm(pjp, ann("user:list"));
        assertEquals("ok", result);
    }

    @Test
    void cacheRefresh_shouldReload() throws Throwable {
        MockHttpServletRequest req = new MockHttpServletRequest();
        req.setAttribute("user_id", 1L);
        RequestContextHolder.setRequestAttributes(new ServletRequestAttributes(req));
        ProceedingJoinPoint pjp = mock(ProceedingJoinPoint.class);
        when(pjp.proceed()).thenReturn("ok");

        aspect.setValidPermsForTest(Set.of("user:list"));
        assertThrows(RequirePermAspect.InvalidPermissionException.class, () -> aspect.checkPerm(pjp, ann("config:list")));

        aspect.refresh();
        aspect.setValidPermsForTest(Set.of("config:list"));
        Object result = aspect.checkPerm(pjp, ann("config:list"));
        assertEquals("ok", result);
    }
}
