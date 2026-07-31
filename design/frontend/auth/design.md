# Frontend Auth — 登录 & 注册

## 职责

登录页、注册页、token 管理、Auth Guard（路由守卫）。

## 页面

### LoginPage

```
┌──────────────────────────────────┐
│                                  │
│         🦜 Kingfisher           │  系统 Logo + 名称
│         后台管理系统              │
│                                  │
│   ┌────────────────────────┐    │
│   │  用户名                 │    │  输入框
│   └────────────────────────┘    │
│   ┌────────────────────────┐    │
│   │  密码                   │    │  密码框
│   └────────────────────────┘    │
│   ┌────────────────────────┐    │
│   │      登 录               │    │  主按钮
│   └────────────────────────┘    │
│         还没有账号？去注册        │
│                                  │
└──────────────────────────────────┘
```

### 实现

```tsx
const LoginPage: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const { login } = useAuthStore();
    const navigate = useNavigate();

    const onFinish = async (values: LoginReq) => {
        setLoading(true);
        try {
            await login(values.username, values.password);
            message.success('登录成功');
            navigate('/', { replace: true });
        } catch (err: unknown) {
            message.error(err.message || '登录失败');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-container">
            <Card className="login-card" bordered={false}>
                <div className="login-header">
                    <h2>Kingfisher</h2>
                    <p>后台管理系统</p>
                </div>
                <Form onFinish={onFinish} size="large">
                    <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
                        <Input prefix={<UserOutlined />} placeholder="用户名" />
                    </Form.Item>
                    <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
                        <Input.Password prefix={<LockOutlined />} placeholder="密码" />
                    </Form.Item>
                    <Form.Item>
                        <Button type="primary" htmlType="submit" loading={loading} block>
                            登 录
                        </Button>
                    </Form.Item>
                </Form>
                <div className="login-footer">
                    <Link to="/register">还没有账号？去注册</Link>
                </div>
            </Card>
        </div>
    );
};
```

## Auth Store（Zustand）

```tsx
// stores/auth.ts
interface AuthState {
    token: string | null;
    refreshToken: string | null;
    userInfo: User | null;
    permissions: string[];             // 权限 code 列表
    isLoggedIn: boolean;

    login: (username: string, password: string) => Promise<void>;
    logout: () => void;
    refreshAccessToken: () => Promise<string>;   // 刷新 token
    fetchUserInfo: () => Promise<void>;
}

const useAuthStore = create<AuthState>((set, get) => ({
    token: getToken(),                   // 从 localStorage 恢复
    refreshToken: getRefreshToken(),
    userInfo: null,
    permissions: [],
    isLoggedIn: !!getToken(),

    login: async (username, password) => {
        const resp = await authApi.login({ username, password });
        setToken(resp.data.access_token);
        setRefreshToken(resp.data.refresh_token);
        set({
            token: resp.data.access_token,
            refreshToken: resp.data.refresh_token,
            userInfo: resp.data.user,
            isLoggedIn: true,
        });
        // 登录后拉取权限
        const permResp = await userApi.getMyPermissions();
        set({ permissions: permResp.data });
    },

    logout: () => {
        removeToken();
        removeRefreshToken();
        set({ token: null, refreshToken: null, userInfo: null, permissions: [], isLoggedIn: false });
    },

    refreshAccessToken: async () => {
        const resp = await authApi.refresh(get().refreshToken!);
        setToken(resp.data.access_token);
        set({ token: resp.data.access_token });
        return resp.data.access_token;
    },
}));
```

## Auth Guard（路由守卫）

```tsx
// router/guard.tsx
const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { isLoggedIn } = useAuthStore();
    const location = useLocation();

    if (!isLoggedIn) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }
    return <>{children}</>;
};
```

## 路由配置

```tsx
const router = createBrowserRouter([
    { path: '/login', element: <LoginPage /> },
    { path: '/register', element: <RegisterPage /> },
    {
        path: '/',
        element: <AuthGuard><AdminLayout /></AuthGuard>,
        children: [
            { path: 'dashboard', element: <Dashboard /> },
            { path: 'system/users', element: <UserList /> },
            { path: 'system/menus', element: <MenuManage /> },
            { path: 'system/roles', element: <RoleList /> },
            { path: 'system/configs', element: <ConfigManage /> },
        ],
    },
]);
```

## Token 存储

```tsx
// utils/token.ts
const TOKEN_KEY = 'kingfisher_token';
const REFRESH_KEY = 'kingfisher_refresh';

export const getToken = () => localStorage.getItem(TOKEN_KEY);
export const setToken = (t: string) => localStorage.setItem(TOKEN_KEY, t);
export const removeToken = () => localStorage.removeItem(TOKEN_KEY);
// refresh token 同理
```

## 设计要点

- 登录成功后立即拉取用户权限列表，存入 store
- Auth Guard 在路由层拦截，未登录跳 `/login`，登录后跳回原页面（`state.from`）
- Token 持久化在 localStorage（access_token、refresh_token 分离存储）
- 退出时清空所有 localStorage 和 store 状态
