import { useState } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Tooltip from '@mui/material/Tooltip';

// icons
import { IconRobot } from '@tabler/icons-react';

// project imports
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getHeaderTriggerTokens } from '../headerPopoverTokens';
import request from 'api/request';

export default function SkillDownloadSection() {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const [loading, setLoading] = useState(false);

  const accentColor = theme.palette.primary.main;
  const { triggerColor, triggerSurface, triggerBorder, activeSurface, activeBorder } = getHeaderTriggerTokens(theme, isDark, accentColor, {
    lightSurfaceAlpha: 0.12,
    lightHoverAlpha: 0.22,
    activeColor: accentColor
  });

  const handleDownload = async () => {
    if (loading) {
      return;
    }
    setLoading(true);
    try {
      const response = await request({
        url: '/v1/skill/download',
        method: 'get',
        responseType: 'blob'
      });

      const data = response.data || response;
      if (!(data instanceof Blob) || data.size === 0) {
        console.error('Invalid skill zip file');
        return;
      }

      // 从 Content-Disposition 提取文件名，否则用默认值
      let filename = 'skill-sublinkpro.zip';
      const contentDisposition = response.headers?.['content-disposition'];
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="?([^"]+)"?/);
        if (match && match[1]) {
          filename = match[1];
        }
      }

      const url = window.URL.createObjectURL(data);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Skill download failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ ml: 2 }}>
      <Tooltip title={t('skill.downloadTooltip')}>
        <Avatar
          variant="rounded"
          sx={{
            ...theme.typography.commonAvatar,
            ...theme.typography.mediumAvatar,
            transition: 'all .2s ease-in-out',
            color: triggerColor,
            background: triggerSurface,
            border: '1px solid',
            borderColor: triggerBorder,
            cursor: 'pointer',
            '&:hover, &:focus-visible': {
              color: triggerColor,
              background: activeSurface,
              borderColor: activeBorder
            }
          }}
          onClick={handleDownload}
          aria-label={t('skill.downloadTooltip')}
        >
          <IconRobot stroke={1.5} size="20px" />
        </Avatar>
      </Tooltip>
    </Box>
  );
}
