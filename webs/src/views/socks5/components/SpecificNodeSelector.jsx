import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { getNodeGroups, getNodeSelector, getNodeSelectorByIds } from 'api/nodes';

const PAGE_SIZE = 100;

function nodeLabel(node) {
  return node?.EffectiveName || node?.Name || node?.LinkName || (node?.ID ? `#${node.ID}` : '');
}

function mergeNodes(current, incoming) {
  const merged = new Map();
  [...(current || []), ...(incoming || [])].forEach((node) => {
    if (Number(node?.ID) > 0) merged.set(Number(node.ID), node);
  });
  return Array.from(merged.values());
}

export default function SpecificNodeSelector({ value, onChange, disabled = false, error = false }) {
  const { t } = useTranslation();
  const [nodes, setNodes] = useState([]);
  const [groups, setGroups] = useState([]);
  const [group, setGroup] = useState('');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loadedCount, setLoadedCount] = useState(0);
  const requestRef = useRef(0);
  const loadingRef = useRef(false);

  const selectedNode = useMemo(() => nodes.find((node) => Number(node.ID) === Number(value)) || null, [nodes, value]);
  const hasMore = loadedCount < total;

  useEffect(() => {
    let active = true;
    getNodeGroups()
      .then((response) => {
        if (active) setGroups((response.data || []).filter(Boolean).sort());
      })
      .catch(() => {
        if (active) setGroups([]);
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    const nodeID = Number(value);
    if (!nodeID || nodes.some((node) => Number(node.ID) === nodeID)) return;

    let active = true;
    getNodeSelectorByIds({ 'ids[]': [nodeID] })
      .then((response) => {
        if (active) setNodes((current) => mergeNodes(current, response.data || []));
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, [nodes, value]);

  const fetchNodes = useCallback(
    async (nextPage = 1, append = false) => {
      if (append && loadingRef.current) return;

      const requestID = requestRef.current + 1;
      requestRef.current = requestID;
      loadingRef.current = true;
      setLoading(true);

      try {
        const response = await getNodeSelector({
          page: nextPage,
          pageSize: PAGE_SIZE,
          search: search.trim(),
          group
        });
        if (requestRef.current !== requestID) return;

        const items = response.data?.items || [];
        const responseTotal = Number(response.data?.total) || items.length;
        setNodes((current) => {
          const selected = Number(value) > 0 ? current.filter((node) => Number(node.ID) === Number(value)) : [];
          return append ? mergeNodes(current, items) : mergeNodes(selected, items);
        });
        setPage(nextPage);
        setTotal(responseTotal);
        setLoadedCount((current) => (append ? Math.min(responseTotal, current + items.length) : items.length));
      } finally {
        if (requestRef.current === requestID) {
          loadingRef.current = false;
          setLoading(false);
        }
      }
    },
    [group, search, value]
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchNodes(1, false);
    }, 300);
    return () => window.clearTimeout(timer);
  }, [fetchNodes]);

  const handleListScroll = (event) => {
    const listbox = event.currentTarget;
    const reachedBottom = listbox.scrollTop + listbox.clientHeight >= listbox.scrollHeight - 12;
    if (reachedBottom && hasMore && !loadingRef.current) void fetchNodes(page + 1, true);
  };

  return (
    <Stack spacing={1.25}>
      <FormControl fullWidth>
        <InputLabel>{t('settings.socks5.form.nodeGroup')}</InputLabel>
        <Select
          label={t('settings.socks5.form.nodeGroup')}
          value={group}
          onChange={(event) => setGroup(event.target.value)}
          disabled={disabled}
        >
          <MenuItem value="">{t('settings.socks5.form.nodeGroupAll')}</MenuItem>
          {groups.map((item) => (
            <MenuItem value={item} key={item}>
              {item}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <Autocomplete
        fullWidth
        options={nodes}
        value={selectedNode}
        loading={loading}
        disabled={disabled}
        filterOptions={(options) => options}
        getOptionLabel={nodeLabel}
        isOptionEqualToValue={(option, selected) => Number(option.ID) === Number(selected.ID)}
        onChange={(_, node) => onChange(Number(node?.ID) || 0)}
        onInputChange={(_, input, reason) => {
          if (reason === 'input' || reason === 'clear') setSearch(input);
        }}
        noOptionsText={t('settings.socks5.form.nodeNoOptions')}
        loadingText={t('settings.socks5.form.nodeLoading')}
        slotProps={{ listbox: { onScroll: handleListScroll } }}
        renderOption={(props, option) => {
          const { key, ...optionProps } = props;
          const metadata = [option.Group, option.Source, option.LinkCountry].filter(Boolean).join(' · ');
          return (
            <Box component="li" key={key} {...optionProps}>
              <Box sx={{ minWidth: 0 }}>
                <Typography variant="body2" noWrap>
                  {nodeLabel(option)}
                </Typography>
                {metadata && (
                  <Typography variant="caption" color="text.secondary" noWrap>
                    {metadata}
                  </Typography>
                )}
              </Box>
            </Box>
          );
        }}
        renderInput={(params) => (
          <TextField
            {...params}
            label={t('settings.socks5.form.node')}
            placeholder={t('settings.socks5.form.nodeSearchPlaceholder')}
            error={error}
          />
        )}
      />
      <FormHelperText error={error}>
        {error ? t('settings.socks5.form.nodeRequired') : t('settings.socks5.form.nodeSearchHelper', { shown: loadedCount, total })}
      </FormHelperText>
    </Stack>
  );
}
