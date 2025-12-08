import { createBrowserRouter } from 'react-router-dom';

// routes
import RouterWrapper from './RouterWrapper';
import AuthenticationRoutes from './AuthenticationRoutes';
import MainRoutes from './MainRoutes';
import ErrorBoundary from './ErrorBoundary';

// ==============================|| ROUTING RENDER ||============================== //

const router = createBrowserRouter(
  [
    {
      element: <RouterWrapper />,
      errorElement: <ErrorBoundary />,
      children: [MainRoutes, AuthenticationRoutes]
    }
  ],
  {
    basename: import.meta.env.VITE_APP_BASE_NAME
  }
);

export default router;
