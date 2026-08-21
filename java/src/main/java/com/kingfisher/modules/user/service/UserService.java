package com.kingfisher.modules.user.service;

import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.query.Condition;
import com.kingfisher.common.query.Operator;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.user.domain.Department;
import com.kingfisher.modules.user.domain.User;
import com.kingfisher.modules.user.mapper.UserMapper;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.regex.Pattern;

@Service
public class UserService {

    private static final Pattern UPPER = Pattern.compile("[A-Z]");
    private static final Pattern LOWER = Pattern.compile("[a-z]");
    private static final Pattern DIGIT = Pattern.compile("[0-9]");

    private final UserMapper userMapper;
    private final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();

    public UserService(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    public record PageResult(List<User> items, long total, int page, int pageSize) {}

    public User getById(Long id) {
        User user = userMapper.findById(id);
        if (user == null) return null;
        enrichUser(user);
        return user;
    }

    public PageResult listUsers(Query query) {
        Map<String, Object> params = buildListParams(query);
        long total = userMapper.countAll(params);
        if (total == 0) return new PageResult(List.of(), 0, query.getPage(), query.getPageSize());

        List<User> users = userMapper.findAll(params);
        enrichUsers(users);
        return new PageResult(users, total, query.getPage(), query.getPageSize());
    }

    @Transactional
    public User createUser(String username, String password, String email,
                           List<Long> roleIds, List<Long> deptIds) {
        User existing = userMapper.findByUsername(username);
        if (existing != null) {
            throw new BizException(ErrorCode.ERR_USER_EXISTS);
        }
        validatePasswordStrength(password);
        String hashed = encoder.encode(password);
        if ((roleIds == null || roleIds.isEmpty()) && (deptIds == null || deptIds.isEmpty())) {
            roleIds = List.of(4L);
        }
        User user = new User();
        user.setUsername(username);
        user.setPassword(hashed);
        user.setEmail(email);
        user.setNickname(username);
        user.setStatus(1);
        userMapper.insert(user);
        if (roleIds != null && !roleIds.isEmpty()) {
            for (Long roleId : roleIds) {
                userMapper.insertUserRole(user.getId(), roleId);
            }
        }
        if (deptIds != null && !deptIds.isEmpty()) {
            for (Long deptId : deptIds) {
                userMapper.insertUserDepartment(user.getId(), deptId);
            }
        }
        user.setPassword(null);
        return user;
    }

    @Transactional
    public void updateUser(Long id, Map<String, Object> updates) {
        if (updates.containsKey("role_ids")) {
            @SuppressWarnings("unchecked")
            List<Long> roleIds = (List<Long>) updates.remove("role_ids");
            userMapper.deleteUserRoles(id);
            if (roleIds != null) {
                for (Long roleId : roleIds) {
                    userMapper.insertUserRole(id, roleId);
                }
            }
        }
        if (updates.containsKey("dept_ids")) {
            @SuppressWarnings("unchecked")
            List<Long> deptIds = (List<Long>) updates.remove("dept_ids");
            userMapper.deleteUserDepartments(id);
            if (deptIds != null) {
                for (Long deptId : deptIds) {
                    userMapper.insertUserDepartment(id, deptId);
                }
            }
        }
        if (!updates.isEmpty()) {
            userMapper.update(id, updates);
        }
    }

    @Transactional
    public void deleteUser(Long id) {
        userMapper.delete(id);
    }

    @Transactional
    public void batchDelete(List<Long> ids) {
        userMapper.deleteBatch(ids);
    }

    @Transactional
    public void batchUpdateStatus(List<Long> ids, Integer status) {
        userMapper.updateStatusBatch(ids, status);
    }

    @Transactional
    public void changePassword(Long userId, String oldPwd, String newPwd) {
        User user = userMapper.findById(userId);
        if (user == null) {
            throw new BizException(ErrorCode.ERR_USER_NOT_FOUND);
        }
        if (!encoder.matches(oldPwd, user.getPassword())) {
            throw new BizException(ErrorCode.ERR_PASSWORD_WRONG);
        }
        validatePasswordStrength(newPwd);
        Map<String, Object> fields = new HashMap<>();
        fields.put("password", encoder.encode(newPwd));
        userMapper.update(userId, fields);
        userMapper.incrementSessionVersion(userId);
    }

    @Transactional
    public void updateProfile(Long userId, String email, String nickname, String avatar) {
        Map<String, Object> fields = new HashMap<>();
        if (email != null && !email.isBlank()) fields.put("email", email);
        if (nickname != null && !nickname.isBlank()) fields.put("nickname", nickname);
        if (avatar != null && !avatar.isBlank()) fields.put("avatar", avatar);
        if (!fields.isEmpty()) {
            userMapper.update(userId, fields);
        }
    }

    @Transactional
    public void revokeSessions(Long userId) {
        userMapper.incrementSessionVersion(userId);
    }

    public void enrichUser(User user) {
        if (user == null) return;
        Long uid = user.getId();
        user.setDirectRoleIds(userMapper.findDirectRoleIds(uid));
        user.setDeptIds(userMapper.findDeptIds(uid));
        user.setDepartments(userMapper.findDepartments(uid));
        List<Long> inherited = userMapper.findDeptInheritedRoleIds(uid);
        if (inherited != null && !inherited.isEmpty() && user.getRoles() != null) {
            Set<Long> existing = new HashSet<>();
            for (var r : user.getRoles()) existing.add(r.getId());
            // inherited role details are already in roles via findById join; dept inheritance is additive
        }
    }

    public void enrichUsers(List<User> users) {
        if (users == null) return;
        for (User u : users) enrichUser(u);
    }

    private Map<String, Object> buildListParams(Query query) {
        Map<String, Object> params = new HashMap<>();
        params.put("keyword", query.getKeyword());
        params.put("limit", query.getPageSize());
        params.put("offset", query.offset());
        params.put("sort", query.sortExpr());

        for (Condition c : query.getFilters()) {
            switch (c.getField()) {
                case "status" -> {
                    if (c.getOp() == Operator.IN) {
                        // use first value for simplicity; SQLite has no array param
                        List<?> vals = (List<?>) c.getValue();
                        if (!vals.isEmpty()) params.put("status", vals.get(0));
                    } else {
                        params.put("status", c.getValue());
                    }
                }
                case "role_id" -> params.put("roleId", c.getValue());
                case "department_id" -> params.put("departmentId", c.getValue());
                case "created_at" -> {
                    if (c.getOp() == Operator.GTE) params.put("createdAtStart", toSqlite(c.getValue()));
                    else if (c.getOp() == Operator.LTE) params.put("createdAtEnd", toSqlite(c.getValue()));
                    else if (c.getOp() == Operator.GT) params.put("createdAtStart", toSqlite(c.getValue()));
                    else if (c.getOp() == Operator.LT) params.put("createdAtEnd", toSqlite(c.getValue()));
                }
                case "username" -> {
                    if (c.getOp() == Operator.CONTAINS) params.put("keyword", c.getValue());
                }
                case "email" -> {
                    if (c.getOp() == Operator.CONTAINS) params.put("keyword", c.getValue());
                }
            }
        }
        return params;
    }

    private String toSqlite(Object val) {
        if (val instanceof Instant inst) {
            return DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")
                    .withZone(ZoneOffset.UTC).format(inst);
        }
        return String.valueOf(val);
    }

    private void validatePasswordStrength(String pwd) {
        if (pwd == null || pwd.length() < 8) throw new BizException(ErrorCode.ERR_PASSWORD_TOO_SHORT);
        if (pwd.length() > 64) throw new BizException(ErrorCode.ERR_PASSWORD_TOO_LONG);
        if (!UPPER.matcher(pwd).find() || !LOWER.matcher(pwd).find() || !DIGIT.matcher(pwd).find()) {
            throw new BizException(ErrorCode.ERR_PASSWORD_WEAK);
        }
    }

    public static class BizException extends RuntimeException {
        private final int code;
        public BizException(int code) {
            super(ErrorCode.message(code));
            this.code = code;
        }
        public int getCode() { return code; }
    }
}
