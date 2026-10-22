import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  LinearProgress,
  MenuItem,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TextField,
  Typography
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { getNodeCheckResults } from 'api/nodeCheck';
import { getTasks, stopTask } from 'api/tasks';
import { formatDateTime } from 'i18n/locales';
import {
  getDelayDisplay,
  getSpeedDisplay,
  getFraudScoreDisplay,
  getNodeUnlockSummaryDisplay,
  getUnlockProviderOptions,
  getUnlockStatusOptions
} from 'views/nodes/utils';

// 历史使用任务系统分页；节点结果使用最新状态分页，两者不建立虚假的历史关联。
export default function DetectionDataPanel({ kind, revision, onRun }) {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [page, setPage] = useState(0);
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const [provider, setProvider] = useState('');
  const [selected, setSelected] = useState(null);
  const [tick, setTick] = useState(0);
  const [stopping, setStopping] = useState(false);
  const history = kind === 'runs';
  const time = (value) =>
    !value || String(value).startsWith('0001-')
      ? '—'
      : formatDateTime(value, i18n.resolvedLanguage || i18n.language, { dateStyle: 'short', timeStyle: 'medium' }) || '—';
  useEffect(() => {
    let active = true;
    const controller = new AbortController();
    setLoading(true);
    setError('');
    const timer = setTimeout(async () => {
      try {
        const response = history
          ? await getTasks({ type: 'speed_test', status, page: page + 1, pageSize: 30 })
          : await getNodeCheckResults(
              { page: page + 1, pageSize: 30, search, unlockProvider: provider, unlockStatus: status },
              { signal: controller.signal }
            );
        if (active) {
          const nextItems = response.data?.items || [];
          const nextTotal = response.data?.total || 0;
          setItems(nextItems);
          setTotal(nextTotal);
          setPage((current) => Math.min(current, Math.max(0, Math.ceil(nextTotal / 30) - 1)));
          setSelected((current) =>
            current ? nextItems.find((item) => (history ? item.id === current.id : item.node.ID === current.node.ID)) || null : null
          );
        }
      } catch (err) {
        if (active && err.code !== 'ERR_CANCELED') {
          setError(err.message || t('detectionCenter.loadFailed'));
          setItems([]);
          setTotal(0);
        }
      } finally {
        if (active) setLoading(false);
      }
    }, 250);
    return () => {
      active = false;
      clearTimeout(timer);
      controller.abort();
    };
  }, [history, revision, page, search, status, provider, tick, t]);
  useEffect(() => {
    if (!history || !items.some((item) => ['running', 'pending'].includes(item.status))) return;
    const timer = setTimeout(() => setTick((n) => n + 1), 5000);
    return () => clearTimeout(timer);
  }, [history, items]);
  const stop = async (id) => {
    setStopping(true);
    try {
      await stopTask(id);
      setTick((n) => n + 1);
      setSelected(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setStopping(false);
    }
  };
  const filter = (set) => (e) => {
    set(e.target.value);
    setPage(0);
  };
  const quality = (n) => getFraudScoreDisplay(n.FraudScore, n.QualityStatus, n.QualityFamily);
  const chip = (display) => <Chip size="small" label={display.label} color={display.color || 'default'} variant="outlined" />;
  const providerChips = (node) => {
    const display = getNodeUnlockSummaryDisplay(node, { limit: 100 });
    if (!display)
      return (
        <Typography variant="caption" color="text.secondary">
          {t('detectionCenter.untested')}
        </Typography>
      );
    const ordered = provider
      ? [...display.items.filter((item) => item.provider === provider), ...display.items.filter((item) => item.provider !== provider)]
      : display.items;
    return (
      <>
        {ordered.slice(0, 2).map((item) => (
          <Chip key={item.provider} label={item.compactLabel} color={item.color || 'default'} variant="outlined" size="small" />
        ))}
        {ordered.length > 2 && (
          <Typography variant="caption" color="text.secondary">
            {t('detectionCenter.moreProviders', { count: ordered.length - 2 })}
          </Typography>
        )}
      </>
    );
  };
  const resultNode = selected?.node;
  const unlock = resultNode ? getNodeUnlockSummaryDisplay(resultNode, { limit: 100 }) : null;
  return (
    <Stack spacing={2}>
      <Typography variant="body2" color="text.secondary">
        {t(history ? 'detectionCenter.historyHelp' : 'detectionCenter.resultsHelp')}
      </Typography>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
        {!history && (
          <TextField size="small" label={t('detectionCenter.search')} value={search} onChange={filter(setSearch)} sx={{ flex: 1 }} />
        )}
        {!history && (
          <TextField
            select
            size="small"
            label={t('detectionCenter.provider')}
            value={provider}
            onChange={filter(setProvider)}
            sx={{ minWidth: 160 }}
          >
            <MenuItem value="">{t('common.all')}</MenuItem>
            {getUnlockProviderOptions().map((p) => (
              <MenuItem key={p.value} value={p.value}>
                {p.label}
              </MenuItem>
            ))}
          </TextField>
        )}
        <TextField select size="small" label={t('common.status')} value={status} onChange={filter(setStatus)} sx={{ minWidth: 150 }}>
          <MenuItem value="">{t('common.all')}</MenuItem>
          {history
            ? ['running', 'pending', 'completed', 'cancelled', 'error'].map((key) => (
                <MenuItem key={key} value={key}>
                  {t(`tasks.status.${key}`)}
                </MenuItem>
              ))
            : getUnlockStatusOptions(false).map((s) => (
                <MenuItem key={s.value} value={s.value}>
                  {s.label}
                </MenuItem>
              ))}
        </TextField>
        <Button onClick={() => setTick((n) => n + 1)} disabled={loading}>
          {t('common.refresh')}
        </Button>
      </Stack>
      {error && <Alert severity="error">{error}</Alert>}
      {loading && <LinearProgress aria-label={t('detectionCenter.loading')} />}
      <TableContainer sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 2, bgcolor: 'background.paper' }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              {(history
                ? ['plan', 'trigger', 'runTime', 'status', 'progress', 'actions']
                : ['node', 'latencySpeed', 'exit', 'quality', 'services']
              ).map((key) => (
                <TableCell key={key} sx={{ whiteSpace: 'nowrap' }}>
                  {t(`detectionCenter.${key}`)}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {!items.length && !loading && (
              <TableRow>
                <TableCell colSpan={history ? 6 : 5} sx={{ py: 5, textAlign: 'center', color: 'text.secondary' }}>
                  {t('common.empty')}
                </TableCell>
              </TableRow>
            )}
            {items.map((item) =>
              history ? (
                <TableRow key={item.id} hover>
                  <TableCell>
                    <Button onClick={() => setSelected(item)} sx={{ textAlign: 'left', overflowWrap: 'anywhere' }}>
                      {item.name}
                    </Button>
                  </TableCell>
                  <TableCell>{t(`tasks.trigger.${item.trigger}`, { defaultValue: item.trigger })}</TableCell>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>{time(item.startedAt || item.createdAt)}</TableCell>
                  <TableCell>
                    <Chip
                      size="small"
                      color={item.status === 'completed' ? 'success' : item.status === 'error' ? 'error' : 'default'}
                      label={t(`tasks.status.${item.status}`, { defaultValue: item.status })}
                      variant="outlined"
                    />
                  </TableCell>
                  <TableCell>
                    {item.progress} / {item.total}
                  </TableCell>
                  <TableCell>
                    <Button onClick={() => setSelected(item)}>{t('detectionCenter.details')}</Button>
                  </TableCell>
                </TableRow>
              ) : (
                <TableRow key={item.node.ID} hover>
                  <TableCell>
                    <Button onClick={() => setSelected(item)} sx={{ textAlign: 'left', overflowWrap: 'anywhere' }}>
                      {item.node.EffectiveName || item.node.LinkName || item.node.Name}
                    </Button>
                    <Typography variant="caption" display="block" color="text.secondary">
                      {item.airportName ||
                        (item.airportId ? t('detectionCenter.missingAirport', { id: item.airportId }) : t('detectionCenter.manualNode'))}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Stack spacing={0.5} alignItems="flex-start">
                      {chip(getDelayDisplay(item.node.DelayTime, item.node.DelayStatus))}
                      {chip(getSpeedDisplay(item.node.Speed, item.node.SpeedStatus))}
                    </Stack>
                  </TableCell>
                  <TableCell>
                    {item.node.LinkCountry || '—'}
                    <Typography variant="caption" display="block" color="text.secondary">
                      {item.node.LandingIP || '—'}
                    </Typography>
                  </TableCell>
                  <TableCell>{chip(quality(item.node))}</TableCell>
                  <TableCell>
                    <Stack spacing={0.5} alignItems="flex-start">
                      {providerChips(item.node)}
                    </Stack>
                  </TableCell>

                </TableRow>
              )
            )}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        component="div"
        count={total}
        page={page}
        rowsPerPage={30}
        rowsPerPageOptions={[30]}
        onPageChange={(_, value) => {
          setPage(value);
          setSelected(null);
        }}
      />
      <Dialog open={!!selected} onClose={() => setSelected(null)} fullWidth maxWidth="sm">
        <DialogTitle>{history ? selected?.name : resultNode?.EffectiveName || resultNode?.Name}</DialogTitle>
        <DialogContent sx={{ px: 2.5 }}>
          <Stack spacing={2} sx={{ mt: 2 }}>
            {history && selected && (
              <>
                <Typography>{selected.message || t(`tasks.status.${selected.status}`)}</Typography>
                <Typography color="text.secondary">{selected.currentItem}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {time(selected.startedAt)} → {time(selected.completedAt)}
                </Typography>
                <Alert severity="info">{t('detectionCenter.notSnapshot')}</Alert>
              </>
            )}
            {!history && resultNode && (
              <>
                <Alert severity="warning">{t('detectionCenter.probeWarning')}</Alert>
                <Typography>{t('detectionCenter.checkedAt', { time: time(resultNode.UnlockCheckAt) })}</Typography>
                {unlock?.items.map((u) => (
                  <Box key={u.provider}>
                    <Chip label={`${u.providerLabel} · ${u.statusLabel}`} color={u.color || 'default'} variant="outlined" />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1, overflowWrap: 'anywhere' }}>
                      {[u.region, u.reason, u.detail].filter(Boolean).join(' · ') || t('detectionCenter.noEvidence')}
                    </Typography>
                  </Box>
                ))}
                {!unlock && <Typography color="text.secondary">{t('detectionCenter.untested')}</Typography>}
                <Typography variant="body2">{t('detectionCenter.latencyCheckedAt', { time: time(resultNode.LatencyCheckAt) })}</Typography>
                <Typography variant="body2">{t('detectionCenter.speedCheckedAt', { time: time(resultNode.SpeedCheckAt) })}</Typography>

              </>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          {history ? (
            <>
              <Button onClick={() => navigate('/system/tasks')}>{t('detectionCenter.taskCenter')}</Button>
              {['running', 'pending'].includes(selected?.status) && (
                <Button color="error" onClick={() => stop(selected.id)} disabled={stopping}>
                  {stopping ? <CircularProgress size={16} /> : t('detectionCenter.stop')}
                </Button>
              )}
            </>
          ) : (
            resultNode && (
              <Button
                onClick={() => {
                  onRun([resultNode.ID]);
                  setSelected(null);
                }}
              >
                {t('detectionCenter.recheckNode')}
              </Button>
            )
          )}
          <Button onClick={() => setSelected(null)}>{t('common.close')}</Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
