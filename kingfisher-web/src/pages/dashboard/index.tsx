import PageCard from '../../components/PageCard';
import { useThemeToken } from '../../hooks/useThemeToken';
import { useAuthStore } from '../../stores/auth';

const Dashboard: React.FC = () => {
  const token = useThemeToken();
  const userInfo = useAuthStore((s) => s.userInfo) as Record<string, unknown> | null;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* 欢迎横幅 */}
      <PageCard>
        <div>
          <div style={{ fontSize: 20, fontWeight: 700, color: token.colorText }}>
            你好，{(userInfo?.nickname as string) || (userInfo?.username as string) || '管理员'} 👋
          </div>
          <div style={{ marginTop: 6, fontSize: 14, color: token.colorTextTertiary }}>欢迎回来，这里是系统总览。</div>
        </div>
      </PageCard>
    </div>
  );
};

export default Dashboard;
