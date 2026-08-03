import React from 'react';
import { Result, Button } from 'antd';
import { useNavigate } from 'react-router-dom';
const Forbidden: React.FC = () => {
  const navigate = useNavigate();
  return <Result status="403" title="403" subTitle="无权限访问" extra={<Button type="primary" onClick={() => navigate('/dashboard')}>返回首页</Button>} />;
};
export default Forbidden;
