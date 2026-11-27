import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Collapse from '@mui/material/Collapse';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import { formatUnlockProviderLabel, getUnlockProviderOptions, getUnlockStatusOptions } from 'views/nodes/utils';
import { createServiceMark, serviceMarksError } from './serviceMarks';

const keys = 'subscriptions.rename.serviceMarks';

export default function ServiceMarksEditor({ marks, blockNumber, onChange, onClose }) {
  const { t } = useTranslation();
  const [showOptions, setShowOptions] = useState(false);
  const providers = getUnlockProviderOptions();
  const statuses = getUnlockStatusOptions(false);
  const statusLabels = new Map(statuses.map(({ value, label }) => [value, label]));
  const selected = new Set(marks.services.map((mark) => mark.provider));
  const available = providers.filter(({ value }) => !selected.has(value));
  const error = marks.services.length > 0 ? serviceMarksError(marks, [...statusLabels.keys()]) : null;

  const updateService = (index, updates) => {
    onChange({ ...marks, services: marks.services.map((mark, position) => (position === index ? { ...mark, ...updates } : mark)) });
  };

  const moveService = (from, to) => {
    const services = [...marks.services];
    const [moved] = services.splice(from, 1);
    services.splice(to, 0, moved);
    onChange({ ...marks, services });
  };

  return (
    <Stack spacing={1.5} sx={{ mb: 2, p: 2, bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 2 }}>
      <Stack direction="row" justifyContent="space-between" alignItems="center" gap={1}>
        <Typography variant="subtitle2">{t(`${keys}.title`)}</Typography>
        <Button size="small" onClick={onClose}>
          {t(`${keys}.done`)}
        </Button>
      </Stack>
      {marks.services.map((mark, index) => {
        const provider = formatUnlockProviderLabel(mark.provider);
        const label = mark.label ?? '';
        return (
          <Stack
            key={mark.provider}
            spacing={1}
            sx={{ p: 1.5, bgcolor: 'background.paper', border: 1, borderColor: 'divider', borderRadius: 1.5 }}
          >
            <Stack direction="row" alignItems="center" justifyContent="space-between" gap={1}>
              <Typography variant="body2" fontWeight={600} sx={{ overflowWrap: 'anywhere' }}>
                {provider}
              </Typography>
              <Box sx={{ display: 'flex' }}>
                <IconButton
                  size="small"
                  aria-label={t(`${keys}.moveUp`, { service: provider })}
                  disabled={index === 0}
                  onClick={() => moveService(index, index - 1)}
                >
                  <ArrowUpwardIcon fontSize="small" />
                </IconButton>
                <IconButton
                  size="small"
                  aria-label={t(`${keys}.moveDown`, { service: provider })}
                  disabled={index === marks.services.length - 1}
                  onClick={() => moveService(index, index + 1)}
                >
                  <ArrowDownwardIcon fontSize="small" />
                </IconButton>
                <IconButton
                  size="small"
                  aria-label={t(`${keys}.remove`, { service: provider })}
                  onClick={() => onChange({ ...marks, services: marks.services.filter((_, position) => position !== index) })}
                >
                  <DeleteOutlineIcon fontSize="small" />
                </IconButton>
              </Box>
            </Stack>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
              <TextField
                size="small"
                fullWidth
                label={t(`${keys}.label`)}
                value={label}
                onChange={(event) => updateService(index, { label: event.target.value })}
              />
              <TextField
                select
                size="small"
                fullWidth
                label={t(`${keys}.statuses`)}
                value={mark.acceptedStatuses}
                onChange={(event) => updateService(index, { acceptedStatuses: event.target.value })}
                slotProps={{
                  select: { multiple: true, renderValue: (values) => values.map((value) => statusLabels.get(value) || value).join(' · ') }
                }}
              >
                {statuses.map(({ value, label: statusLabel }) => (
                  <MenuItem key={value} value={value}>
                    <Checkbox size="small" checked={mark.acceptedStatuses.includes(value)} />
                    {statusLabel}
                  </MenuItem>
                ))}
              </TextField>
            </Stack>
            {showOptions && (
              <FormControlLabel
                control={
                  <Switch size="small" checked={mark.region} onChange={(event) => updateService(index, { region: event.target.checked })} />
                }
                label={t(`${keys}.region`)}
              />
            )}
          </Stack>
        );
      })}
      <TextField
        select
        size="small"
        label={t(`${keys}.addService`)}
        disabled={available.length === 0}
        value=""
        onChange={(event) => {
          const service = available.find(({ value }) => value === event.target.value);
          if (service) onChange({ ...marks, services: [...marks.services, createServiceMark(service)] });
        }}
      >
        {available.map((service) => (
          <MenuItem key={service.value} value={service.value}>
            {formatUnlockProviderLabel(service.value)}
          </MenuItem>
        ))}
      </TextField>
      {error && (
        <Alert severity="error">{t(`${keys}.errors.${error.reason}`, { block: blockNumber, provider: error.provider || '' })}</Alert>
      )}
      <Box>
        <Button size="small" aria-expanded={showOptions} onClick={() => setShowOptions((open) => !open)}>
          {t(`${keys}.displayOptions`)}
        </Button>
      </Box>
      <Collapse in={showOptions}>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
          {['prefix', 'separator', 'suffix'].map((field) => (
            <TextField
              key={field}
              size="small"
              fullWidth
              label={t(`${keys}.${field}`)}
              value={marks[field]}
              onChange={(event) => onChange({ ...marks, [field]: event.target.value })}
            />
          ))}
        </Stack>
      </Collapse>
    </Stack>
  );
}
