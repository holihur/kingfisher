import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider, App as AntApp, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import AppRoutes from './router';
import { GlobalFeedback } from './utils/feedback';
import { ThemeProvider } from './hooks/ThemeProvider';
import { useTheme } from './hooks/useTheme';

function ThemedApp() {
  const { theme: current } = useTheme();
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: { colorPrimary: '#1677ff' },
        algorithm: current === 'dark' ? theme.darkAlgorithm : undefined,
      }}
    >
      <AntApp>
        <GlobalFeedback />
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
