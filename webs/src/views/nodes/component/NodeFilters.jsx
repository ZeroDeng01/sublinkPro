import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import IconButton from '@mui/material/IconButton';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// utils
import {
  createEmptyUnlockRule,
  formatCountry,
  formatUnlockProviderLabel,
  getUnlockProviderOptions,
  getUnlockRuleModeOptions,
  getUnlockStatusOptions,
  QUALITY_STATUS_OPTIONS,
  STATUS_OPTIONS
} from '../utils';
import { getNodeActionButtonSx, getNodeColorChipSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

const UNGROUPED_GROUP_VALUE = '\u672a\u5206\u7ec4';

export default function NodeFilters({
  searchQuery,
  setSearchQuery,
  groupFilter,
  setGroupFilter,
  sourceFilter,
  setSourceFilter,
  maxDelay,
  setMaxDelay,
  minSpeed,
  setMinSpeed,
  maxFraudScore,
  setMaxFraudScore,
  speedStatusFilter,
  setSpeedStatusFilter,
  delayStatusFilter,
  setDelayStatusFilter,
  residentialType,
  setResidentialType,
  ipType,
  setIpType,
  qualityStatus,
  setQualityStatus,
  unlockRules,
  setUnlockRules,
  unlockRuleMode,
  setUnlockRuleMode,
  countryFilter,
  setCountryFilter,
  tagFilter,
  setTagFilter,
  protocolFilter,
  setProtocolFilter,
  groupOptions,
  sourceOptions,
  countryOptions,
  tagOptions,
  protocolOptions,
  onReset
}) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens);
  const unlockProviderOptions = getUnlockProviderOptions();
  const normalizedUnlockRules = Array.isArray(unlockRules) ? unlockRules : [];
  const [unlockExpanded, setUnlockExpanded] = useState(false);
  const allLabel = t('common.all');

  const updateUnlockRule = (index, patch) => {
    setUnlockRules(normalizedUnlockRules.map((rule, ruleIndex) => (ruleIndex === index ? { ...rule, ...patch } : rule)));
  };

  const addUnlockRule = () => setUnlockRules([...normalizedUnlockRules, createEmptyUnlockRule()]);

  const removeUnlockRule = (index) => {
    const nextRules = normalizedUnlockRules.filter((_, ruleIndex) => ruleIndex !== index);
    setUnlockRules(nextRules);
  };

  return (
    <Stack
      direction="row"
      spacing={2}
      sx={{
        mb: 2,
        p: 2,
        borderRadius: 2.5,
        bgcolor: tokens.toolbarSurface,
        border: '1px solid',
        borderColor: tokens.softBorder,
        alignItems: 'flex-start'
      }}
      flexWrap="wrap"
      useFlexGap
    >
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.group')}</InputLabel>
        <Select value={groupFilter} label={t('nodes.filters.group')} onChange={(e) => setGroupFilter(e.target.value)} variant={'outlined'}>
          <MenuItem value="">{allLabel}</MenuItem>
          <MenuItem value={UNGROUPED_GROUP_VALUE}>{t('nodes.filters.ungrouped')}</MenuItem>
          {groupOptions.map((group) => (
            <MenuItem key={group} value={group}>
              {group}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <TextField
        size="small"
        placeholder={t('nodes.filters.searchPlaceholder')}
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        sx={{ minWidth: 200, flex: { xs: '1 1 100%', md: '1 1 220px' }, ...fieldControlSx }}
      />
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.source')}</InputLabel>
        <Select
          value={sourceFilter}
          label={t('nodes.filters.source')}
          onChange={(e) => setSourceFilter(e.target.value)}
          variant={'outlined'}
        >
          <MenuItem value="">{allLabel}</MenuItem>
          {sourceOptions.map((source) => (
            <MenuItem key={source} value={source}>
              {source === 'manual' ? t('nodes.filters.manualSource') : source}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.protocol')}</InputLabel>
        <Select
          value={protocolFilter}
          label={t('nodes.filters.protocol')}
          onChange={(e) => setProtocolFilter(e.target.value)}
          variant={'outlined'}
        >
          <MenuItem value="">{allLabel}</MenuItem>
          {protocolOptions.map((protocol) => (
            <MenuItem key={protocol} value={protocol}>
              {protocol.toUpperCase()}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 100, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.delayStatus')}</InputLabel>
        <Select value={delayStatusFilter} label={t('nodes.filters.delayStatus')} onChange={(e) => setDelayStatusFilter(e.target.value)}>
          {STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 100, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.speedStatus')}</InputLabel>
        <Select value={speedStatusFilter} label={t('nodes.filters.speedStatus')} onChange={(e) => setSpeedStatusFilter(e.target.value)}>
          {STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <TextField
        size="small"
        placeholder={t('nodes.filters.maxDelay')}
        type="number"
        value={maxDelay}
        onChange={(e) => setMaxDelay(e.target.value)}
        sx={{ width: 150, ...fieldControlSx }}
        InputProps={{ endAdornment: <InputAdornment position="end">ms</InputAdornment> }}
      />
      <TextField
        size="small"
        placeholder={t('nodes.filters.minSpeed')}
        type="number"
        value={minSpeed}
        onChange={(e) => setMinSpeed(e.target.value)}
        sx={{ width: 150, ...fieldControlSx }}
        InputProps={{ endAdornment: <InputAdornment position="end">MB/s</InputAdornment> }}
      />
      <TextField
        size="small"
        placeholder={t('nodes.filters.maxFraudScore')}
        type="number"
        value={maxFraudScore}
        onChange={(e) => setMaxFraudScore(e.target.value)}
        sx={{ width: 160, ...fieldControlSx }}
      />
      <FormControl size="small" sx={{ minWidth: 140, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.qualityStatus')}</InputLabel>
        <Select value={qualityStatus} label={t('nodes.filters.qualityStatus')} onChange={(e) => setQualityStatus(e.target.value)}>
          {QUALITY_STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.residentialType')}</InputLabel>
        <Select value={residentialType} label={t('nodes.filters.residentialType')} onChange={(e) => setResidentialType(e.target.value)}>
          <MenuItem value="">{allLabel}</MenuItem>
          <MenuItem value="residential">{t('nodeConditions.residential.residential')}</MenuItem>
          <MenuItem value="datacenter">{t('nodeConditions.residential.datacenter')}</MenuItem>
          <MenuItem value="untested">{t('nodeConditions.residential.untested')}</MenuItem>
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 120, ...fieldControlSx }}>
        <InputLabel>{t('nodes.filters.ipType')}</InputLabel>
        <Select value={ipType} label={t('nodes.filters.ipType')} onChange={(e) => setIpType(e.target.value)}>
          <MenuItem value="">{allLabel}</MenuItem>
          <MenuItem value="native">{t('nodeConditions.ipType.native')}</MenuItem>
          <MenuItem value="broadcast">{t('nodeConditions.ipType.broadcast')}</MenuItem>
          <MenuItem value="untested">{t('nodeConditions.ipType.untested')}</MenuItem>
        </Select>
      </FormControl>
      {countryOptions.length > 0 && (
        <Autocomplete
          multiple
          size="small"
          options={countryOptions}
          value={countryFilter}
          onChange={(_, newValue) => setCountryFilter(newValue)}
          sx={{ minWidth: 150, ...fieldControlSx }}
          getOptionLabel={(option) => formatCountry(option)}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                {formatCountry(option)}
              </li>
            );
          }}
          renderTags={(value, getTagProps) =>
            value.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return (
                <Chip
                  key={key}
                  label={formatCountry(option)}
                  size="small"
                  sx={{
                    ...getNodeColorChipSx(theme, tokens, theme.palette.primary.main, { deletable: true }),
                    bgcolor: tokens.isDark ? withAlpha(theme.palette.primary.main, 0.12) : undefined,
                    borderColor: tokens.isDark ? withAlpha(theme.palette.primary.main, 0.3) : undefined
                  }}
                  {...tagProps}
                />
              );
            })
          }
          renderInput={(params) => (
            <TextField {...params} label={t('nodes.filters.countryCode')} placeholder={t('nodes.filters.countryPlaceholder')} />
          )}
        />
      )}
      {tagOptions && tagOptions.length > 0 && (
        <Autocomplete
          multiple
          size="small"
          options={tagOptions}
          value={tagFilter}
          onChange={(_, newValue) => setTagFilter(newValue)}
          sx={{ minWidth: 150, ...fieldControlSx }}
          getOptionLabel={(option) => option.name || option}
          isOptionEqualToValue={(option, value) => option.name === (value.name || value)}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                <Box
                  sx={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    backgroundColor: option.color || tokens.palette.primary.main,
                    mr: 1,
                    flexShrink: 0
                  }}
                />
                {option.name}
              </li>
            );
          }}
          renderTags={(value, getTagProps) =>
            value.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return (
                <Chip
                  key={key}
                  label={option.name || option}
                  size="small"
                  sx={{
                    ...getNodeColorChipSx(theme, tokens, option.color || theme.palette.primary.main, { deletable: true }),
                    bgcolor: tokens.isDark ? withAlpha(option.color || theme.palette.primary.main, 0.12) : undefined,
                    borderColor: tokens.isDark ? withAlpha(option.color || theme.palette.primary.main, 0.3) : undefined
                  }}
                  {...tagProps}
                />
              );
            })
          }
          renderInput={(params) => (
            <TextField {...params} label={t('nodes.filters.tags')} placeholder={t('nodes.filters.tagsPlaceholder')} sx={fieldControlSx} />
          )}
        />
      )}
      <Box sx={{ width: '100%', minWidth: 320 }}>
        <Accordion
          expanded={unlockExpanded}
          onChange={(_, expanded) => setUnlockExpanded(expanded)}
          sx={{
            boxShadow: 'none',
            border: '1px solid',
            borderColor: tokens.softBorder,
            borderRadius: 2,
            bgcolor: tokens.nestedPanelSurface,
            '&:before': { display: 'none' }
          }}
        >
          <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={{ minHeight: 56 }}>
            <Stack spacing={0.25} sx={{ width: '100%' }}>
              <Typography variant="subtitle2">{t('nodes.filters.unlock.title')}</Typography>
              <Typography variant="caption" color="text.secondary">
                {normalizedUnlockRules.length > 0
                  ? t('nodes.filters.unlock.summary', {
                      count: normalizedUnlockRules.length,
                      mode: unlockRuleMode === 'and' ? t('nodes.filters.unlock.modeAll') : t('nodes.filters.unlock.modeAny')
                    })
                  : t('nodes.filters.unlock.emptySummary')}
              </Typography>
            </Stack>
          </AccordionSummary>
          <AccordionDetails>
            <Stack spacing={1.5}>
              <Alert severity="info" variant="outlined">
                {t('nodes.filters.unlock.description')}
              </Alert>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} alignItems={{ md: 'center' }}>
                <FormControl size="small" sx={{ minWidth: 200, ...fieldControlSx }}>
                  <InputLabel>{t('nodes.filters.unlock.relation')}</InputLabel>
                  <Select
                    value={unlockRuleMode || 'or'}
                    label={t('nodes.filters.unlock.relation')}
                    onChange={(e) => setUnlockRuleMode(e.target.value)}
                  >
                    {getUnlockRuleModeOptions().map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.labelKey ? t(option.labelKey, option.label) : option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <Typography variant="caption" color="text.secondary">
                  {unlockRuleMode === 'and' ? t('nodes.filters.unlock.andHelper') : t('nodes.filters.unlock.orHelper')}
                </Typography>
              </Stack>
              {normalizedUnlockRules.length > 0 ? (
                normalizedUnlockRules.map((rule, index) => (
                  <Stack
                    key={`node-unlock-rule-${index}`}
                    direction={{ xs: 'column', md: 'row' }}
                    spacing={1.5}
                    alignItems={{ md: 'flex-start' }}
                  >
                    <Autocomplete
                      size="small"
                      options={unlockProviderOptions}
                      value={unlockProviderOptions.find((item) => item.value === rule.provider) || null}
                      onChange={(_, newValue) => updateUnlockRule(index, { provider: newValue?.value || '' })}
                      getOptionLabel={(option) =>
                        option?.labelKey
                          ? t(option.labelKey, option.label || option.value)
                          : option?.label || formatUnlockProviderLabel(option?.value || '')
                      }
                      sx={{ minWidth: 220, flex: 1, ...fieldControlSx }}
                      renderOption={(props, option) => (
                        <li {...props} key={option.value}>
                          <Box>
                            <Typography variant="body2">
                              {option.labelKey ? t(option.labelKey, option.label || option.value) : option.label}
                            </Typography>
                            <Typography variant="caption" color="text.secondary">
                              {option.description || option.value}
                            </Typography>
                          </Box>
                        </li>
                      )}
                      renderInput={(params) => <TextField {...params} label="Provider" sx={fieldControlSx} />}
                    />
                    <FormControl size="small" sx={{ minWidth: 180, ...fieldControlSx }}>
                      <InputLabel>{t('nodes.filters.unlock.status')}</InputLabel>
                      <Select
                        value={rule.status || ''}
                        label={t('nodes.filters.unlock.status')}
                        onChange={(e) => updateUnlockRule(index, { status: e.target.value })}
                      >
                        {getUnlockStatusOptions(true).map((opt) => (
                          <MenuItem key={opt.value || 'all'} value={opt.value}>
                            {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <TextField
                      size="small"
                      label={t('nodes.filters.unlock.keyword')}
                      value={rule.keyword || ''}
                      onChange={(e) => updateUnlockRule(index, { keyword: e.target.value })}
                      sx={{ minWidth: 220, flex: 1, ...fieldControlSx }}
                    />
                    <IconButton color="error" onClick={() => removeUnlockRule(index)} sx={{ alignSelf: { xs: 'flex-end', md: 'center' } }}>
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
                  </Stack>
                ))
              ) : (
                <Alert severity="info" variant="outlined">
                  {t('nodes.filters.unlock.inactive')}
                </Alert>
              )}
              <Box>
                <Button size="small" startIcon={<AddIcon />} variant="outlined" onClick={addUnlockRule}>
                  {t('nodes.filters.unlock.addRule')}
                </Button>
              </Box>
            </Stack>
          </AccordionDetails>
        </Accordion>
      </Box>
      <Button onClick={onReset} sx={getNodeActionButtonSx(theme, tokens, tokens.palette.text.secondary)}>
        {t('common.reset')}
      </Button>
    </Stack>
  );
}

NodeFilters.propTypes = {
  searchQuery: PropTypes.string.isRequired,
  setSearchQuery: PropTypes.func.isRequired,
  groupFilter: PropTypes.string.isRequired,
  setGroupFilter: PropTypes.func.isRequired,
  sourceFilter: PropTypes.string.isRequired,
  setSourceFilter: PropTypes.func.isRequired,
  maxDelay: PropTypes.string.isRequired,
  setMaxDelay: PropTypes.func.isRequired,
  minSpeed: PropTypes.string.isRequired,
  setMinSpeed: PropTypes.func.isRequired,
  maxFraudScore: PropTypes.string.isRequired,
  setMaxFraudScore: PropTypes.func.isRequired,
  speedStatusFilter: PropTypes.string.isRequired,
  setSpeedStatusFilter: PropTypes.func.isRequired,
  delayStatusFilter: PropTypes.string.isRequired,
  setDelayStatusFilter: PropTypes.func.isRequired,
  residentialType: PropTypes.string.isRequired,
  setResidentialType: PropTypes.func.isRequired,
  ipType: PropTypes.string.isRequired,
  setIpType: PropTypes.func.isRequired,
  qualityStatus: PropTypes.string.isRequired,
  setQualityStatus: PropTypes.func.isRequired,
  unlockRules: PropTypes.array.isRequired,
  setUnlockRules: PropTypes.func.isRequired,
  unlockRuleMode: PropTypes.string.isRequired,
  setUnlockRuleMode: PropTypes.func.isRequired,
  countryFilter: PropTypes.array.isRequired,
  setCountryFilter: PropTypes.func.isRequired,
  tagFilter: PropTypes.array.isRequired,
  setTagFilter: PropTypes.func.isRequired,
  protocolFilter: PropTypes.string.isRequired,
  setProtocolFilter: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  sourceOptions: PropTypes.array.isRequired,
  countryOptions: PropTypes.array.isRequired,
  tagOptions: PropTypes.array.isRequired,
  protocolOptions: PropTypes.array.isRequired,
  onReset: PropTypes.func.isRequired
};
