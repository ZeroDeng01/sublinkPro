import PropTypes from 'prop-types';

// material-ui
import { alpha, useTheme } from '@mui/material/styles';
import Chip from '@mui/material/Chip';
import Tooltip from '@mui/material/Tooltip';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// utils
import { resolveProtocolPresentationFromLink } from 'utils/protocolPresentation';
import { getNodeThemeTokens } from '../nodeTheme';

export default function NodeProtocolChip({ link, protocolMeta, maxWidth = 92 }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const protocol = resolveProtocolPresentationFromLink(link, protocolMeta);
  const chipColor = protocol.color || tokens.palette.primary.main;

  return (
    <Tooltip title={`协议：${protocol.label}`}>
      <Chip
        label={protocol.label}
        size="small"
        sx={{
          maxWidth,
          color: chipColor,
          bgcolor: alpha(chipColor, isDark ? 0.14 : 0.08),
          border: '1px solid',
          borderColor: alpha(chipColor, isDark ? 0.42 : 0.28),
          '& .MuiChip-label': {
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            fontWeight: 700,
            letterSpacing: 0.2
          }
        }}
      />
    </Tooltip>
  );
}

NodeProtocolChip.propTypes = {
  link: PropTypes.string,
  protocolMeta: PropTypes.array,
  maxWidth: PropTypes.oneOfType([PropTypes.number, PropTypes.string])
};
