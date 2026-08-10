import React, { useEffect, useState } from 'react';
import { Descriptions, Progress, Row, Col, Statistic, Spin, Tag, Space } from 'antd';
import {
  CodeOutlined,
  HddOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  AppstoreOutlined,
  DashboardOutlined,
  DeploymentUnitOutlined,
} from '@ant-design/icons';
import PageCard from '../../components/PageCard';
import { systemApi } from '../../api/system';

interface DiskInfo {
  path: string;
  fs_type: string;
  total: number;
  used: number;
  used_percent: number;
}

interface SystemInfo {
  backend_version: string;
  backend_commit: string;
  build_time: string;
  go_version: string;
  os: string;
  os_version: string;
  hostname: string;
  arch: string;
  redis_version: string;
  db_version: string;
  db_driver: string;
  cpu_cores: number;
  cpu_percent: number;
  cpu_vendor: string;
  cpu_model: string;
  mem_total: number;
  mem_used: number;
  mem_used_percent: number;
  mem_available: number;
  disk: DiskInfo[];
  net_bytes_recv: number;
  net_bytes_sent: number;
  net_recv_rate: number;
  net_sent_rate: number;
  uptime: number;
}

// 字节数格式化
function fmtBytes(n: number): string {
  if (!n) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

// 速率格式化（bytes/s → KB/s / MB/s）
function fmtRate(n: number): string {
  return `${fmtBytes(n)}/s`;
}

function fmtUptime(sec: number): string {
  if (!sec) return '-';
  const d = Math.floor(sec / 86400);
  const h = Math.floor((sec % 86400) / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (d > 0) return `${d}天 ${h}小时 ${m}分`;
  if (h > 0) return `${h}小时 ${m}分`;
  return `${m}分`;
}

const SystemInfo: React.FC = () => {
  const [info, setInfo] = useState<SystemInfo | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    systemApi
      .getInfo()
      .then((r) => {
        const data = r.data as SystemInfo;
        if (data && typeof data === 'object' && 'backend_version' in data) {
          setInfo(data);
        } else {
          // 兼容嵌套结构 {data: {...}}
          const nested = (r.data as Record<string, unknown>)?.data as SystemInfo;
          setInfo(nested || null);
        }
      })
      .catch(() => setInfo(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <Spin style={{ display: 'block', marginTop: 80 }} />;

  if (!info) {
    return <PageCard title="系统信息">加载失败或无权限查看系统信息。</PageCard>;
  }

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      {/* 版本信息 */}
      <PageCard title={<span><CodeOutlined /> 版本信息</span>} >
        <Descriptions column={3} size="small" bordered>
          <Descriptions.Item label="后端版本"><Tag color="blue">{info.backend_version}</Tag></Descriptions.Item>
          <Descriptions.Item label="Go 版本"><Tag>{info.go_version}</Tag></Descriptions.Item>
          <Descriptions.Item label="前端版本"><Tag color="green">{__APP_VERSION__}</Tag></Descriptions.Item>
          <Descriptions.Item label="构建 commit">{info.backend_commit}</Descriptions.Item>
          <Descriptions.Item label="构建时间">{info.build_time}</Descriptions.Item>
          <Descriptions.Item label="运行时长">{fmtUptime(info.uptime)}</Descriptions.Item>
        </Descriptions>
      </PageCard>

      {/* 运行环境 + 依赖 */}
      <Row gutter={16}>
        <Col span={12}>
          <PageCard title={<span><AppstoreOutlined /> 运行环境</span>} >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="操作系统">{info.os}</Descriptions.Item>
              <Descriptions.Item label="内核版本">{info.os_version}</Descriptions.Item>
              <Descriptions.Item label="主机名">{info.hostname}</Descriptions.Item>
              <Descriptions.Item label="架构">{info.arch}</Descriptions.Item>
            </Descriptions>
          </PageCard>
        </Col>
        <Col span={12}>
          <PageCard title={<span><CloudServerOutlined /> 依赖服务</span>} >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Redis 版本"><Tag color="red">{info.redis_version || '-'}</Tag></Descriptions.Item>
              <Descriptions.Item label="数据库类型">{info.db_driver || '-'}</Descriptions.Item>
              <Descriptions.Item label="数据库版本">{info.db_version || '-'}</Descriptions.Item>
            </Descriptions>
          </PageCard>
        </Col>
      </Row>

      {/* CPU + 内存 */}
      <Row gutter={16}>
        <Col span={12}>
          <PageCard title={<span><DashboardOutlined /> CPU</span>} >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="核心数">{info.cpu_cores} 核</Descriptions.Item>
              <Descriptions.Item label="型号">{info.cpu_model || info.cpu_vendor || '-'}</Descriptions.Item>
            </Descriptions>
            <div style={{ marginTop: 8 }}>
              <div style={{ marginBottom: 4 }}>使用率 {info.cpu_percent.toFixed(1)}%</div>
              <Progress percent={Number(info.cpu_percent.toFixed(1))} status={info.cpu_percent > 90 ? 'exception' : 'active'} />
            </div>
          </PageCard>
        </Col>
        <Col span={12}>
          <PageCard title={<span><HddOutlined /> 内存</span>} >
            <Row gutter={16}>
              <Col span={12}>
                <Statistic title="总内存" value={fmtBytes(info.mem_total)} />
              </Col>
              <Col span={12}>
                <Statistic title="已使用" value={fmtBytes(info.mem_used)} />
              </Col>
              <Col span={12}>
                <Statistic title="可用" value={fmtBytes(info.mem_available)} />
              </Col>
              <Col span={12}>
                <Statistic title="使用率" value={info.mem_used_percent.toFixed(1)} suffix="%" />
              </Col>
            </Row>
            <div style={{ marginTop: 8 }}>
              <Progress percent={Number(info.mem_used_percent.toFixed(1))} status={info.mem_used_percent > 90 ? 'exception' : 'active'} />
            </div>
          </PageCard>
        </Col>
      </Row>

      {/* 磁盘 + 网络 */}
      <Row gutter={16}>
        <Col span={12}>
          <PageCard title={<span><DatabaseOutlined /> 磁盘</span>} >
            {info.disk.length === 0 ? (
              <div>无磁盘数据</div>
            ) : (
              info.disk.map((d) => (
                <div key={d.path} style={{ marginBottom: 12 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <span><b>{d.path}</b> ({d.fs_type})</span>
                    <span>{fmtBytes(d.used)} / {fmtBytes(d.total)} ({d.used_percent.toFixed(1)}%)</span>
                  </div>
                  <Progress percent={Number(d.used_percent.toFixed(1))} status={d.used_percent > 90 ? 'exception' : 'normal'} />
                </div>
              ))
            )}
          </PageCard>
        </Col>
        <Col span={12}>
          <PageCard title={<span><DeploymentUnitOutlined /> 网络</span>} >
            <Row gutter={16}>
              <Col span={12}>
                <Statistic title="实时接收" value={fmtRate(info.net_recv_rate)} />
              </Col>
              <Col span={12}>
                <Statistic title="实时发送" value={fmtRate(info.net_sent_rate)} />
              </Col>
              <Col span={12}>
                <Statistic title="累计接收" value={fmtBytes(info.net_bytes_recv)} />
              </Col>
              <Col span={12}>
                <Statistic title="累计发送" value={fmtBytes(info.net_bytes_sent)} />
              </Col>
            </Row>
          </PageCard>
        </Col>
      </Row>
    </Space>
  );
};

export default SystemInfo;
