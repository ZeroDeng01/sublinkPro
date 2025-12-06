// assets
import { IconDashboard } from '@tabler/icons-react';

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
      url: '/dashboard',
      icon: IconDashboard,
      breadcrumbs: false
    }
  ]
};

export default dashboard;
