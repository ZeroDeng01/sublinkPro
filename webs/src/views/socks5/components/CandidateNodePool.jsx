import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import FormHelperText from '@mui/material/FormHelperText';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { getNodeCountries, getNodeGroups, getNodeProtocols, getNodeSources } from 'api/nodes';
import { getTags } from 'api/tags';
import { withAlpha } from 'utils/colorUtils';
import { formatCountry } from 'utils/countryDisplay';

const UNGROUPED_VALUE = '__ungrouped__';
const TAG_COLOR_PATTERN = /^#(?:[0-9a-f]{3}|[0-9a-f]{6})$/i;

function mergeOptions(options, selected) {
  return Array.from(new Set([...(options || []), ...(selected || [])].filter(Boolean)));
}

function PoolTagChip({ label, color, ...props }) {
  return (
    <Chip
      {...props}
      label={label}
      icon={
        color ? (
          <Box
            component="span"
            aria-hidden="true"
            sx={{
              width: 9,
              height: 9,
              flexShrink: 0,
              borderRadius: '50%',
              bgcolor: color,
              boxShadow: '0 0 0 1px currentColor',
              color: 'divider'
            }}
          />
        ) : undefined
      }
      sx={
        color
          ? {
              color: 'text.primary',
              bgcolor: withAlpha(color, 0.16),
              border: '1px solid',
              borderColor: withAlpha(color, 0.42)
            }
          : undefined
      }
    />
  );
}

function PoolAutocomplete({ label, placeholder, options, value, onChange, disabled, getOptionLabel, helperText, tagColors }) {
  return (
    <Autocomplete
      multiple
      fullWidth
      disableCloseOnSelect
      limitTags={3}
      options={mergeOptions(options, value)}
      value={value || []}
      disabled={disabled}
      onChange={(_, nextValue) => onChange(nextValue)}
      getOptionLabel={getOptionLabel || ((option) => option)}
      renderOption={(props, option) => {
        const { key, ...optionProps } = props;
        const color = tagColors?.get(option);
        return (
          <li key={key} {...optionProps}>
            {color ? <PoolTagChip label={option} color={color} size="small" /> : (getOptionLabel?.(option) ?? option)}
          </li>
        );
      }}
      renderValue={
        tagColors
          ? (selected, getItemProps) =>
              selected.map((tag, index) => {
                const { key, ...itemProps } = getItemProps({ index });
                return <PoolTagChip key={key} label={tag} color={tagColors.get(tag)} {...itemProps} />;
              })
          : undefined
      }
      renderInput={(params) => (
        <TextField {...params} label={label} placeholder={value?.length ? '' : placeholder} helperText={helperText} />
      )}
    />
  );
}

export default function CandidateNodePool({ value, onChange, disabled = false }) {
  const { t } = useTranslation();
  const [options, setOptions] = useState({ groups: [], sources: [], protocols: [], countries: [] });
  const [loading, setLoading] = useState(true);
  const [tagOptions, setTagOptions] = useState([]);
  const [loadingTags, setLoadingTags] = useState(true);
  const [tagLoadFailed, setTagLoadFailed] = useState(false);

  useEffect(() => {
    let active = true;
    setLoading(true);
    Promise.all([getNodeGroups(), getNodeSources(), getNodeProtocols(), getNodeCountries()])
      .then(([groups, sources, protocols, countries]) => {
        if (!active) return;
        setOptions({
          groups: [UNGROUPED_VALUE, ...(groups.data || []).filter(Boolean).sort()],
          sources: Array.from(new Set(['manual', ...(sources.data || []).filter(Boolean).sort()])),
          protocols: (protocols.data || []).filter(Boolean).sort(),
          countries: (countries.data || []).filter(Boolean).sort()
        });
      })
      .catch(() => {
        if (active) setOptions({ groups: [UNGROUPED_VALUE], sources: ['manual'], protocols: [], countries: [] });
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    getTags()
      .then(({ data }) => {
        if (active) setTagOptions(data || []);
      })
      .catch(() => {
        if (active) setTagLoadFailed(true);
      })
      .finally(() => {
        if (active) setLoadingTags(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const tagColors = useMemo(
    () => new Map(tagOptions.filter((tag) => tag.name).map(({ name, color }) => [name, TAG_COLOR_PATTERN.test(color) ? color : undefined])),
    [tagOptions]
  );

  const pool = useMemo(
    () => ({
      candidateGroups: Array.isArray(value?.candidateGroups) ? value.candidateGroups : [],
      candidateSources: Array.isArray(value?.candidateSources) ? value.candidateSources : [],
      candidateProtocols: Array.isArray(value?.candidateProtocols) ? value.candidateProtocols : [],
      candidateCountries: Array.isArray(value?.candidateCountries) ? value.candidateCountries : [],
      candidateTags: Array.isArray(value?.candidateTags) ? value.candidateTags : []
    }),
    [value]
  );

  const update = (field, nextValue) => onChange({ ...pool, [field]: nextValue });
  const groupLabel = (group) => (group === UNGROUPED_VALUE ? t('nodes.filters.ungrouped') : group);
  const sourceLabel = (source) => (source === 'manual' ? t('nodes.filters.manualSource') : source);
  const protocolLabel = (protocol) => protocol.toUpperCase();

  return (
    <Box sx={{ p: 2, border: '1px solid', borderColor: 'divider', borderRadius: 2 }}>
      <Stack spacing={1.5}>
        <Box>
          <Typography variant="subtitle1">{t('settings.socks5.form.poolTitle')}</Typography>
          <FormHelperText>{t('settings.socks5.form.poolHelper')}</FormHelperText>
        </Box>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
          <PoolAutocomplete
            label={t('settings.socks5.form.groups')}
            placeholder={t('settings.socks5.form.groupsPlaceholder')}
            options={options.groups}
            value={pool.candidateGroups}
            onChange={(nextValue) => update('candidateGroups', nextValue)}
            disabled={disabled || loading}
            getOptionLabel={groupLabel}
          />
          <PoolAutocomplete
            label={t('settings.socks5.form.sources')}
            placeholder={t('settings.socks5.form.sourcesPlaceholder')}
            options={options.sources}
            value={pool.candidateSources}
            onChange={(nextValue) => update('candidateSources', nextValue)}
            disabled={disabled || loading}
            getOptionLabel={sourceLabel}
          />
        </Stack>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
          <PoolAutocomplete
            label={t('settings.socks5.form.protocols')}
            placeholder={t('settings.socks5.form.protocolsPlaceholder')}
            options={options.protocols}
            value={pool.candidateProtocols}
            onChange={(nextValue) => update('candidateProtocols', nextValue)}
            disabled={disabled || loading}
            getOptionLabel={protocolLabel}
          />
          <PoolAutocomplete
            label={t('settings.socks5.form.countries')}
            placeholder={t('settings.socks5.form.countriesPlaceholder')}
            options={options.countries}
            value={pool.candidateCountries}
            onChange={(nextValue) => update('candidateCountries', nextValue)}
            disabled={disabled || loading}
            getOptionLabel={formatCountry}
          />
        </Stack>
        <PoolAutocomplete
          label={t('settings.socks5.form.tags')}
          placeholder={t('settings.socks5.form.tagsPlaceholder')}
          options={[...tagColors.keys()].sort()}
          value={pool.candidateTags}
          onChange={(nextValue) => update('candidateTags', nextValue)}
          disabled={disabled || loadingTags}
          helperText={tagLoadFailed ? t('settings.socks5.form.tagsLoadFailed') : undefined}
          tagColors={tagColors}
        />
      </Stack>
    </Box>
  );
}
