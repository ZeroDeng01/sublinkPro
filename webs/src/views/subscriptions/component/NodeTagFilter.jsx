import { useState } from 'react';
import PropTypes from 'prop-types';

import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import { useTheme } from '@mui/material/styles';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import BlockIcon from '@mui/icons-material/Block';

export default function NodeTagFilter({ tagOptions, whitelistValue, blacklistValue, onWhitelistChange, onBlacklistChange }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);
  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const [expanded, setExpanded] = useState(false);

  const parseTagString = (str) => {
    if (!str) return [];
    return str
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t);
  };

  const toTagString = (arr) => arr.map((t) => t.name || t).join(',');

  const whitelistTags = parseTagString(whitelistValue);
  const blacklistTags = parseTagString(blacklistValue);
  const hasAnyRules = whitelistTags.length > 0 || blacklistTags.length > 0;

  const getTagOption = (tagName) => {
    return (tagOptions || []).find((t) => t.name === tagName) || { name: tagName, color: palette.primary.main };
  };

  const headerHoverSurface = isDark ? withAlpha(palette.background.paper, 0.2) : withAlpha(palette.primary.main, 0.04);
  const contentSurface = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.08)} 0%, ${dialogSurface} 100%)`
    : 'none';
  const infoAlertSx = {
    borderColor: withAlpha(palette.info.main, isDark ? 0.28 : 0.18),
    bgcolor: withAlpha(palette.info.main, isDark ? 0.12 : 0.05),
    boxShadow: insetHighlight,
    '& .MuiAlert-icon, & .MuiAlert-message': {
      color: primaryText
    }
  };

  const getTagChipSx = (color, fallbackColor) => {
    const resolvedColor = color || fallbackColor;
    return {
      bgcolor: withAlpha(resolvedColor, isDark ? 0.18 : 0.1),
      color: resolvedColor,
      border: '1px solid',
      borderColor: withAlpha(resolvedColor, isDark ? 0.34 : 0.2),
      '& .MuiChip-deleteIcon': {
        color: withAlpha(resolvedColor, 0.72),
        '&:hover': {
          color: resolvedColor
        }
      }
    };
  };

  const sectionCardSx = {
    p: 1.75,
    borderRadius: 2,
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: insetHighlight
  };

  const autocompleteSx = {
    '& .MuiOutlinedInput-root': {
      bgcolor: mutedPanelSurface,
      transition: 'background-color 0.2s ease, border-color 0.2s ease',
      '& fieldset': {
        borderColor: panelBorder
      },
      '&:hover fieldset': {
        borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.22)
      },
      '&.Mui-focused': {
        bgcolor: dialogSurface
      }
    },
    '& .MuiAutocomplete-tag': {
      maxWidth: '100%'
    }
  };

  const renderOption = (props, option) => (
    <Box component="li" {...props} key={option.name} sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Box
        sx={{
          width: 10,
          height: 10,
          borderRadius: '50%',
          bgcolor: option.color || palette.primary.main,
          flexShrink: 0,
          boxShadow: `0 0 0 1px ${withAlpha(option.color || palette.primary.main, 0.22)}`
        }}
      />
      {option.name}
    </Box>
  );

  return (
    <Paper
      elevation={0}
      sx={{
        mb: 0,
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
          }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Stack direction="row" alignItems="center" spacing={1.25} sx={{ minWidth: 0, flex: 1 }}>
          <LocalOfferIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600} sx={{ color: primaryText }}>
            节点标签过滤
          </Typography>
          {hasAnyRules && (
            <Chip
              size="small"
              variant="outlined"
              label={`白 ${whitelistTags.length} / 黑 ${blacklistTags.length}`}
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
        <Box sx={{ px: 2.25, py: 2.25, bgcolor: dialogSurface, backgroundImage: contentSurface }}>
          <Stack spacing={2.25}>
            <Typography variant="body2" sx={{ color: secondaryText }}>
              直接复用节点标签体系控制订阅输出，适合把已有标签规则接到白名单和黑名单过滤上。
            </Typography>

            <Alert variant="outlined" severity="info" sx={infoAlertSx}>
              <Typography variant="body2" sx={{ color: secondaryText }}>
                按节点标签过滤。<strong>黑名单优先级高于白名单</strong>
                ：匹配黑名单标签的节点会被排除，然后必须匹配白名单标签的节点才会保留。
              </Typography>
            </Alert>

            <Box sx={sectionCardSx}>
              <Stack spacing={1.25}>
                <Stack direction="row" alignItems="center" spacing={1}>
                  <CheckCircleOutlineIcon color="success" fontSize="small" />
                  <Typography variant="subtitle2" fontWeight={600} sx={{ color: theme.palette.success.main }}>
                    白名单标签
                  </Typography>
                </Stack>
                <Autocomplete
                  multiple
                  options={tagOptions || []}
                  getOptionLabel={(option) => option.name || option}
                  value={whitelistTags.map(getTagOption)}
                  onChange={(_, newValue) => onWhitelistChange(toTagString(newValue))}
                  isOptionEqualToValue={(option, value) => (option.name || option) === (value.name || value)}
                  filterSelectedOptions
                  sx={autocompleteSx}
                  renderTags={(value, getTagProps) =>
                    value.map((option, index) => {
                      const { key, ...tagProps } = getTagProps({ index });
                      return (
                        <Chip
                          key={key}
                          label={option.name || option}
                          size="small"
                          sx={getTagChipSx(option.color, theme.palette.success.main)}
                          {...tagProps}
                        />
                      );
                    })
                  }
                  renderOption={renderOption}
                  renderInput={(params) => <TextField {...params} placeholder="选择白名单标签（节点必须包含其中一个标签）" size="small" />}
                />
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  节点必须包含白名单中的至少一个标签才会被保留。
                </Typography>
              </Stack>
            </Box>

            <Box sx={sectionCardSx}>
              <Stack spacing={1.25}>
                <Stack direction="row" alignItems="center" spacing={1}>
                  <BlockIcon color="error" fontSize="small" />
                  <Typography variant="subtitle2" fontWeight={600} sx={{ color: theme.palette.error.main }}>
                    黑名单标签
                  </Typography>
                </Stack>
                <Autocomplete
                  multiple
                  options={tagOptions || []}
                  getOptionLabel={(option) => option.name || option}
                  value={blacklistTags.map(getTagOption)}
                  onChange={(_, newValue) => onBlacklistChange(toTagString(newValue))}
                  isOptionEqualToValue={(option, value) => (option.name || option) === (value.name || value)}
                  filterSelectedOptions
                  sx={autocompleteSx}
                  renderTags={(value, getTagProps) =>
                    value.map((option, index) => {
                      const { key, ...tagProps } = getTagProps({ index });
                      return (
                        <Chip
                          key={key}
                          label={option.name || option}
                          size="small"
                          sx={getTagChipSx(option.color, theme.palette.error.main)}
                          {...tagProps}
                        />
                      );
                    })
                  }
                  renderOption={renderOption}
                  renderInput={(params) => (
                    <TextField {...params} placeholder="选择黑名单标签（包含这些标签的节点将被排除）" size="small" />
                  )}
                />
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  包含黑名单标签的节点会被排除。
                </Typography>
              </Stack>
            </Box>
          </Stack>
        </Box>
      </Collapse>
    </Paper>
  );
}

NodeTagFilter.propTypes = {
  tagOptions: PropTypes.array,
  whitelistValue: PropTypes.string,
  blacklistValue: PropTypes.string,
  onWhitelistChange: PropTypes.func.isRequired,
  onBlacklistChange: PropTypes.func.isRequired
};
