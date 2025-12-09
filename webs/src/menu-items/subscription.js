// assets
import { IconNetwork, IconList, IconTemplate, IconScript, IconKey, IconSettings, IconDeviceDesktopAnalytics } from '@tabler/icons-react';

// ==============================|| SUBSCRIPTION MENU ITEMS ||============================== //

const subscription = {
  id: 'subscription',
  title: '订阅管理',
  type: 'group',
  children: [
    {
      id: 'nodes',
      title: '节点管理',
      type: 'item',
      url: '/subscription/nodes',
      icon: IconNetwork,
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
      id: 'monitor',
      title: '系统监控',
      type: 'item',
      url: '/system/monitor',
      icon: IconDeviceDesktopAnalytics,
      breadcrumbs: true
    },
    {
      id: 'user',
      title: '用户设置',
      type: 'item',
      url: '/system/user',
      icon: IconSettings,
      breadcrumbs: true
    }
  ]
};

export { subscription, script, accesskey, system };

