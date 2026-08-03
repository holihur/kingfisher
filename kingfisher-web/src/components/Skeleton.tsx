import React from 'react';
import { Card, Row, Col, Skeleton as AntSkeleton } from 'antd';

interface SkeletonProps { type: 'table' | 'card' | 'detail'; rows?: number; }

const Skeleton: React.FC<SkeletonProps> = ({ type, rows = 5 }) => {
  if (type === 'table') return <Card>{Array.from({ length: rows }).map((_, i) => <AntSkeleton key={i} active paragraph={{ rows: 1 }} />)}</Card>;
  if (type === 'card') return <Row gutter={[16, 16]}>{Array.from({ length: 4 }).map((_, i) => <Col key={i} span={6}><Card><AntSkeleton active paragraph={{ rows: 2 }} /></Card></Col>)}</Row>;
  return <Card><AntSkeleton active paragraph={{ rows: 6 }} /></Card>;
};
export default Skeleton;
