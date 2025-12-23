import {
  IconNetwork,
  IconList,
  IconTemplate,
  IconScript,
  IconKey,
  IconSettings,
  IconDeviceDesktopAnalytics,
  IconTags,
  IconListCheck,
  IconWorld,
  IconPlane
} from '@tabler/icons-react';

// ==============================|| SUBSCRIPTION MENU ITEMS ||============================== //

const subscription = {
  id: 'subscription',
  title: '订阅管理',
  type: 'group',
  children: [
    {
      id: 'airports',
      title: '机场管理',
      type: 'item',
      url: '/subscription/airports',
      icon: IconPlane,
      breadcrumbs: true
    },
    {
      id: 'nodes',
      title: '节点管理',
      type: 'item',
      url: '/subscription/nodes',
      icon: IconNetwork,
      breadcrumbs: true
    },
    {
      id: 'node-check',
      title: '节点检测',
      type: 'item',
      url: '/subscription/node-check',
      icon: IconDeviceDesktopAnalytics,
      breadcrumbs: true
    },
    {
      id: 'subs',
      title: '订阅列表',
      type: 'item',
      url: '/subscription/subs',
      icon: IconList,
      breadcrumbs: true
    },
    {
      id: 'templates',
      title: '模板管理',
      type: 'item',
      url: '/subscription/templates',
      icon: IconTemplate,
      breadcrumbs: true
    },
    {
      id: 'tags',
      title: '标签管理',
      type: 'item',
      url: '/subscription/tags',
      icon: IconTags,
      breadcrumbs: true
    }
  ]
};

// ==============================|| SCRIPT MENU ITEMS ||============================== //

const script = {
  id: 'script-group',
  title: '脚本管理',
  type: 'group',
  children: [
    {
      id: 'script',
      title: '脚本列表',
      type: 'item',
      url: '/script',
      icon: IconScript,
      breadcrumbs: true
    }
  ]
};

// ==============================|| ACCESS KEY MENU ITEMS ||============================== //

const accesskey = {
  id: 'accesskey-group',
  title: 'API 密钥',
  type: 'group',
  children: [
    {
      id: 'accesskey',
      title: 'API 密钥',
      type: 'item',
      url: '/accesskey',
      icon: IconKey,
      breadcrumbs: true
    }
  ]
};

// ==============================|| SYSTEM MENU ITEMS ||============================== //

const system = {
  id: 'system',
  title: '系统设置',
  type: 'group',
  children: [
    {
      id: 'tasks',
      title: '任务管理',
      type: 'item',
      url: '/system/tasks',
      icon: IconListCheck,
      breadcrumbs: true
    },
    {
      id: 'hosts',
      title: 'Host管理',
      type: 'item',
      url: '/system/hosts',
      icon: IconWorld,
      breadcrumbs: true
    },
    {
      id: 'monitor',
      title: '系统监控',
      type: 'item',
      url: '/system/monitor',
      icon: IconDeviceDesktopAnalytics,
      breadcrumbs: true
    },
    {
      id: 'user',
      title: '个人中心',
      type: 'item',
      url: '/settings',
      icon: IconSettings,
      breadcrumbs: true
    }
  ]
};

export { subscription, script, accesskey, system };
