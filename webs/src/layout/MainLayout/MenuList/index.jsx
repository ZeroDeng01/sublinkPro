import { Activity, memo, useMemo, useState } from 'react';

import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

// project imports
import NavItem from './NavItem';
import NavGroup from './NavGroup';
import menuItems from 'menu-items';

import { useGetMenuMaster } from 'api/menu';
import { useAuth } from 'contexts/AuthContext';

// ==============================|| SIDEBAR MENU LIST ||============================== //

function filterMenuItemsByRole(items, isAdmin) {
  return items
    .filter((item) => isAdmin || !item.adminOnly)
    .map((item) => {
      if (!item.children) return item;
      return { ...item, children: filterMenuItemsByRole(item.children, isAdmin) };
    })
    .filter((item) => !item.children || item.children.length > 0);
}

function MenuList() {
  const { menuMaster } = useGetMenuMaster();
  const { user } = useAuth();
  const drawerOpen = menuMaster.isDashboardDrawerOpened;
  const isAdmin = [user?.role, ...(user?.roles || [])].some((role) => String(role || '').toLowerCase() === 'admin');
  const visibleMenuItems = useMemo(() => filterMenuItemsByRole(menuItems.items, isAdmin), [isAdmin]);

  const [selectedID, setSelectedID] = useState('');

  const lastItem = null;

  let lastItemIndex = visibleMenuItems.length - 1;
  let remItems = [];
  let lastItemId;

  if (lastItem && lastItem < visibleMenuItems.length) {
    lastItemId = visibleMenuItems[lastItem - 1].id;
    lastItemIndex = lastItem - 1;
    remItems = visibleMenuItems.slice(lastItem - 1, visibleMenuItems.length).map((item) => ({
      title: item.title,
      elements: item.children,
      icon: item.icon,
      ...(item.url && {
        url: item.url
      })
    }));
  }

  const navItems = visibleMenuItems.slice(0, lastItemIndex + 1).map((item, index) => {
    switch (item.type) {
      case 'group':
        if (item.url && item.id !== lastItemId) {
          return (
            <List key={item.id}>
              <NavItem item={item} level={1} isParents setSelectedID={() => setSelectedID('')} />
              <Activity mode={index !== 0 ? 'visible' : 'hidden'}>
                <Divider sx={{ py: 0.5 }} />
              </Activity>
            </List>
          );
        }

        return (
          <NavGroup
            key={item.id}
            setSelectedID={setSelectedID}
            selectedID={selectedID}
            item={item}
            lastItem={lastItem}
            remItems={remItems}
            lastItemId={lastItemId}
          />
        );
      default:
        return (
          <Typography key={item.id} variant="h6" align="center" sx={{ color: 'error.main' }}>
            Menu Items Error
          </Typography>
        );
    }
  });

  return <Box {...(drawerOpen && { sx: { mt: 1.5 } })}>{navItems}</Box>;
}

export default memo(MenuList);
