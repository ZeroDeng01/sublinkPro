import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import FormHelperText from '@mui/material/FormHelperText';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { getNodeCountries, getNodeGroups, getNodeProtocols, getNodeSources } from 'api/nodes';
import { formatCountry } from 'utils/countryDisplay';

const UNGROUPED_VALUE = '__ungrouped__';

function mergeOptions(options, selected) {
  return Array.from(new Set([...(selected || []), ...(options || [])].filter(Boolean)));
}

function PoolAutocomplete({ label, placeholder, options, value, onChange, disabled, getOptionLabel }) {
  return (
    <Autocomplete
      multiple
      fullWidth
      limitTags={3}
      options={mergeOptions(options, value)}
      value={value || []}
      disabled={disabled}
      onChange={(_, nextValue) => onChange(nextValue)}
      getOptionLabel={getOptionLabel || ((option) => option)}
      renderOption={(props, option) => {
        const { key, ...optionProps } = props;
        return (
          <li key={key} {...optionProps}>
            {getOptionLabel ? getOptionLabel(option) : option}
          </li>
        );
      }}
      renderInput={(params) => <TextField {...params} label={label} placeholder={value?.length ? '' : placeholder} />}
    />
  );
}

export default function CandidateNodePool({ value, onChange, disabled = false }) {
  const { t } = useTranslation();
  const [options, setOptions] = useState({ groups: [], sources: [], protocols: [], countries: [] });
  const [loading, setLoading] = useState(true);

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
    return () => {
      active = false;
    };
  }, []);

  const pool = useMemo(
    () => ({
      candidateGroups: Array.isArray(value?.candidateGroups) ? value.candidateGroups : [],
      candidateSources: Array.isArray(value?.candidateSources) ? value.candidateSources : [],
      candidateProtocols: Array.isArray(value?.candidateProtocols) ? value.candidateProtocols : [],
      candidateCountries: Array.isArray(value?.candidateCountries) ? value.candidateCountries : []
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
      </Stack>
    </Box>
  );
}
