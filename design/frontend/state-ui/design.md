# Frontend State UI — 加载/空/错误状态规范

## 三种状态必须覆盖

每个页面和组件在处理异步数据时，必须覆盖三种状态。不允许只写"正常"状态，上线后白屏/永转圈。

| 状态 | 何时出现 | 视觉 |
|------|----------|------|
| **Loading** | 首次加载数据时 | 骨架屏 / Spin |
| **Empty** | 数据为空 | 插画 + 提示文案 + 操作按钮 |
| **Error** | 网络错误 / 接口报错 | 错误提示 + 重试按钮 |

## 1. Loading — 骨架屏

### 全局 Skeleton 组件

```tsx
// components/Skeleton.tsx
interface SkeletonProps {
    type: 'table' | 'card' | 'detail' | 'form';
    rows?: number;  // table 模式下的行数
}

const Skeleton: React.FC<SkeletonProps> = ({ type, rows = 5 }) => {
    if (type === 'table') {
        return (
            <Card>
                <Row gutter={[16, 16]}>
                    <Col span={6}><AntSkeleton.Input active /></Col>
                    <Col span={6}><AntSkeleton.Input active /></Col>
                </Row>
                {Array.from({ length: rows }).map((_, i) => (
                    <AntSkeleton key={i} active paragraph={{ rows: 1 }} />
                ))}
            </Card>
        );
    }
    if (type === 'card') {
        return (
            <Row gutter={[16, 16]}>
                {Array.from({ length: 4 }).map((_, i) => (
                    <Col key={i} span={6}>
                        <Card><AntSkeleton active paragraph={{ rows: 2 }} /></Card>
                    </Col>
                ))}
            </Row>
        );
    }
    if (type === 'detail') {
        return <Card><AntSkeleton active paragraph={{ rows: 6 }} /></Card>;
    }
    return <Card><AntSkeleton active paragraph={{ rows: 8 }} /></Card>;
};
```

### 使用原则

- **相同形状**：骨架屏的外形匹配真实内容布局
- **≤ 3 秒出现**：超过 3 秒没数据才值得骨架屏，否则直接 Spin
- **首次加载用骨架屏，刷新用 Spin**——体验差异

```tsx
const UserList: React.FC = () => {
    const [loading, setLoading] = useState(true);

    if (loading && isFirstLoad) return <Skeleton type="table" />;
    // ProTable 自带 loading={loading} 处理后续刷新
};
```

## 2. Empty — 空状态

### 全局 EmptyState 组件

```tsx
// components/EmptyState.tsx
interface EmptyStateProps {
    type?: 'default' | 'search' | 'nodata';
    description?: string;
    actionText?: string;
    onAction?: () => void;
}

const config = {
    default: { image: Empty.PRESENTED_IMAGE_SIMPLE, text: '暂无数据' },
    search:  { image: Empty.PRESENTED_IMAGE_SIMPLE, text: '没有找到匹配的结果，请尝试其他关键词' },
    nodata:  { image: Empty.PRESENTED_IMAGE_SIMPLE, text: '还没有任何数据' },
};

const EmptyState: React.FC<EmptyStateProps> = ({
    type = 'default', description, actionText, onAction
}) => {
    const { image, text } = config[type];
    return (
        <Empty
            image={image}
            description={description || text}
            style={{ padding: '80px 0' }}
        >
            {actionText && onAction && (
                <Button type="primary" onClick={onAction}>{actionText}</Button>
            )}
        </Empty>
    );
};
```

### 使用示例

```tsx
// 列表为空
if (!loading && users.length === 0) {
    return (
        <EmptyState
            type="nodata"
            actionText="新增用户"
            onAction={() => setModalOpen(true)}
        />
    );
}

// 搜索结果为空
if (!loading && keyword && users.length === 0) {
    return <EmptyState type="search" actionText="清除搜索" onAction={clearSearch} />;
}
```

### 各类页面空状态指引

| 页面 | 空状态类型 | 操作按钮 |
|------|-----------|----------|
| 用户列表（无数据） | nodata | 新增用户 |
| 用户列表（搜索无果） | search | 清除搜索 |
| 角色列表 | nodata | 新增角色 |
| 菜单树（整个系统无菜单） | nodata | 新增根菜单 → 弹出引导提示 |
| 角色详情页（未分配权限） | default | 分配权限 → 打开弹窗 |
| 系统配置 | default | —（预设不可为空） |

