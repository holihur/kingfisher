import { Card, Row, Col, Statistic, List, Tag, Badge, Avatar, Empty } from 'antd';
import { UserOutlined, MenuOutlined, SafetyOutlined, SettingOutlined, ArrowRightOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import request from '../../api/request';
import { useAuthStore } from '../../stores/auth';
import { formatRelativeTime } from '../../utils/format';

const statCards = [
  { key: 'users', title: '用户', icon: <UserOutlined />, color: '#1677ff', link: '/system/users' },
  { key: 'roles', title: '角色', icon: <SafetyOutlined />, color: '#722ed1', link: '/system/roles' },
  { key: 'menus', title: '菜单', icon: <MenuOutlined />, color: '#13c2c2', link: '/system/menus' },
  { key: 'configs', title: '配置', icon: <SettingOutlined />, color: '#fa8c16', link: '/system/configs' },
];

const ACTION_MAP: Record<string, { text: string; color?: string }> = {
  create: { text: '创建', color: 'green' },
  update: { text: '更新', color: 'blue' },
  delete: { text: '删除', color: 'red' },
  login: { text: '登录' },
  logout: { text: '退出' },
  register: { text: '注册' },
};

interface RecentUser {
  id: number;
  username: string;
  nickname?: string;
  avatar?: string;
  role?: { name?: string };
  role_name?: string;
  created_at: string;
}

interface AuditRow {
  id: number;
  username: string;
  action: string;
  resource: string;
  resource_id: number;
  ip: string;
  created_at: string;
}

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState({ users: 0, menus: 0, roles: 0, configs: 0 });
  const [recentUsers, setRecentUsers] = useState<RecentUser[]>([]);
  const [recentLogs, setRecentLogs] = useState<AuditRow[]>([]);
  const userInfo = useAuthStore((s) => s.userInfo) as Record<string, unknown> | null;
  const navigate = useNavigate();

  useEffect(() => {
    Promise.all([
      request.get('/users?page=1&page_size=1'),
      request.get('/menus/tree'),
      request.get('/roles?page=1&page_size=1'),
      request.get('/configs?page=1&page_size=1'),
      request.get('/users?page=1&page_size=5&sort=-created_at'),
      request.get('/audit-logs?page=1&page_size=6&sort=-created_at'),
    ])
      .then(([u, m, r, c, u2, a]) => {
        setStats({
          users: (u.data as Record<string, unknown>).total as number,
          menus: ((m.data as unknown[]) || []).length,
          roles: ((r.data as Record<string, unknown>).total as number) || 0,
          configs: ((c.data as Record<string, unknown>).total as number) || 0,
        });
        const userData = u2.data as Record<string, unknown>;
        setRecentUsers((userData.items as RecentUser[]) || []);
        const auditData = a.data as Record<string, unknown>;
        setRecentLogs((auditData.items as AuditRow[]) || []);
      })
      .catch(() => {});
  }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* 欢迎横幅 */}
      <Card style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 700, color: '#1a1a2e' }}>
            你好，{(userInfo?.nickname as string) || (userInfo?.username as string) || '管理员'} 👋
          </div>
          <div style={{ marginTop: 6, fontSize: 14, color: '#8c8c8c' }}>
            欢迎回来，这里是系统总览。
          </div>
        </div>
      </Card>

      {/* 统计卡 */}
      <Row gutter={[16, 16]}>
        {statCards.map((c) => (
          <Col xs={12} sm={12} md={6} key={c.key}>
            <Card
              hoverable
              style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}
              styles={{ body: { padding: '20px 24px' } }}
              onClick={() => navigate(c.link)}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <div
                  style={{
                    width: 48,
                    height: 48,
                    borderRadius: 10,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 22,
                    color: '#fff',
                    background: c.color,
                    flexShrink: 0,
                  }}
                >
                  {c.icon}
                </div>
                <Statistic
                  title={c.title}
                  value={stats[c.key as keyof typeof stats]}
                  valueStyle={{ fontSize: 24, fontWeight: 600 }}
                />
                <ArrowRightOutlined style={{ marginLeft: 'auto', color: '#bfbfbf', fontSize: 12 }} />
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 数据支撑：最近用户 + 最近审计 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card
            title={<span style={{ fontSize: 15, fontWeight: 600 }}>最近用户</span>}
            extra={<a onClick={() => navigate('/system/users')}>全部 <ArrowRightOutlined style={{ fontSize: 11 }} /></a>}
            style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}
          >
            <List
              dataSource={recentUsers}
              locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无用户" /> }}
              renderItem={(u) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar size="small" src={u.avatar}>{u.username?.charAt(0)?.toUpperCase()}</Avatar>}
                    title={
                      <span>
                        {u.username}
                        {u.nickname ? <span style={{ color: '#8c8c8c', marginLeft: 6, fontSize: 12 }}>({u.nickname})</span> : null}
                      </span>
                    }
                    description={formatRelativeTime(u.created_at)}
                  />
                  <Tag>{u.role_name || u.role?.name || '未知角色'}</Tag>
                </List.Item>
              )}
            />
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card
            title={<span style={{ fontSize: 15, fontWeight: 600 }}>最近操作</span>}
            extra={<a onClick={() => navigate('/system/audit')}>全部 <ArrowRightOutlined style={{ fontSize: 11 }} /></a>}
            style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}
          >
            <List
              dataSource={recentLogs}
              locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无操作记录" /> }}
              renderItem={(l) => {
                const a = ACTION_MAP[l.action] || { text: l.action };
                return (
                  <List.Item>
                    <List.Item.Meta
                      title={
                        <span>
                          <Badge status="processing" style={{ marginRight: 6 }} />
                          {l.username} <Tag color={a.color} style={{ marginLeft: 4 }}>{a.text}</Tag>
                          <span style={{ color: '#8c8c8c', marginLeft: 6 }}>{l.resource}</span>
                        </span>
                      }
                      description={formatRelativeTime(l.created_at)}
                    />
                    <span style={{ color: '#bfbfbf', fontSize: 12 }}>{l.ip}</span>
                  </List.Item>
                );
              }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
