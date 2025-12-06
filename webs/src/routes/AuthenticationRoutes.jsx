import { lazy } from 'react';

// project imports
import Loadable from 'ui-component/Loadable';
import MinimalLayout from 'layout/MinimalLayout';
import GuestGuard from 'auth/GuestGuard';

// login routing
const LoginPage = Loadable(lazy(() => import('views/pages/authentication/Login')));

// ==============================|| AUTHENTICATION ROUTING ||============================== //

const AuthenticationRoutes = {
  path: '/',
  element: <MinimalLayout />,
  children: [
    {
      path: '/login',
      element: (
        <GuestGuard>
          <LoginPage />
        </GuestGuard>
      )
    }
  ]
};

export default AuthenticationRoutes;
