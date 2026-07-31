import { Card, Row, Col, Statistic } from 'antd';
import { UserOutlined, MenuOutlined, SafetyOutlined, SettingOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import request from '../../api/request';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState({ users: 0, menus: 0, roles: 0, configs: 0 });
  useEffect(() => {
    Promise.all([
      request.get('/users?page=1&page_size=1'),
      request.get('/menus/tree'),
      request.get('/roles'),
      request.get('/configs'),
    ])
      .then(([u, m, r, c]) => {
        setStats({
          users: (u.data as Record<string, unknown>).total as number,
          menus: ((m.data as unknown[]) || []).length,
          roles: ((r.data as unknown[]) || []).length,
          configs: ((c.data as unknown[]) || []).length,
        });
      })
      .catch(() => {});
  }, []);
  return (
    <div>
      <h3>Dashboard</h3>
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic title="用户" value={stats.users} prefix={<UserOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="菜单" value={stats.menus} prefix={<MenuOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="角色" value={stats.roles} prefix={<SafetyOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="配置" value={stats.configs} prefix={<SettingOutlined />} />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
