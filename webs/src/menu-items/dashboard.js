// assets
import { IconDashboard, IconWorldLatitude } from '@tabler/icons-react';

// ==============================|| DASHBOARD MENU ITEMS ||============================== //

const dashboard = {
  id: 'dashboard',
  title: 'Dashboard',
  titleKey: 'navigation.groups.dashboard',
  type: 'group',
  children: [
    {
      id: 'default',
      title: 'Dashboard',
      titleKey: 'navigation.items.dashboard',
      type: 'item',
      url: '/dashboard/default',
      icon: IconDashboard,
      breadcrumbs: false
    },
    {
      id: 'node-map',
      title: 'Node Map',
      titleKey: 'navigation.items.nodeMap',
      type: 'item',
      url: '/dashboard/map',
      icon: IconWorldLatitude,
      breadcrumbs: false
    }
  ]
};

export default dashboard;