## 3. Error — 错误状态

### 全局 ErrorResult 组件

```tsx
// components/ErrorResult.tsx
interface ErrorResultProps {
    status?: 'error' | '404' | '403' | '500';
    message?: string;
    showRetry?: boolean;
    onRetry?: () => void;
}

const ErrorResult: React.FC<ErrorResultProps> = ({
    status = 'error', message, showRetry = true, onRetry
}) => {
    const config = {
        'error': { title: '数据加载失败', subTitle: '服务器开小差了，请稍后再试' },
        '404':   { title: '页面不存在', subTitle: '请检查 URL 是否正确' },
        '403':   { title: '无权限访问', subTitle: '如需开通权限，请联系管理员' },
        '500':   { title: '服务器错误', subTitle: '我们已经记录此问题，请稍后重试' },
    };
    const { title, subTitle } = config[status];

    return (
        <Result
            status={status === 'error' ? 'error' : status === '404' ? '404' : status === '403' ? '403' : '500'}
            title={title}
            subTitle={message || subTitle}
            extra={showRetry && onRetry && (
                <Button type="primary" onClick={onRetry}>重试</Button>
            )}
        />
    );
};
```

### 使用示例

```tsx
// Page 级别
const UserList: React.FC = () => {
    const [error, setError] = useState<string | null>(null);

    const fetchUsers = async () => {
        setError(null);
        try {
            const resp = await userApi.getList({ page: 1, pageSize: 20 });
            setUsers(resp.data.items);
        } catch (err: unknown) {
            setError(err.message || '加载失败');
        }
    };

    if (error) return <ErrorResult message={error} onRetry={fetchUsers} />;
    // ...
};
```

## ProTable 三种状态合一（推荐）

Ant Design ProTable 内置了这三种状态的处理，不需要手写 if/else：

```tsx
<ProTable
    request={async (params) => {
        const resp = await userApi.getList(params);
        return { data: resp.data.items, total: resp.data.total, success: true };
    }}
    // 内置：
    // - loading → 表格内 Spin
    // - empty → 自动显示 Empty
    // - 报错 → 自动显示 Result + 重试按钮
/>
```

> **优先用 ProTable/ProList 自带的状态处理**，它是经过验证的最佳实践。只在自定义页面（如 Dashboard）才手写 Skeleton/Empty/Error。

## CRUD 操作的状态处理

| 操作 | 状态处理 |
|------|----------|
| 创建/编辑（弹窗） | Form 提交时按钮 loading；成功后 message.success + 关闭弹窗 + 刷新列表；失败后 message.error + 弹窗不关 |
| 删除 | Popconfirm 二次确认；确认后按钮 loading + 加载中禁用；成功 message.success + 移除该行 |
| 批量操作 | 选中计数显示；操作中按钮 loading；成功 message.success；失败列出具体失败的项 |

## Dashboard 的状态片段示例

```tsx
// 统计卡片
if (loading) return <Row>{[1,2,3,4].map(i => <Col span={6} key={i}><Card><Skeleton active /></Card></Col>)}</Row>;
if (error) return <ErrorResult message={error} onRetry={fetchStats} />;
if (stats.allZero) return <EmptyState type="nodata" actionText="去创建第一个用户" onAction={() => navigate('/system/users')} />;

return <Row>{/* 渲染统计卡片 */}</Row>;
```

## 全局配置

```tsx
// App.tsx
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';

<ConfigProvider
    locale={zhCN}
    theme={{ token: { colorPrimary: '#1677ff' } }}
    renderEmpty={() => <EmptyState type="default" />}  // 全局默认空状态
>
    <RouterProvider router={router} />
</ConfigProvider>
```

这样所有 Ant Design 组件（Table、Select、TreeSelect）的空状态自动统一。

## 设计要点

1. **组件库自带优先** — ProTable 的 loading/empty/error 已处理好 80% 场景
2. **首次加载用骨架屏** — 不闪烁 Spin，视觉更稳定
3. **Empty 和 Error 必须可操作** — 空状态引导"去创建"，错误状态给"重试"
4. **全局 `renderEmpty` 统一** — 不改 AntD 默认空状态的页面自动兜底
