import React from 'react';
import { Result, Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import { HomeOutlined } from '@ant-design/icons';
import { useThemeToken } from '../../hooks/useThemeToken';

const Forbidden: React.FC = () => {
  const navigate = useNavigate();
  const token = useThemeToken();
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', background: token.colorBgLayout }}>
      <Result
        status="403"
        title="403"
        subTitle="抱歉，您没有权限访问该页面。如有疑问请联系管理员。"
        extra={[
          <Button key="home" type="primary" icon={<HomeOutlined />} onClick={() => navigate('/dashboard')}>
            返回首页
          </Button>,
        ]}
      />
    </div>
  );
};
export default Forbidden;
