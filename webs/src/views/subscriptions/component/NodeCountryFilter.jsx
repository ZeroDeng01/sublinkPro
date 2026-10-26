import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { useTheme } from '@mui/material/styles';
import PublicIcon from '@mui/icons-material/Public';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import BlockIcon from '@mui/icons-material/Block';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import CountryFlag from 'components/CountryFlag';
import NodeFilterSection from './NodeFilterSection';

const normalizeCountryCode = (value) => (typeof value === 'string' ? value.trim().toUpperCase() : '');

const normalizeCountryCodeList = (values) => {
  if (!Array.isArray(values)) return [];
  return Array.from(new Set(values.map((value) => normalizeCountryCode(value)).filter(Boolean)));
};

/**
 * 订阅节点国家筛选
 * 将落地 IP 国家白名单/黑名单独立为单独筛选分区
 */
export default function NodeCountryFilter({
  countryOptions = [],
  whitelistValue = [],
  blacklistValue = [],
  onWhitelistChange,
  onBlacklistChange
}) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { secondaryText } = getReadableTextTokens(theme, isDark);
  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const [whitelistInput, setWhitelistInput] = useState('');
  const [blacklistInput, setBlacklistInput] = useState('');

  const whitelist = normalizeCountryCodeList(whitelistValue);
  const blacklist = normalizeCountryCodeList(blacklistValue);
  const hasRules = whitelist.length > 0 || blacklist.length > 0;

  const commitInput = (field, inputValue, setInputValue) => {
    const code = normalizeCountryCode(inputValue);
    if (!code) return false;
    const current = field === 'whitelist' ? whitelist : blacklist;
    const next = normalizeCountryCodeList([...current, code]);
    if (field === 'whitelist') {
      onWhitelistChange(next);
    } else {
      onBlacklistChange(next);
    }
    setInputValue('');
    return true;
  };

  const handleKeyDown = (event, field, inputValue, setInputValue) => {
    if (event.key !== 'Enter') return;
    if (commitInput(field, inputValue, setInputValue)) {
      event.preventDefault();
    }
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

  const renderOption = (props, option) => {
    const { key, ...otherProps } = props;
    return (
      <Box component="li" {...otherProps} key={key} sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <CountryFlag code={option} height={12} />
        {normalizeCountryCode(option)}
      </Box>
    );
  };

  const renderTags = (value, getTagProps) =>
    value.map((option, index) => {
      const { key, ...tagProps } = getTagProps({ index });
      return (
        <Chip
          key={key}
          label={
            <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.5 }}>
              <CountryFlag code={option} height={11} />
              {normalizeCountryCode(option)}
            </Box>
          }
          size="small"
          sx={{
            bgcolor: isDark ? withAlpha(palette.primary.main, 0.12) : undefined,
            borderColor: isDark ? withAlpha(palette.primary.main, 0.3) : undefined,
            '& .MuiChip-label': { display: 'flex', alignItems: 'center', px: 1 }
          }}
          {...tagProps}
        />
      );
    });

  return (
    <NodeFilterSection
      icon={<PublicIcon color="primary" fontSize="small" />}
      title={t('subscriptions.form.nodeFilter.countryTitle')}
      summary={
        hasRules ? t('subscriptions.form.nodeFilter.countrySummary', { whitelist: whitelist.length, blacklist: blacklist.length }) : ''
      }
    >
      <Stack spacing={2.25}>
        <Box sx={sectionCardSx}>
          <Stack spacing={1.25}>
            <Stack direction="row" alignItems="center" spacing={1}>
              <CheckCircleOutlineIcon color="success" fontSize="small" />
              <Typography variant="subtitle2" fontWeight={600} sx={{ color: theme.palette.success.main }}>
                {t('subscriptions.form.nodeFilter.countryWhitelist')}
              </Typography>
            </Stack>
            <Autocomplete
              multiple
              freeSolo
              size="small"
              options={countryOptions}
              value={whitelist}
              inputValue={whitelistInput}
              onInputChange={(event, newInputValue) => {
                void event;
                setWhitelistInput(newInputValue);
              }}
              onChange={(event, newValue) => {
                void event;
                onWhitelistChange(normalizeCountryCodeList(newValue));
                setWhitelistInput('');
              }}
              getOptionLabel={(option) => normalizeCountryCode(option)}
              isOptionEqualToValue={(option, value) => normalizeCountryCode(option) === normalizeCountryCode(value)}
              filterSelectedOptions
              sx={autocompleteSx}
              renderOption={renderOption}
              renderTags={renderTags}
              renderInput={(params) => (
                <TextField {...params} onKeyDown={(event) => handleKeyDown(event, 'whitelist', whitelistInput, setWhitelistInput)} />
              )}
            />
            <Typography variant="caption" sx={{ color: secondaryText }}>
              {t('subscriptions.form.nodeFilter.countryWhitelistHelper')}
            </Typography>
          </Stack>
        </Box>

        <Box sx={sectionCardSx}>
          <Stack spacing={1.25}>
            <Stack direction="row" alignItems="center" spacing={1}>
              <BlockIcon color="error" fontSize="small" />
              <Typography variant="subtitle2" fontWeight={600} sx={{ color: theme.palette.error.main }}>
                {t('subscriptions.form.nodeFilter.countryBlacklist')}
              </Typography>
            </Stack>
            <Autocomplete
              multiple
              freeSolo
              size="small"
              options={countryOptions}
              value={blacklist}
              inputValue={blacklistInput}
              onInputChange={(event, newInputValue) => {
                void event;
                setBlacklistInput(newInputValue);
              }}
              onChange={(event, newValue) => {
                void event;
                onBlacklistChange(normalizeCountryCodeList(newValue));
                setBlacklistInput('');
              }}
              getOptionLabel={(option) => normalizeCountryCode(option)}
              isOptionEqualToValue={(option, value) => normalizeCountryCode(option) === normalizeCountryCode(value)}
              filterSelectedOptions
              sx={autocompleteSx}
              renderOption={renderOption}
              renderTags={renderTags}
              renderInput={(params) => (
                <TextField {...params} onKeyDown={(event) => handleKeyDown(event, 'blacklist', blacklistInput, setBlacklistInput)} />
              )}
            />
            <Typography variant="caption" sx={{ color: secondaryText }}>
              {t('subscriptions.form.nodeFilter.countryBlacklistHelper')}
            </Typography>
          </Stack>
        </Box>
      </Stack>
    </NodeFilterSection>
  );
}

NodeCountryFilter.propTypes = {
  countryOptions: PropTypes.arrayOf(PropTypes.string),
  whitelistValue: PropTypes.arrayOf(PropTypes.string),
  blacklistValue: PropTypes.arrayOf(PropTypes.string),
  onWhitelistChange: PropTypes.func.isRequired,
  onBlacklistChange: PropTypes.func.isRequired
};
