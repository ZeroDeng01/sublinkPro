import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  IconButton,
  MenuItem,
  Paper,
  Snackbar,
  Stack,
  Switch,
  Tab,
  Tabs,
  TextField,
  Tooltip,
  Typography
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import RefreshIcon from '@mui/icons-material/Refresh';
import MainCard from 'ui-component/cards/MainCard';
import { useTaskProgress } from 'contexts/TaskProgressContext';
import {
  getNodeCheckProfiles,
  getNodeCheckMeta,
  deleteNodeCheckProfile,
  getNodeCheckJobs,
  saveNodeCheckJob,
  deleteNodeCheckJob,
  runNodeCheckJob,
  runNodeCheckSelection
} from 'api/nodeCheck';
import { getNodeGroups } from 'api/nodes';
import { getTags } from 'api/tags';
import { getAirports } from 'api/airports';
import { formatDateTime } from 'i18n/locales';
import { formatUnlockProvidersSummary, setUnlockMeta } from 'views/nodes/utils';
import NodeCheckProfileFormDialog from 'views/nodes/component/NodeCheckProfileFormDialog';
import CheckJobDialog from './components/CheckJobDialog';

const listData = (response) => (Array.isArray(response?.data) ? response.data : response?.data?.items || []);

// 方案定义检测能力，任务定义编排；旧策略 ID 与单节点执行入口保持兼容。
export default function NodeCheckList() {
  const { t, i18n } = useTranslation();
  const { registerOnComplete, unregisterOnComplete } = useTaskProgress();
  const [tab, setTab] = useState('plans');
  const [profiles, setProfiles] = useState([]);
  const [jobs, setJobs] = useState([]);
  const [options, setOptions] = useState({ groups: [], tags: [], airports: [] });
  const [timezone, setTimezone] = useState('');
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [busy, setBusy] = useState(false);
  const [planEditor, setPlanEditor] = useState(null);
  const [jobEditor, setJobEditor] = useState(null);
  const [runPlan, setRunPlan] = useState(null);
  const [selection, setSelection] = useState({ scope: 'all', groups: [], tags: [] });
  const [message, setMessage] = useState(null);
  const notify = useCallback((text, severity = 'success') => setMessage({ text, severity }), []);
  const load = useCallback(async () => {
    setLoadError('');
    try {
      const [p, j] = await Promise.all([getNodeCheckProfiles(), getNodeCheckJobs()]);
      setProfiles(listData(p));
      setJobs(j.data?.items || []);
      setTimezone(j.data?.timezone || '');
    } catch (err) {
      setLoadError(err.message || t('detectionCenter.loadFailed'));
    } finally {
      setLoading(false);
    }
  }, [t]);
  useEffect(() => {
    load();
  }, [load]);
  useEffect(() => {
    let active = true;
    Promise.all([getNodeGroups(), getTags(), getAirports(), getNodeCheckMeta()])
      .then(([groups, tags, airports, meta]) => {
        if (!active) return;
        setOptions({ groups: listData(groups), tags: listData(tags), airports: listData(airports) });
        setUnlockMeta(meta.data || {});
      })
      .catch((err) => {
        if (active) notify(err.message || t('detectionCenter.loadFailed'), 'error');
      });
    return () => {
      active = false;
    };
  }, [notify, t]);
  useEffect(() => {
    const completed = (task) => {
      if (['speed_test', 'sub_update'].includes(task.taskType)) {
        load();
      }
    };
    registerOnComplete(completed);
    return () => unregisterOnComplete(completed);
  }, [load, registerOnComplete, unregisterOnComplete]);
  useEffect(() => {
    if (!jobs.some((job) => job.running)) return;
    const timer = setTimeout(load, 5000);
    return () => clearTimeout(timer);
  }, [jobs, load]);
  const action = async (operation, text) => {
    setBusy(true);
    try {
      await operation();
      notify(text);
      await load();
    } catch (err) {
      notify(err.message || t('detectionCenter.operationFailed'), 'error');
    } finally {
      setBusy(false);
    }
  };
  const date = (value) =>
    !value || String(value).startsWith('0001-')
      ? '—'
      : formatDateTime(value, i18n.resolvedLanguage || i18n.language, { dateStyle: 'short', timeStyle: 'medium' }) || '—';
  const modules = (p) => [
    t('detectionCenter.module.latency'),
    ...(p.mode === 'mihomo' ? [t('detectionCenter.module.speed')] : []),
    ...(p.detectCountry ? [t('detectionCenter.module.exit')] : []),
    ...(p.detectQuality ? [t('detectionCenter.module.quality')] : []),
    ...(p.detectUnlock ? [formatUnlockProvidersSummary(p.unlockProviders, 3)] : [])
  ];
  const openRun = (p) => {
    setRunPlan(p);
    setSelection({ scope: 'all', groups: [], tags: [] });
  };
  const removePlan = (p) => {
    if (window.confirm(t('detectionCenter.confirmDelete', { name: p.name })))
      action(() => deleteNodeCheckProfile(p.id), t('detectionCenter.deleted'));
  };
  const removeJob = (job) => {
    if (window.confirm(t('detectionCenter.confirmDelete', { name: job.name })))
      action(() => deleteNodeCheckJob(job.id), t('detectionCenter.deleted'));
  };
  const planSaved = () => {
    setPlanEditor(null);
    load();
    notify(t('common.saveSuccess'));
  };
  return (
    <MainCard
      title={t('detectionCenter.title')}
      secondary={
        <Stack direction="row" spacing={1}>
          <Tooltip title={t('common.refresh')}>
            <span>
              <IconButton
                aria-label={t('common.refresh')}
                disabled={loading || busy}
                onClick={() => {
                  load();
                }}
              >
                <RefreshIcon />
              </IconButton>
            </span>
          </Tooltip>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => (tab === 'plans' ? setPlanEditor({}) : setJobEditor({}))}
            disabled={busy || (tab === 'jobs' && !profiles.length)}
          >
            {t(tab === 'plans' ? 'detectionCenter.newPlan' : 'detectionCenter.newJob')}
          </Button>
        </Stack>
      }
    >
      <Tabs
        value={tab}
        onChange={(_, value) => setTab(value)}
        variant="scrollable"
        scrollButtons="auto"
        aria-label={t('detectionCenter.title')}
        sx={{ mb: 3, borderBottom: 1, borderColor: 'divider' }}
      >
        {['plans', 'jobs'].map((key) => (
          <Tab
            id={`check-tab-${key}`}
            aria-controls={`check-panel-${key}`}
            key={key}
            value={key}
            label={t(`detectionCenter.tabs.${key}`)}
          />
        ))}
      </Tabs>
      {loadError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {loadError}
        </Alert>
      )}
      <Box role="tabpanel" id={`check-panel-${tab}`} aria-labelledby={`check-tab-${tab}`}>
        {loading ? (
          <Box sx={{ textAlign: 'center', py: 5 }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {tab === 'plans' && (
              <>
                <Paper variant="outlined" sx={{ overflow: 'hidden', bgcolor: 'background.paper' }}>
                  {!profiles.length && (
                    <Typography color="text.secondary" sx={{ p: 4, textAlign: 'center' }}>
                      {t('detectionCenter.emptyPlans')}
                    </Typography>
                  )}
                  {profiles.map((p) => (
                    <Box
                      key={p.id}
                      sx={{
                        display: 'grid',
                        gridTemplateColumns: { xs: '1fr', md: 'minmax(140px, 1fr) minmax(180px, 1.7fr) auto' },
                        gap: 2,
                        alignItems: 'center',
                        p: 2.5,
                        borderBottom: 1,
                        borderColor: 'divider'
                      }}
                    >
                      <Box sx={{ minWidth: 0 }}>
                        <Typography variant="subtitle1" sx={{ overflowWrap: 'anywhere' }}>
                          {p.name}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {t('detectionCenter.jobCount', { count: jobs.filter((job) => job.profileId === p.id).length })} ·{' '}
                          {t('nodes.nodeCheckProfiles.timeout', { seconds: p.timeout })}
                        </Typography>
                      </Box>
                      <Stack direction="row" gap={0.75} flexWrap="wrap">
                        {modules(p).map((label, index) => (
                          <Chip size="small" key={`${label}-${index}`} label={label} color="primary" variant="outlined" />
                        ))}
                      </Stack>
                      <Stack direction="row" spacing={0.5} alignItems="center">
                        <Button onClick={() => setPlanEditor(p)} disabled={busy}>
                          {t('detectionCenter.configure')}
                        </Button>
                        <Button variant="outlined" startIcon={<PlayArrowIcon />} onClick={() => openRun(p)} disabled={busy}>
                          {t('detectionCenter.run')}
                        </Button>
                        <Tooltip title={t('common.delete')}>
                          <span>
                            <IconButton aria-label={t('common.delete')} color="error" onClick={() => removePlan(p)} disabled={busy}>
                              <DeleteIcon fontSize="small" />
                            </IconButton>
                          </span>
                        </Tooltip>
                      </Stack>
                    </Box>
                  ))}
                </Paper>
              </>
            )}
            {tab === 'jobs' && (
              <Stack spacing={2}>
                {jobs.length === 0 && (
                  <Typography sx={{ p: 4, textAlign: 'center' }} color="text.secondary">
                    {t('detectionCenter.noJobs')}
                  </Typography>
                )}
                {jobs.map((job) => (
                  <Paper variant="outlined" key={job.id} sx={{ p: 2.5, bgcolor: 'background.paper' }}>
                    <Stack direction={{ xs: 'column', sm: 'row' }} justifyContent="space-between" gap={1}>
                      <Stack direction="row" flexWrap="wrap" gap={1} alignItems="center">
                        <Typography variant="subtitle1" sx={{ overflowWrap: 'anywhere' }}>
                          {job.name}
                        </Typography>
                        {job.legacy && <Chip label={t('detectionCenter.legacy')} size="small" variant="outlined" />}
                        {job.running && <Chip label={t('tasks.status.running')} size="small" color="info" />}
                      </Stack>
                      <FormControlLabel
                        sx={{ mr: 0 }}
                        label={t('detectionCenter.autoEnabled')}
                        control={
                          <Switch
                            checked={job.enabled}
                            disabled={busy}
                            onChange={(_, enabled) => action(() => saveNodeCheckJob({ ...job, enabled }), t('common.saveSuccess'))}
                          />
                        }
                      />
                    </Stack>
                    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', lg: '1fr 1fr 1fr' }, gap: 2, my: 2 }}>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          {t('detectionCenter.plan')}
                        </Typography>
                        <Typography>{profiles.find((p) => p.id === job.profileId)?.name || t('detectionCenter.missingPlan')}</Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          {t('detectionCenter.scope')}
                        </Typography>
                        <Typography>{t(`detectionCenter.scopes.${job.scope}`)}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          {[...(job.groups || []), ...(job.tags || [])].join(' · ')}
                        </Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          {t('detectionCenter.trigger')}
                        </Typography>
                        <Typography>{t(`detectionCenter.triggers.${job.trigger}`)}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          {job.trigger === 'scheduled' ? job.cronExpr : '—'} · {timezone}
                        </Typography>
                      </Box>
                    </Box>
                    <Typography variant="caption" display="block" color="text.secondary">
                      {t('detectionCenter.runTimes', { last: date(job.lastRunTime), next: date(job.nextRunTime) })}
                    </Typography>
                    {job.lastError && (
                      <Alert severity="warning" sx={{ mt: 1 }}>
                        {job.lastError}
                      </Alert>
                    )}
                    <Stack direction="row" spacing={1} sx={{ mt: 2 }} flexWrap="wrap">
                      <Button startIcon={<EditIcon />} onClick={() => setJobEditor(job)} disabled={busy}>
                        {t('common.edit')}
                      </Button>
                      <Button
                        startIcon={<PlayArrowIcon />}
                        onClick={() => action(() => runNodeCheckJob(job.id), t('nodes.nodeCheckProfiles.messages.started'))}
                        disabled={busy || job.running}
                      >
                        {t('detectionCenter.run')}
                      </Button>
                      <Button color="error" onClick={() => removeJob(job)} disabled={busy || job.running}>
                        {t('common.delete')}
                      </Button>
                    </Stack>
                  </Paper>
                ))}
              </Stack>
            )}
          </>
        )}
      </Box>
      <NodeCheckProfileFormDialog
        open={planEditor !== null}
        profile={planEditor?.id ? planEditor : null}
        onClose={() => setPlanEditor(null)}
        onSuccess={planSaved}
      />
      {jobEditor && (
        <CheckJobDialog
          job={jobEditor}
          profiles={profiles}
          options={options}
          timezone={timezone}
          onClose={() => setJobEditor(null)}
          onSaved={() => {
            setJobEditor(null);
            load();
            notify(t('common.saveSuccess'));
          }}
        />
      )}
      <Dialog open={!!runPlan} onClose={busy ? undefined : () => setRunPlan(null)} fullWidth maxWidth="sm">
        <DialogTitle>{t('detectionCenter.runPlan', { name: runPlan?.name })}</DialogTitle>
        <DialogContent sx={{ px: 2.5 }}>
          <Stack spacing={2} sx={{ mt: 2 }}>
            <TextField
              select
              label={t('detectionCenter.scope')}
              value={selection.scope}
              onChange={(e) => setSelection((prev) => ({ ...prev, scope: e.target.value }))}
            >
              {['all', 'filtered'].map((key) => (
                <MenuItem key={key} value={key}>
                  {t(`detectionCenter.scopes.${key}`)}
                </MenuItem>
              ))}
            </TextField>
            {selection.scope === 'filtered' && (
              <>
                {['groups', 'tags'].map((key) => (
                  <Autocomplete
                    key={key}
                    multiple
                    freeSolo
                    options={key === 'groups' ? options.groups : options.tags.map((tag) => tag.name || tag)}
                    value={selection[key]}
                    onChange={(_, value) => setSelection((prev) => ({ ...prev, [key]: value }))}
                    renderInput={(params) => <TextField {...params} label={t(`detectionCenter.${key}`)} />}
                  />
                ))}
              </>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={() => setRunPlan(null)}>
            {t('common.cancel')}
          </Button>
          <Button
            variant="contained"
            disabled={busy || (selection.scope === 'filtered' && !selection.groups.length && !selection.tags.length)}
            onClick={() =>
              action(async () => {
                await runNodeCheckSelection(runPlan.id, selection.scope === 'all' ? { scope: 'all' } : selection);
                setRunPlan(null);
              }, t('nodes.nodeCheckProfiles.messages.started'))
            }
          >
            {t('detectionCenter.run')}
          </Button>
        </DialogActions>
      </Dialog>
      <Snackbar open={!!message} autoHideDuration={5000} onClose={() => setMessage(null)}>
        <Alert severity={message?.severity || 'success'} onClose={() => setMessage(null)}>
          {message?.text}
        </Alert>
      </Snackbar>
    </MainCard>
  );
}
