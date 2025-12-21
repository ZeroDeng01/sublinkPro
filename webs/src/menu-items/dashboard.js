// assets
import { IconDashboard, IconWorldLatitude } from '@tabler/icons-react';

// ==============================|| DASHBOARD MENU ITEMS ||============================== //

const dashboard = {
  id: 'dashboard',
  title: '仪表盘',
  type: 'group',
  children: [
    {
      id: 'default',
      title: '仪表盘',
      type: 'item',
      url: '/dashboard/default',
      icon: IconDashboard,
      breadcrumbs: false
    },
    {
      id: 'node-map',
      title: '节点地图',
      type: 'item',
      url: '/dashboard/map',
      icon: IconWorldLatitude,
      breadcrumbs: false
    }
  ]
};

export default dashboard;
