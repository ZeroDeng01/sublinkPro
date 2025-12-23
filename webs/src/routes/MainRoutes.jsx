import { lazy } from 'react';
import { Navigate } from 'react-router-dom';

// project imports
import MainLayout from 'layout/MainLayout';
import Loadable from 'ui-component/Loadable';
import AuthGuard from 'auth/AuthGuard';

// dashboard routing
const DashboardDefault = Loadable(lazy(() => import('views/dashboard/Default')));
const NodeMapPage = Loadable(lazy(() => import('views/dashboard/NodeMap')));

// views routing
const NodeList = Loadable(lazy(() => import('views/nodes')));
const SubscriptionList = Loadable(lazy(() => import('views/subscriptions')));
const TemplateList = Loadable(lazy(() => import('views/templates')));
const ScriptList = Loadable(lazy(() => import('views/scripts')));
const AccessKeyList = Loadable(lazy(() => import('views/accesskeys')));
const UserSettings = Loadable(lazy(() => import('views/settings')));
const SystemMonitor = Loadable(lazy(() => import('views/monitor')));
const TagList = Loadable(lazy(() => import('views/tags')));
const TaskList = Loadable(lazy(() => import('views/tasks')));
const HostList = Loadable(lazy(() => import('views/hosts')));
const AirportList = Loadable(lazy(() => import('views/airports')));
const NodeCheckList = Loadable(lazy(() => import('views/node-check')));

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
      element: <Navigate to="/dashboard/default" replace />
    },
    {
      path: 'dashboard',
      children: [
        {
          index: true,
          element: <Navigate to="/dashboard/default" replace />
        },
        {
          path: 'default',
          element: <DashboardDefault />
        },
        {
          path: 'map',
          element: <NodeMapPage />
        }
      ]
    },
    {
      path: 'subscription',
      children: [
        {
          path: 'nodes',
          element: <NodeList />
        },
        {
          path: 'node-check',
          element: <NodeCheckList />
        },
        {
          path: 'subs',
          element: <SubscriptionList />
        },
        {
          path: 'templates',
          element: <TemplateList />
        },
        {
          path: 'tags',
          element: <TagList />
        },
        {
          path: 'airports',
          element: <AirportList />
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
        },
        {
          path: 'tasks',
          element: <TaskList />
        },
        {
          path: 'hosts',
          element: <HostList />
        }
      ]
    }
  ]
};

export default MainRoutes;
