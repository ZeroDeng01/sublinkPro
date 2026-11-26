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
  IconButton,
  MenuItem,
  Paper,
  Snackbar,
  Stack,
  Tab,
  Tabs,
  TextField,
  Tooltip,
  Typography
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import RefreshIcon from '@mui/icons-material/Refresh';
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
import { formatUnlockProvidersSummary, setUnlockMeta } from 'views/nodes/utils';
import NodeCheckProfileFormDialog from 'views/nodes/component/NodeCheckProfileFormDialog';
import CheckJobDialog from './components/CheckJobDialog';
import CheckJobList from './components/CheckJobList';

const listData = (response) => (Array.isArray(response?.data) ? response.data : response?.data?.items || []);

// 方案定义检测能力，任务定义编排；旧策略 ID 与单节点执行入口保持兼容。
export default function NodeCheckList() {
  const { t } = useTranslation();
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
    <Paper variant="outlined" sx={{ bgcolor: 'background.paper', borderColor: 'divider', borderRadius: 3, minWidth: 0 }}>
      <Stack
        direction="row"
        alignItems="center"
        flexWrap="wrap"
        gap={0.5}
        sx={{ px: { xs: 1.25, sm: 2 }, py: 0.5, borderBottom: 1, borderColor: 'divider' }}
      >
        <Tabs
          value={tab}
          onChange={(_, value) => setTab(value)}
          variant="scrollable"
          scrollButtons="auto"
          aria-label={t('detectionCenter.title')}
          sx={{ flex: '1 0 auto', maxWidth: '100%', minHeight: 44, '& .MuiTab-root': { minWidth: 0, minHeight: 44, px: 1.75 } }}
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
        <Stack direction="row" alignItems="center" gap={0.5} sx={{ ml: 'auto' }}>
          <Tooltip title={t('common.refresh')}>
            <span>
              <IconButton aria-label={t('common.refresh')} disabled={loading || busy} onClick={load}>
                <RefreshIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
          <Button
            size="small"
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => (tab === 'plans' ? setPlanEditor({}) : setJobEditor({}))}
            disabled={busy || (tab === 'jobs' && !profiles.length)}
            sx={{ minHeight: 36 }}
          >
            {t(tab === 'plans' ? 'detectionCenter.newPlan' : 'detectionCenter.newJob')}
          </Button>
        </Stack>
      </Stack>
      <Box sx={{ container: 'detection-center / inline-size', p: { xs: 1.25, sm: 2 } }}>
        {loadError && (
          <Alert severity="error" sx={{ mb: 1.5 }}>
            {loadError}
          </Alert>
        )}
        <Box role="tabpanel" id="check-panel-plans" aria-labelledby="check-tab-plans" hidden={tab !== 'plans'}>
          {tab === 'plans' &&
            (loading ? (
              <LoadingIndicator />
            ) : (
              <Paper variant="outlined" sx={{ overflow: 'hidden', bgcolor: 'background.paper' }}>
                {!profiles.length && (
                  <Typography color="text.secondary" sx={{ p: 3, textAlign: 'center' }}>
                    {t('detectionCenter.emptyPlans')}
                  </Typography>
                )}
                {profiles.map((p) => (
                  <Box
                    key={p.id}
                    sx={{
                      display: 'grid',
                      gridTemplateColumns: 'minmax(0, 1fr)',
                      '@container detection-center (min-width: 640px)': {
                        gridTemplateColumns: 'minmax(0, 1fr) minmax(0, 1.6fr) auto'
                      },
                      gap: 1,
                      alignItems: 'center',
                      px: 1.5,
                      py: 1,
                      '&:not(:last-child)': { borderBottom: 1, borderColor: 'divider' }
                    }}
                  >
                    <Box sx={{ minWidth: 0 }}>
                      <Typography variant="body2" sx={{ overflowWrap: 'anywhere', fontWeight: 600 }}>
                        {p.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {t('detectionCenter.jobCount', { count: jobs.filter((job) => job.profileId === p.id).length })} ·{' '}
                        {t('nodes.nodeCheckProfiles.timeout', { seconds: p.timeout })}
                      </Typography>
                    </Box>
                    <Stack direction="row" gap={0.5} flexWrap="wrap" sx={{ minWidth: 0 }}>
                      {modules(p).map((label, index) => (
                        <Chip
                          size="small"
                          key={`${label}-${index}`}
                          label={label}
                          color="primary"
                          variant="outlined"
                          sx={{ height: 'auto', minHeight: 24, '& .MuiChip-label': { whiteSpace: 'normal', overflowWrap: 'anywhere' } }}
                        />
                      ))}
                    </Stack>
                    <Stack direction="row" gap={0.5} flexWrap="wrap" alignItems="center">
                      <Button size="small" onClick={() => setPlanEditor(p)} disabled={busy}>
                        {t('detectionCenter.configure')}
                      </Button>
                      <Button size="small" variant="outlined" startIcon={<PlayArrowIcon />} onClick={() => openRun(p)} disabled={busy}>
                        {t('detectionCenter.run')}
                      </Button>
                      <Tooltip title={t('common.delete')}>
                        <span>
                          <IconButton
                            size="small"
                            aria-label={t('common.delete')}
                            color="error"
                            onClick={() => removePlan(p)}
                            disabled={busy}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </span>
                      </Tooltip>
                    </Stack>
                  </Box>
                ))}
              </Paper>
            ))}
        </Box>
        <Box role="tabpanel" id="check-panel-jobs" aria-labelledby="check-tab-jobs" hidden={tab !== 'jobs'}>
          {tab === 'jobs' &&
            (loading ? (
              <LoadingIndicator />
            ) : (
              <CheckJobList
                jobs={jobs}
                profiles={profiles}
                timezone={timezone}
                busy={busy}
                onEdit={setJobEditor}
                onRun={(job) => action(() => runNodeCheckJob(job.id), t('nodes.nodeCheckProfiles.messages.started'))}
                onRemove={removeJob}
                onToggle={(job, enabled) => action(() => saveNodeCheckJob({ ...job, enabled }), t('common.saveSuccess'))}
              />
            ))}
        </Box>
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
    </Paper>
  );
}

function LoadingIndicator() {
  return (
    <Box sx={{ textAlign: 'center', py: 3 }}>
      <CircularProgress size={28} />
    </Box>
  );
}
