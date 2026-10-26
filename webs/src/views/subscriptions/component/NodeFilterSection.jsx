import { useState } from 'react';
import PropTypes from 'prop-types';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { useTheme } from '@mui/material/styles';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

/**
 * 订阅节点筛选分区
 * 统一各筛选条件的折叠块结构，使通用限制、国家、标签、协议与名称筛选保持一致的视觉层级
 */
export default function NodeFilterSection({ icon, title, summary, defaultExpanded = false, children }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, tertiaryText } = getReadableTextTokens(theme, isDark);
  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const [expanded, setExpanded] = useState(defaultExpanded);

  const headerHoverSurface = isDark ? withAlpha(palette.background.paper, 0.2) : withAlpha(palette.primary.main, 0.04);
  const contentSurface = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.08)} 0%, ${dialogSurface} 100%)`
    : 'none';

  const toggleExpanded = () => setExpanded((prev) => !prev);

  const handleHeaderKeyDown = (event) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      toggleExpanded();
    }
  };

  return (
    <Paper
      elevation={0}
      sx={{
        border: '1px solid',
        borderColor: panelBorder,
        borderRadius: 2,
        overflow: 'hidden',
        bgcolor: dialogSurface,
        backgroundImage: dialogSurfaceGradient,
        boxShadow: insetHighlight
      }}
    >
      <Box
        role="button"
        tabIndex={0}
        aria-expanded={expanded}
        onClick={toggleExpanded}
        onKeyDown={handleHeaderKeyDown}
        sx={{
          px: 1.75,
          py: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          bgcolor: expanded ? nestedPanelSurface : mutedPanelSurface,
          borderBottom: expanded ? '1px solid' : 'none',
          borderColor: panelBorder,
          cursor: 'pointer',
          transition: 'background-color 0.2s ease, border-color 0.2s ease',
          '&:hover': {
            bgcolor: expanded ? nestedPanelSurface : headerHoverSurface
          },
          '&:focus-visible': {
            outline: `2px solid ${palette.primary.main}`,
            outlineOffset: -2
          }
        }}
      >
        <Stack direction="row" alignItems="center" spacing={1.25} sx={{ minWidth: 0, flex: 1, flexWrap: 'wrap', rowGap: 0.5 }}>
          {icon}
          <Typography variant="subtitle2" fontWeight={600} sx={{ color: primaryText }}>
            {title}
          </Typography>
          {summary && (
            <Chip
              size="small"
              variant="outlined"
              label={summary}
              sx={{
                height: 22,
                color: tertiaryText,
                bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
                borderColor: withAlpha(palette.primary.main, isDark ? 0.28 : 0.16)
              }}
            />
          )}
        </Stack>
        <Box sx={{ display: 'flex', alignItems: 'center', color: tertiaryText, ml: 1 }}>
          {expanded ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />}
        </Box>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ px: 2.25, py: 2.25, bgcolor: dialogSurface, backgroundImage: contentSurface }}>{children}</Box>
      </Collapse>
    </Paper>
  );
}

NodeFilterSection.propTypes = {
  icon: PropTypes.node,
  title: PropTypes.string.isRequired,
  summary: PropTypes.string,
  defaultExpanded: PropTypes.bool,
  children: PropTypes.node.isRequired
};
