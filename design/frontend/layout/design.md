# Frontend Layout — 后台布局框架

## 职责

提供标准后台布局：可折叠侧边栏 + 顶栏（用户头像/退出）+ 内容区。菜单从后端动态获取。

## 组件结构

```tsx
<AdminLayout>
  <Sider>                         {/* 侧边栏 */}
    <Logo />                      {/* 系统名称（可折叠） */}
    <SideMenu />                  {/* 菜单树 */}
  </Sider>
  <Layout>
    <Header>                      {/* 顶栏 */}
      <CollapseBtn />             {/* 折叠切换 */}
      <Breadcrumb />
      <UserDropdown />            {/* 头像 + 退出 */}
    </Header>
    <Content>                     {/* 内容区 */}
      <Outlet />                  {/* React Router 渲染子路由 */}
    </Content>
  </Layout>
</AdminLayout>
```

## 核心实现

### AdminLayout.tsx

```tsx
const AdminLayout: React.FC = () => {
    const [collapsed, setCollapsed] = useState(false);
    const { menuTree, fetchMenus } = useMenuStore();
    const { userInfo, logout } = useAuthStore();
    const navigate = useNavigate();

    useEffect(() => {
        fetchMenus();  // 登录后获取菜单树
    }, []);

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Layout.Sider
                collapsible
                collapsed={collapsed}
                onCollapse={setCollapsed}
                theme="dark"
                width={220}
            >
                <div className="logo">
                    {collapsed ? 'K' : 'Kingfisher'}
                </div>
                <SideMenu menus={menuTree} />
            </Layout.Sider>
            <Layout>
                <Layout.Header className="header">
                    <div className="header-left">
                        <MenuFoldOutlined onClick={() => setCollapsed(!collapsed)} />
                        <Breadcrumb items={breadcrumbItems} />
                    </div>
                    <div className="header-right">
                        <Dropdown menu={{ items: [
                            { key: 'logout', label: '退出登录', icon: <LogoutOutlined />, onClick: handleLogout }
                        ]}}>
                            <Space>
                                <Avatar src={userInfo?.avatar} />
                                <span>{userInfo?.username}</span>
                            </Space>
                        </Dropdown>
                    </div>
                </Layout.Header>
                <Layout.Content className="content">
                    <Outlet />
                </Layout.Content>
            </Layout>
        </Layout>
    );
};
```

### SideMenu——递归渲染

```tsx
interface SideMenuProps {
    menus: MenuItem[];
    parentPath?: string;
}

const SideMenu: React.FC<SideMenuProps> = ({ menus, parentPath = '' }) => {
    const navigate = useNavigate();
    const location = useLocation();

    const buildItems = (items: MenuItem[]): MenuProps['items'] => {
        return items
            .filter(m => m.status === 1)   // 隐藏的不渲染
            .sort((a, b) => a.sort - b.sort)
            .map(item => {
                const fullPath = item.path.startsWith("/") ? item.path : parentPath + "/" + item.path;  // 绝对路径直用，相对路径拼父路径
                if (item.type === 1) {  // 目录——有子节点
                    return {
                        key: fullPath,
                        icon: icons[item.icon],
                        label: item.name,
                        children: item.children ?
                            buildItems(item.children, fullPath) : undefined,
                    };
                }
                // 菜单——无子节点
                return {
                    key: fullPath,
                    icon: icons[item.icon],
                    label: item.name,
                    onClick: () => navigate(fullPath),
                };
            });
    };

    return (
        <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[location.pathname]}
            items={buildItems(menus)}
        />
    );
};
```

## 菜单 Store（Zustand）

```tsx
// stores/menu.ts
interface MenuState {
    menuTree: MenuItem[];
    flatMenus: MenuItem[];       // 扁平化，用于面包屑、权限校验
    loading: boolean;
    fetchMenus: () => Promise<void>;
}

const useMenuStore = create<MenuState>((set) => ({
    menuTree: [],
    flatMenus: [],
    loading: false,
    fetchMenus: async () => {
        set({ loading: true });
        const resp = await menuApi.getTree();
        set({
            menuTree: resp.data,
            flatMenus: flatten(resp.data),  // 递归拍平
            loading: false,
        });
    },
}));
```

## 样式要点

```css
.logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 18px;
    font-weight: bold;
    border-bottom: 1px solid rgba(255,255,255,0.1);
}
.header {
    background: #fff;
    padding: 0 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}
.content {
    margin: 24px;
    padding: 24px;
    background: #fff;
    border-radius: 8px;
    min-height: calc(100vh - 64px - 48px);
}
```

## 设计要点

- 侧边栏菜单完全由后端数据驱动，前端的路由配置是兜底（hardcode 路由表 + 菜单动态渲染）
- 折叠状态存 localStorage，刷新不丢失
- 面包屑从 `flatMenus` 根据当前 path 反查生成
- 退出登录时清空 token + 所有 store 状态
