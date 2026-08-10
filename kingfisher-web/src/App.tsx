import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider, App as AntApp, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import AppRoutes from './router';
import { GlobalFeedback } from './utils/feedback';
import { ThemeProvider } from './hooks/ThemeProvider';
import { useTheme } from './hooks/useTheme';
import SiteNotice from './components/SiteNotice';

function ThemedApp() {
  const { theme: current } = useTheme();
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: { colorPrimary: '#1677ff' },
        algorithm: current === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
      }}
    >
      <AntApp>
        <GlobalFeedback />
        {/* 站点通知：全局顶部横幅，未登录也可见，可关闭 */}
        <SiteNotice />
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
}

function App() {
  return (
    <ThemeProvider>
      <ThemedApp />
    </ThemeProvider>
  );
}

export default App;
