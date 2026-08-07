import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import router from './router';
import { GlobalFeedback } from './utils/feedback';

function App() {
  return (
    <ConfigProvider locale={zhCN} theme={{ token: { colorPrimary: '#1677ff' } }}>
      <AntApp>
        <GlobalFeedback />
        <RouterProvider router={router} />
      </AntApp>
    </ConfigProvider>
  );
}

export default App;
