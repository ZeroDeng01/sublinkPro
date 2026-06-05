import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';

// Icons
import SortIcon from '@mui/icons-material/Sort';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import VerticalAlignTopIcon from '@mui/icons-material/VerticalAlignTop';
import VerticalAlignBottomIcon from '@mui/icons-material/VerticalAlignBottom';
import MoveDownIcon from '@mui/icons-material/MoveDown';
import ClearAllIcon from '@mui/icons-material/ClearAll';
import SourceIcon from '@mui/icons-material/Source';
import AbcIcon from '@mui/icons-material/Abc';
import RouterIcon from '@mui/icons-material/Router';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';
import PublicIcon from '@mui/icons-material/Public';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

/**
 * Sort toolbar component
 * Provides quick sort and batch move functions
 */
export default function SortToolbar({ selectedItems = [], onBatchSort, onBatchMove, onClearSelection, totalItems = 0 }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const { palette, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);

  // Sort order
  const [sortOrder, setSortOrder] = useState('asc');
  // Move position input
  const [movePosition, setMovePosition] = useState('');
  // Sort menu anchor
  const [sortMenuAnchor, setSortMenuAnchor] = useState(null);

  // Sort options
  const sortOptions = [
    { value: 'source', label: t('subscriptions.sort.options.source'), icon: <SourceIcon fontSize="small" /> },
    { value: 'name', label: t('subscriptions.sort.options.name'), icon: <AbcIcon fontSize="small" /> },
    { value: 'protocol', label: t('subscriptions.sort.options.protocol'), icon: <RouterIcon fontSize="small" /> },
    { value: 'delay', label: t('subscriptions.sort.options.delay'), icon: <TimerIcon fontSize="small" /> },
    { value: 'speed', label: t('subscriptions.sort.options.speed'), icon: <SpeedIcon fontSize="small" /> },
    { value: 'country', label: t('subscriptions.sort.options.country'), icon: <PublicIcon fontSize="small" /> }
  ];

  // Handle batch sort
  const handleBatchSort = (sortBy) => {
    if (onBatchSort) {
      onBatchSort(sortBy, sortOrder);
    }
    setSortMenuAnchor(null);
  };

  // Handle batch move
  const handleBatchMove = (position) => {
    if (onBatchMove && selectedItems.length > 0) {
      onBatchMove(position);
    }
  };

  // Handle move to specific position
  const handleMoveToPosition = () => {
    const pos = parseInt(movePosition, 10);
    if (!isNaN(pos) && pos >= 1 && pos <= totalItems) {
      handleBatchMove(pos - 1); // convert to 0-indexed
      setMovePosition('');
    }
  };

  const hasSelection = selectedItems.length > 0;
  const shellInset = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const accentSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.06);
  const accentHoverSurface = withAlpha(palette.primary.main, isDark ? 0.18 : 0.1);
  const accentBorder = withAlpha(palette.primary.main, isDark ? 0.3 : 0.18);
  const strongAccentBorder = withAlpha(palette.primary.main, isDark ? 0.42 : 0.26);
  const errorSurface = withAlpha(palette.error.main, isDark ? 0.12 : 0.06);
  const errorBorder = withAlpha(palette.error.main, isDark ? 0.34 : 0.18);
  const neutralBorder = withAlpha(palette.divider, isDark ? 0.66 : 0.84);
  const compactFieldSx = {
    width: 82,
    '& .MuiOutlinedInput-root': {
      borderRadius: 2,
      bgcolor: mutedPanelSurface,
      border: '1px solid',
      borderColor: neutralBorder,
      boxShadow: shellInset,
      '& fieldset': {
        borderColor: 'transparent'
      },
      '&:hover': {
        '& fieldset': {
          borderColor: accentBorder
        }
      },
      '&.Mui-focused': {
        bgcolor: nestedPanelSurface,
        '& fieldset': {
          borderColor: strongAccentBorder
        }
      }
    },
    '& .MuiInputBase-input': {
      color: primaryText,
      textAlign: 'center'
    },
    '& .MuiInputBase-input::placeholder': {
      color: secondaryText,
      opacity: 1
    }
  };
  const actionButtonSx = {
    borderRadius: 2,
    borderColor: accentBorder,
    bgcolor: nestedPanelSurface,
    color: 'primary.main',
    fontWeight: 700,
    '&:hover': {
      borderColor: strongAccentBorder,
      bgcolor: accentSurface
    }
  };

  return (
    <Box
      sx={{
        p: 1.5,
        mb: 1.25,
        borderRadius: 2.5,
        bgcolor: mutedPanelSurface,
        border: '1px solid',
        borderColor: panelBorder,
        boxShadow: shellInset
      }}
    >
      <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" sx={{ mb: hasSelection ? 1 : 0 }}>
        <Tooltip title={t('subscriptions.sort.quickSort')}>
          <Chip
            icon={<SortIcon />}
            label={t('subscriptions.sort.quickSort')}
            size="small"
            onClick={(e) => setSortMenuAnchor(e.currentTarget)}
            sx={{
              cursor: 'pointer',
              bgcolor: accentSurface,
              color: 'primary.main',
              border: '1px solid',
              borderColor: accentBorder,
              fontWeight: 700,
              '&:hover': {
                bgcolor: accentHoverSurface,
                borderColor: strongAccentBorder
              },
              '& .MuiChip-icon': {
                color: 'inherit'
              }
            }}
          />
        </Tooltip>

        <Tooltip title={sortOrder === 'asc' ? t('subscriptions.sort.asc') : t('subscriptions.sort.desc')}>
          <IconButton
            size="small"
            onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
            sx={{
              bgcolor: sortOrder === 'asc' ? accentSurface : errorSurface,
              color: sortOrder === 'asc' ? 'primary.main' : 'error.main',
              border: '1px solid',
              borderColor: sortOrder === 'asc' ? accentBorder : errorBorder,
              '&:hover': {
                bgcolor: sortOrder === 'asc' ? accentHoverSurface : withAlpha(palette.error.main, isDark ? 0.18 : 0.1),
                borderColor: sortOrder === 'asc' ? strongAccentBorder : withAlpha(palette.error.main, isDark ? 0.44 : 0.24)
              }
            }}
          >
            {sortOrder === 'asc' ? <ArrowUpwardIcon fontSize="small" /> : <ArrowDownwardIcon fontSize="small" />}
          </IconButton>
        </Tooltip>

        <Menu
          anchorEl={sortMenuAnchor}
          open={Boolean(sortMenuAnchor)}
          onClose={() => setSortMenuAnchor(null)}
          slotProps={{
            paper: {
              sx: {
                mt: 0.5,
                bgcolor: nestedPanelSurface,
                border: '1px solid',
                borderColor: panelBorder,
                boxShadow: shellInset
              }
            }
          }}
        >
          {sortOptions.map((opt) => (
            <MenuItem
              key={opt.value}
              onClick={() => handleBatchSort(opt.value)}
              sx={{
                borderRadius: 1.5,
                mx: 0.5,
                my: 0.25,
                color: primaryText,
                '&:hover': {
                  bgcolor: accentSurface
                }
              }}
            >
              <ListItemIcon sx={{ color: secondaryText }}>{opt.icon}</ListItemIcon>
              <ListItemText>{opt.label}</ListItemText>
            </MenuItem>
          ))}
        </Menu>

        {hasSelection && (
          <>
            <Divider orientation="vertical" flexItem sx={{ mx: 0.5, borderColor: neutralBorder }} />
            <Typography variant="caption" sx={{ color: secondaryText, fontWeight: 600 }}>
              {t('subscriptions.sort.selectedCount', { count: selectedItems.length })}
            </Typography>
          </>
        )}
      </Stack>

      {hasSelection && (
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <Tooltip title={t('subscriptions.sort.moveTop')}>
            <Button
              size="small"
              variant="outlined"
              startIcon={<VerticalAlignTopIcon />}
              onClick={() => handleBatchMove(0)}
              sx={actionButtonSx}
            >
              {t('subscriptions.sort.top')}
            </Button>
          </Tooltip>

          <Tooltip title={t('subscriptions.sort.moveBottom')}>
            <Button
              size="small"
              variant="outlined"
              startIcon={<VerticalAlignBottomIcon />}
              onClick={() => handleBatchMove(totalItems - 1)}
              sx={actionButtonSx}
            >
              {t('subscriptions.sort.bottom')}
            </Button>
          </Tooltip>

          <Divider orientation="vertical" flexItem sx={{ borderColor: neutralBorder }} />

          <TextField
            size="small"
            type="number"
            placeholder={t('subscriptions.sort.position')}
            value={movePosition}
            onChange={(e) => setMovePosition(e.target.value)}
            sx={compactFieldSx}
            slotProps={{
              htmlInput: { min: 1, max: totalItems }
            }}
          />
          <Tooltip title={t('subscriptions.sort.moveToPosition')}>
            <IconButton
              size="small"
              onClick={handleMoveToPosition}
              disabled={!movePosition}
              sx={{
                bgcolor: accentSurface,
                color: 'primary.main',
                border: '1px solid',
                borderColor: accentBorder,
                '&:hover': {
                  bgcolor: accentHoverSurface,
                  borderColor: strongAccentBorder
                },
                '&.Mui-disabled': {
                  bgcolor: 'action.disabledBackground',
                  color: 'text.disabled',
                  borderColor: 'transparent'
                }
              }}
            >
              <MoveDownIcon />
            </IconButton>
          </Tooltip>

          <Divider orientation="vertical" flexItem sx={{ borderColor: neutralBorder }} />

          <Tooltip title={t('subscriptions.sort.clearSelection')}>
            <Button
              size="small"
              variant="outlined"
              startIcon={<ClearAllIcon />}
              onClick={onClearSelection}
              sx={{
                borderRadius: 2,
                borderColor: neutralBorder,
                bgcolor: nestedPanelSurface,
                color: secondaryText,
                fontWeight: 600,
                '&:hover': {
                  borderColor: withAlpha(palette.error.main, isDark ? 0.32 : 0.18),
                  bgcolor: withAlpha(palette.error.main, isDark ? 0.1 : 0.05),
                  color: 'error.main'
                }
              }}
            >
              {t('subscriptions.sort.clearSelection')}
            </Button>
          </Tooltip>
        </Stack>
      )}
    </Box>
  );
}
