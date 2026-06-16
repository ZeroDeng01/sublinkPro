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
  IconPlane,
  IconFlag
} from '@tabler/icons-react';

// ==============================|| SUBSCRIPTION MENU ITEMS ||============================== //

const subscription = {
  id: 'subscription',
  title: 'Subscription Management',
  titleKey: 'navigation.groups.subscription',
  type: 'group',
  children: [
    {
      id: 'airports',
      title: 'Airport Management',
      titleKey: 'navigation.items.airports',
      type: 'item',
      url: '/subscription/airports',
      icon: IconPlane,
      breadcrumbs: true
    },
    {
      id: 'nodes',
      title: 'Node Management',
      titleKey: 'navigation.items.nodes',
      type: 'item',
      url: '/subscription/nodes',
      icon: IconNetwork,
      breadcrumbs: true
    },
    {
      id: 'node-check',
      title: 'Node Check',
      titleKey: 'navigation.items.nodeCheck',
      type: 'item',
      url: '/subscription/node-check',
      icon: IconDeviceDesktopAnalytics,
      breadcrumbs: true
    },
    {
      id: 'subs',
      title: 'Subscriptions',
      titleKey: 'navigation.items.subscriptions',
      type: 'item',
      url: '/subscription/subs',
      icon: IconList,
      breadcrumbs: true
    },
    {
      id: 'templates',
      title: 'Templates',
      titleKey: 'navigation.items.templates',
      type: 'item',
      url: '/subscription/templates',
      icon: IconTemplate,
      breadcrumbs: true
    },
    {
      id: 'tags',
      title: 'Tags',
      titleKey: 'navigation.items.tags',
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
  title: 'Script Management',
  titleKey: 'navigation.groups.script',
  type: 'group',
  children: [
    {
      id: 'script',
      title: 'Scripts',
      titleKey: 'navigation.items.scripts',
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
  title: 'API Keys',
  titleKey: 'navigation.groups.accessKey',
  type: 'group',
  children: [
    {
      id: 'accesskey',
      title: 'API Keys',
      titleKey: 'navigation.items.accessKeys',
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
  title: 'System Settings',
  titleKey: 'navigation.groups.system',
  type: 'group',
  children: [
    {
      id: 'tasks',
      title: 'Tasks',
      titleKey: 'navigation.items.tasks',
      type: 'item',
      url: '/system/tasks',
      icon: IconListCheck,
      breadcrumbs: true
    },
    {
      id: 'hosts',
      title: 'Host Management',
      titleKey: 'navigation.items.hosts',
      type: 'item',
      url: '/system/hosts',
      icon: IconWorld,
      breadcrumbs: true
    },
    {
      id: 'country-rules',
      title: 'Country Rules',
      titleKey: 'navigation.items.countryRules',
      type: 'item',
      url: '/system/country-rules',
      icon: IconFlag,
      breadcrumbs: true
    },
    {
      id: 'webhooks',
      title: 'Webhooks',
      titleKey: 'navigation.items.webhooks',
      type: 'item',
      url: '/system/webhooks',
      icon: IconList,
      breadcrumbs: true
    },
    {
      id: 'monitor',
      title: 'System Monitor',
      titleKey: 'navigation.items.monitor',
      type: 'item',
      url: '/system/monitor',
      icon: IconDeviceDesktopAnalytics,
      breadcrumbs: true
    },
    {
      id: 'app-settings',
      title: 'Application Settings',
      titleKey: 'navigation.items.appSettings',
      type: 'item',
      url: '/system/settings',
      icon: IconSettings,
      breadcrumbs: true
    }
  ]
};

export { subscription, script, accesskey, system };
