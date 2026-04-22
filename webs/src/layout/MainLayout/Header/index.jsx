// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';

// project imports
import LogoSection from '../LogoSection';
import ProfileSection from './ProfileSection';
import NotificationSection from './NotificationSection';
import DonationSection from './DonationSection';
import ThemeModeSection from './ThemeModeSection';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getHeaderTriggerTokens } from './headerPopoverTokens';
import { donationConfig } from 'config/donation';

import { handlerDrawerOpen, useGetMenuMaster } from 'api/menu';

// assets
import { IconLayoutSidebarLeftCollapse, IconLayoutSidebarLeftExpand } from '@tabler/icons-react';

// ==============================|| MAIN NAVBAR / HEADER ||============================== //

export default function Header() {
  const theme = useTheme();
  const downMD = useMediaQuery(theme.breakpoints.down('md'));
  const { isDark } = useResolvedColorScheme();

  const { menuMaster } = useGetMenuMaster();
  const drawerOpen = menuMaster.isDashboardDrawerOpened;
  const sidebarAccent = theme.palette[donationConfig.headerIconColor].main;
  const { triggerColor, triggerSurface, triggerBorder, activeSurface, activeBorder } = getHeaderTriggerTokens(
    theme,
    isDark,
    sidebarAccent,
    {
      lightSurfaceAlpha: 0.12,
      lightHoverAlpha: 0.22,
      activeColor: sidebarAccent
    }
  );
  const MenuToggleIcon = drawerOpen ? IconLayoutSidebarLeftCollapse : IconLayoutSidebarLeftExpand;

  return (
    <>
      {/* logo & toggler button */}
      <Box sx={{ width: downMD ? 'auto' : 228, display: 'flex' }}>
        <Box component="span" sx={{ display: { xs: 'none', md: 'block' }, flexGrow: 1 }}>
          <LogoSection />
        </Box>
        <Avatar
          variant="rounded"
          sx={{
            ...theme.typography.commonAvatar,
            ...theme.typography.mediumAvatar,
            overflow: 'hidden',
            transition: 'all .2s ease-in-out',
            color: triggerColor,
            background: triggerSurface,
            border: '1px solid',
            borderColor: triggerBorder,
            '&:hover, &:focus-visible': {
              color: triggerColor,
              background: activeSurface,
              borderColor: activeBorder
            }
          }}
          onClick={() => handlerDrawerOpen(!drawerOpen)}
        >
          <MenuToggleIcon stroke={1.5} size="20px" />
        </Avatar>
      </Box>

      {/* spacer */}
      <Box sx={{ flexGrow: 1 }} />

      {/* donation */}
      <DonationSection />

      {/* notification */}
      <NotificationSection />

      <ThemeModeSection />

      {/* profile */}
      <ProfileSection />
    </>
  );
}
