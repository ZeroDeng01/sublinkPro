import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Autocomplete,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  MenuItem,
  Stack,
  Switch,
  TextField,
  Typography,
  useMediaQuery,
  useTheme
} from '@mui/material';
import CronExpressionGenerator from 'components/CronExpressionGenerator';
import { saveNodeCheckJob } from 'api/nodeCheck';

// 任务配置只保存编排参数，不复制方案的检测配置。已删除的分组/标签保留为可编辑值。
export default function CheckJobDialog({ job, profiles, options, timezone, onClose, onSaved }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const mobile = useMediaQuery(theme.breakpoints.down('sm'));
  const [form, setForm] = useState(() => ({
    name: '',
    profileId: profiles[0]?.id || '',
    enabled: true,
    trigger: 'scheduled',
    cronExpr: '30 5 * * *',
    scope: 'all',
    updateBeforeCheck: false,
    ...job,
    airportIds: Array.isArray(job.airportIds) ? job.airportIds : [],
    groups: Array.isArray(job.groups) ? job.groups : [],
    tags: Array.isArray(job.tags) ? job.tags : []
  }));
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const update = (key, value) => setForm((prev) => ({ ...prev, [key]: value }));
  // 这里只检查五字段结构；月份/星期名称和单值步进等合法 Cron 由服务端校验，不能被预览器的数字语法限制。
  const invalid =
    !form.name.trim() ||
    !form.profileId ||
    (form.trigger === 'scheduled' && form.cronExpr.trim().split(/\s+/).length !== 5) ||
    (form.scope === 'filtered' && !form.airportIds.length && !form.groups.length && !form.tags.length) ||
    (form.updateBeforeCheck && !form.airportIds.length);
  const save = async () => {
    setBusy(true);
    setError('');
    try {
      await saveNodeCheckJob({
        ...form,
        profileId: Number(form.profileId),
        ...(form.scope === 'all' ? { airportIds: [], groups: [], tags: [], updateBeforeCheck: false } : {})
      });
      onSaved();
    } catch (err) {
      setError(err.message || t('detectionCenter.saveFailed'));
    } finally {
      setBusy(false);
    }
  };
  const stringSelect = (key, values) => (
    <Autocomplete
      multiple
      freeSolo
      options={values}
      value={form[key] || []}
      onChange={(_, value) => update(key, value)}
      renderInput={(params) => <TextField {...params} label={t(`detectionCenter.${key}`)} />}
    />
  );
  return (
    <Dialog open fullWidth maxWidth="md" fullScreen={mobile} onClose={busy ? undefined : onClose}>
      <DialogTitle>{t(job.id ? 'detectionCenter.editJob' : 'detectionCenter.newJob')}</DialogTitle>
      <DialogContent sx={{ px: 2.5 }}>
        <Stack spacing={2} sx={{ mt: 2 }}>
          {error && <Alert severity="error">{error}</Alert>}
          {job.legacy && <Alert severity="info">{t('detectionCenter.legacyHelp')}</Alert>}
          <TextField
            label={t('detectionCenter.jobName')}
            disabled={job.legacy}
            value={form.name}
            onChange={(e) => update('name', e.target.value)}
            required
            inputProps={{ maxLength: 191 }}
          />
          <TextField
            select
            label={t('detectionCenter.plan')}
            value={form.profileId}
            disabled={job.legacy}
            onChange={(e) => update('profileId', e.target.value)}
          >
            {profiles.map((p) => (
              <MenuItem key={p.id} value={p.id}>
                {p.name}
              </MenuItem>
            ))}
          </TextField>
          <FormControlLabel
            control={<Switch checked={form.enabled} onChange={(_, value) => update('enabled', value)} />}
            label={t('detectionCenter.autoEnabled')}
          />
          <TextField
            select
            label={t('detectionCenter.trigger')}
            value={form.trigger}
            disabled={job.legacy}
            onChange={(e) => setForm((prev) => ({ ...prev, trigger: e.target.value, updateBeforeCheck: false }))}
          >
            {['scheduled', 'airport_update'].map((key) => (
              <MenuItem key={key} value={key}>
                {t(`detectionCenter.triggers.${key}`)}
              </MenuItem>
            ))}
          </TextField>
          {form.trigger === 'scheduled' && (
            <CronExpressionGenerator
              value={form.cronExpr}
              onChange={(value) => update('cronExpr', value)}
              label={t('detectionCenter.cron')}
              error={form.cronExpr.trim().split(/\s+/).length !== 5}
              helperText={t('detectionCenter.cronHelp')}
            />
          )}
          <Typography variant="caption" color="text.secondary">
            {t('detectionCenter.timezone', { timezone })}
          </Typography>
          <TextField
            select
            label={t('detectionCenter.scope')}
            value={form.scope}
            onChange={(e) =>
              setForm((prev) => ({
                ...prev,
                scope: e.target.value,
                ...(e.target.value === 'all' ? { airportIds: [], groups: [], tags: [], updateBeforeCheck: false } : {})
              }))
            }
          >
            {['all', 'filtered'].map((key) => (
              <MenuItem key={key} value={key}>
                {t(`detectionCenter.scopes.${key}`)}
              </MenuItem>
            ))}
          </TextField>
          {form.scope !== 'all' && (
            <>
              {!job.legacy && (
                <Autocomplete
                  multiple
                  options={options.airports}
                  getOptionLabel={(a) => a.Name || a.name || String(a.ID || a.id)}
                  isOptionEqualToValue={(a, b) => (a.ID || a.id) === (b.ID || b.id)}
                  value={form.airportIds.map(
                    (id) =>
                      options.airports.find((a) => (a.ID || a.id) === id) || { ID: id, Name: t('detectionCenter.missingAirport', { id }) }
                  )}
                  onChange={(_, value) =>
                    update(
                      'airportIds',
                      value.map((a) => a.ID || a.id)
                    )
                  }
                  renderInput={(params) => (
                    <TextField {...params} label={t('detectionCenter.airports')} helperText={t('detectionCenter.scopeAnd')} />
                  )}
                />
              )}
              {stringSelect('groups', options.groups)}
              {stringSelect(
                'tags',
                options.tags.map((tag) => tag.name || tag.Name || tag)
              )}
            </>
          )}
          {!job.legacy && form.trigger === 'scheduled' && form.scope !== 'all' && (
            <FormControlLabel
              control={<Switch checked={form.updateBeforeCheck} onChange={(_, value) => update('updateBeforeCheck', value)} />}
              label={t('detectionCenter.updateFirst')}
            />
          )}
          <Alert severity="info">{t('detectionCenter.dynamicMembership')}</Alert>
          <Typography variant="caption" color="text.secondary">
            {t('detectionCenter.openClashHelp')}
          </Typography>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={busy}>
          {t('common.cancel')}
        </Button>
        <Button variant="contained" onClick={save} disabled={busy || invalid}>
          {t('common.save')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
