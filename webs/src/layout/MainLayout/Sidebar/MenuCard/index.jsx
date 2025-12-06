import { memo, useState, useEffect } from 'react';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import Card from '@mui/material/Card';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemAvatar from '@mui/material/ListItemAvatar';
import ListItemText from '@mui/material/ListItemText';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import Button from '@mui/material/Button';

// assets
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import NewReleasesIcon from '@mui/icons-material/NewReleases';

// project imports
import useConfig from 'hooks/useConfig';

// GitHub 仓库配置
const GITHUB_REPO = 'ZeroDeng01/sublinkPro';
const GITHUB_URL = `https://github.com/${GITHUB_REPO}`;
const GITHUB_API_RELEASES = `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`;

// ==============================|| SIDEBAR - VERSION CARD ||============================== //

function MenuCard() {
  const theme = useTheme();
  const { version } = useConfig();
  const [latestVersion, setLatestVersion] = useState('');
  const [hasUpdate, setHasUpdate] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchLatestVersion = async () => {
      setLoading(true);
      try {
        const res = await fetch(GITHUB_API_RELEASES);
        if (res.ok) {
          const data = await res.json();
          setLatestVersion(data.tag_name || '');
        }
      } catch (error) {
        console.error('获取版本信息失败:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchLatestVersion();
  }, []);

  useEffect(() => {
    if (latestVersion && version) {
      const current = version.replace(/^v/, '');
      const latest = latestVersion.replace(/^v/, '');

      if (latest && current && latest !== current) {
        setHasUpdate(true);
      }
    }
  }, [version, latestVersion]);

  return (
    <Card
      sx={{
        bgcolor: hasUpdate ? alpha(theme.palette.warning.main, 0.1) : 'primary.light',
        mb: 2.75,
        overflow: 'hidden',
        position: 'relative',
        border: hasUpdate ? `1px solid ${alpha(theme.palette.warning.main, 0.3)}` : 'none',
        '&:after': {
          content: '""',
          position: 'absolute',
          width: 130,
          height: 130,
          bgcolor: hasUpdate ? alpha(theme.palette.warning.main, 0.1) : 'primary.200',
          borderRadius: '50%',
          top: -85,
          right: -76
        }
      }}
    >
      <Box sx={{ p: 2 }}>
        <List disablePadding sx={{ pb: 1 }}>
          <ListItem alignItems="flex-start" disableGutters disablePadding>
            <ListItemAvatar sx={{ mt: 0 }}>
              <Avatar
                variant="rounded"
                sx={{
                  ...theme.typography.largeAvatar,
                  borderRadius: 2,
                  color: hasUpdate ? 'warning.main' : 'primary.main',
                  border: 'none',
                  bgcolor: 'background.paper',
                  boxShadow: theme.shadows[2]
                }}
              >
                {hasUpdate ? <NewReleasesIcon fontSize="inherit" /> : <InfoOutlinedIcon fontSize="inherit" />}
              </Avatar>
            </ListItemAvatar>
            <ListItemText
              sx={{ mt: 0 }}
              primary={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Typography
                    variant="subtitle1"
                    sx={{
                      color: hasUpdate ? 'warning.dark' : 'primary.800',
                      fontWeight: 600
                    }}
                  >
                    SublinkPro
                  </Typography>
                  <Tooltip title="查看 GitHub">
                    <IconButton size="small" component="a" href={GITHUB_URL} target="_blank" rel="noopener noreferrer" sx={{ p: 0.5 }}>
                      <OpenInNewIcon fontSize="small" sx={{ fontSize: 14 }} />
                    </IconButton>
                  </Tooltip>
                </Box>
              }
              secondary={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                  <Chip
                    label={loading ? '加载中...' : version || 'dev'}
                    size="small"
                    color="primary"
                    variant="outlined"
                    sx={{ height: 20, fontSize: '0.7rem' }}
                  />
                </Box>
              }
              secondaryTypographyProps={{ component: 'div' }}
            />
          </ListItem>
        </List>
        {hasUpdate && latestVersion && (
          <Box sx={{ mt: 1 }}>
            <Typography variant="caption" sx={{ color: 'warning.dark', display: 'block', mb: 0.5 }}>
              💡 发现新版本，建议更新
            </Typography>
            <Button
              size="small"
              variant="contained"
              color="warning"
              href={`${GITHUB_URL}/releases/tag/${latestVersion}`}
              target="_blank"
              rel="noopener noreferrer"
              sx={{ fontSize: '0.7rem', py: 0.25, px: 1 }}
            >
              {latestVersion} 可用
            </Button>
          </Box>
        )}
        {!hasUpdate && latestVersion && (
          <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block' }}>
            最新版本: {latestVersion}
          </Typography>
        )}
      </Box>
    </Card>
  );
}

export default memo(MenuCard);
