import React from 'react';
import { Result, Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import { HomeOutlined } from '@ant-design/icons';
import { useThemeToken } from '../../hooks/useThemeToken';

const NotFound: React.FC = () => {
  const navigate = useNavigate();
  const token = useThemeToken();
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', background: token.colorBgLayout }}>
      <Result
        status="404"
        title="404"
        subTitle="抱歉，您访问的页面不存在或已被移除。"
        extra={[
          <Button key="home" type="primary" icon={<HomeOutlined />} onClick={() => navigate('/dashboard')}>
            返回首页
          </Button>,
        ]}
      />
    </div>
  );
};
export default NotFound;
