import { lazy } from 'react';
import { Navigate } from 'react-router-dom';

// project imports
import MainLayout from 'layout/MainLayout';
import Loadable from 'ui-component/Loadable';
import AuthGuard from 'auth/AuthGuard';

// dashboard routing
const DashboardDefault = Loadable(lazy(() => import('views/dashboard/Default')));

// views routing
const NodeList = Loadable(lazy(() => import('views/nodes')));
const SubscriptionList = Loadable(lazy(() => import('views/subscriptions')));
const TemplateList = Loadable(lazy(() => import('views/templates')));
const ScriptList = Loadable(lazy(() => import('views/scripts')));
const AccessKeyList = Loadable(lazy(() => import('views/accesskeys')));
const UserSettings = Loadable(lazy(() => import('views/settings')));
const SystemMonitor = Loadable(lazy(() => import('views/monitor')));

// ==============================|| MAIN ROUTING ||==============================  //

const MainRoutes = {
  path: '/',
  element: (
    <AuthGuard>
      <MainLayout />
    </AuthGuard>
  ),
  children: [
    {
      path: '/',
      element: <Navigate to="/dashboard" replace />
    },
    {
      path: 'dashboard',
      element: <DashboardDefault />
    },
    {
      path: 'subscription',
      children: [
        {
          path: 'nodes',
          element: <NodeList />
        },
        {
          path: 'subs',
          element: <SubscriptionList />
        },
        {
          path: 'templates',
          element: <TemplateList />
        }
      ]
    },
    {
      path: 'script',
      element: <ScriptList />
    },
    {
      path: 'accesskey',
      element: <AccessKeyList />
    },
    {
      path: 'settings',
      element: <UserSettings />
    },
    {
      path: 'system',
      children: [
        {
          path: 'user',
          element: <UserSettings />
        },
        {
          path: 'monitor',
          element: <SystemMonitor />
        }
      ]
    }
  ]
};

export default MainRoutes;

