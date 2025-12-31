import { createBrowserRouter } from 'react-router-dom';

// routes
import RouterWrapper from './RouterWrapper';
import AuthenticationRoutes from './AuthenticationRoutes';
import MainRoutes from './MainRoutes';
import ErrorBoundary from './ErrorBoundary';

// ==============================|| ROUTING RENDER ||============================== //

// 从后端注入的配置中获取 basePath，回退到环境变量或根路径
const basePath = window.__SUBLINK_CONFIG__?.basePath || import.meta.env.VITE_APP_BASE_NAME || '/';

const router = createBrowserRouter(
  [
    {
      element: <RouterWrapper />,
      errorElement: <ErrorBoundary />,
      children: [MainRoutes, AuthenticationRoutes]
    }
  ],
  {
    basename: basePath
  }
);

export default router;
