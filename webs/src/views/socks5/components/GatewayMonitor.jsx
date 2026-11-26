import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import {
  closeAllSocks5Connections,
  closeSocks5Connection,
  getSocks5Connections,
  getSocks5RoutingProfiles,
  getSocks5RoutingSnapshot,
  probeSocks5Health,
  resetProxyGatewayStats,
  resetSocks5RuntimeStats
} from 'api/proxyGateway';
import { formatBytes } from 'views/airports/utils';

function formatElapsed(value) {
  if (!value) return '-';
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 1000));
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
}

function healthColor(status) {
  if (status === 'healthy') return 'success';
  if (status === 'unhealthy') return 'error';
  if (status === 'checking' || status === 'cooling') return 'warning';
  return 'default';
}

function formatCount(value) {
  return Number(value || 0).toLocaleString();
}

export default function GatewayMonitor({ showMessage, status, statusError, refreshStatus, probeSettings, focusNodes }) {
  const { t } = useTranslation();
  const label = (key, params) => t(`settings.socks5.monitor.${key}`, params);
  const [view, setView] = useState('connections');
  const [connections, setConnections] = useState([]);
  const [routingRead, setRoutingRead] = useState(null);
  const [profiles, setProfiles] = useState([]);
  const [selectedProfileID, setSelectedProfileID] = useState('');
  const [selectedListenerID, setSelectedListenerID] = useState('');
  const [nodeKeyword, setNodeKeyword] = useState('');
  const [debouncedKeyword, setDebouncedKeyword] = useState('');
  const [nodeStatus, setNodeStatus] = useState('all');
  const [nodePage, setNodePage] = useState(0);
  const [nodePageSize, setNodePageSize] = useState(25);
  const [sortBy, setSortBy] = useState('activeConnections');
  const [sortOrder, setSortOrder] = useState('desc');
  const [revision, setRevision] = useState(0);
  const [loaded, setLoaded] = useState({});
  const [error, setError] = useState('');
  const [action, setAction] = useState('');
  const [confirmation, setConfirmation] = useState('');
  const [gatewayResetRefreshPending, setGatewayResetRefreshPending] = useState(false);
  const [nodeResetRefreshPending, setNodeResetRefreshPending] = useState(false);
  const monitorRequest = useRef(null);
  const activeView = view;
  const runningListeners = (status?.listeners || []).filter((listener) => listener.running);
  const listenerID = runningListeners.some((listener) => listener.id === selectedListenerID) ? selectedListenerID : '';
  const listenerLabel = (id) => status?.listeners?.find((listener) => listener.id === id)?.name || id;
  const resettingNodes = action === 'nodeReset';

  useEffect(() => {
    if (status?.stats && !statusError) setGatewayResetRefreshPending(false);
  }, [status, statusError]);

  useEffect(() => {
    if (status && listenerID !== selectedListenerID) {
      setSelectedListenerID('');
      setNodePage(0);
    }
  }, [listenerID, selectedListenerID, status]);

  useEffect(() => {
    if (focusNodes) setView('nodes');
  }, [focusNodes]);

  useEffect(() => {
    const timer = setTimeout(() => {
      const keyword = nodeKeyword.trim();
      if (keyword !== debouncedKeyword) {
        setNodePage(0);
        setDebouncedKeyword(keyword);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [nodeKeyword, debouncedKeyword]);

  useEffect(() => {
    if (activeView !== 'nodes') return undefined;
    let active = true;
    getSocks5RoutingProfiles()
      .then((response) => {
        if (!active) return;
        const available = Array.isArray(response.data) ? response.data.filter((profile) => profile.enabled) : [];
        setProfiles(available);
        setSelectedProfileID((current) => (current === 'default' || available.some((profile) => profile.id === current) ? current : ''));
      })
      .catch((failure) => {
        if (active) showMessage(failure.message, 'error');
      });
    return () => {
      active = false;
    };
  }, [activeView, showMessage]);

  const load = useCallback(
    (signal) => {
      if (activeView === 'connections') return getSocks5Connections(signal);
      return getSocks5RoutingSnapshot(
        {
          listenerId: listenerID || undefined,
          profileId: selectedProfileID || undefined,
          allProfiles: selectedProfileID ? undefined : true,
          keyword: debouncedKeyword,
          status: nodeStatus === 'all' ? '' : nodeStatus,
          sortBy,
          sortOrder,
          page: nodePage + 1,
          pageSize: nodePageSize
        },
        signal
      );
    },
    [activeView, debouncedKeyword, listenerID, nodePage, nodePageSize, nodeStatus, selectedProfileID, sortBy, sortOrder]
  );

  useEffect(() => {
    if (activeView === 'nodes' && resettingNodes) return undefined;
    let controller;
    let timer;
    let stopped = false;
    const poll = async () => {
      if (document.visibilityState !== 'visible' || controller) return;
      controller = new AbortController();
      monitorRequest.current = controller;
      const { signal } = controller;
      try {
        const response = await load(signal);
        if (signal.aborted) return;
        if (activeView === 'connections') {
          setConnections(Array.isArray(response.data) ? response.data : []);
        } else {
          setRoutingRead({ query: load, data: response.data });
          setNodeResetRefreshPending(false);
          if (response.data?.page > 0) setNodePage(response.data.page - 1);
        }
        setLoaded((current) => ({ ...current, [activeView]: true }));
        setError('');
      } catch (failure) {
        if (!signal.aborted) setError(failure.message);
      } finally {
        if (monitorRequest.current?.signal === signal) monitorRequest.current = null;
        if (!signal.aborted) {
          controller = null;
          if (!stopped) timer = window.setTimeout(poll, 3000);
        }
      }
    };
    const onVisibility = () => {
      window.clearTimeout(timer);
      if (document.visibilityState === 'visible') {
        void poll();
      } else {
        controller?.abort();
        controller = null;
      }
    };
    void poll();
    document.addEventListener('visibilitychange', onVisibility);
    return () => {
      stopped = true;
      controller?.abort();
      window.clearTimeout(timer);
      document.removeEventListener('visibilitychange', onVisibility);
    };
  }, [activeView, load, resettingNodes, revision]);

  const runAction = async (name, request) => {
    if (name === 'nodeReset') monitorRequest.current?.abort();
    setAction(name);
    try {
      const response = await request();
      if (name === 'probe') {
        showMessage(
          t(
            response.data?.started === false ? 'settings.socks5.messages.healthProbeRunning' : 'settings.socks5.messages.healthProbeStarted'
          )
        );
      } else if (name === 'nodeReset') {
        setRoutingRead(null);
        setNodeResetRefreshPending(true);
        setError('');
        showMessage(t('settings.socks5.messages.runtimeStatsReset'));
      } else if (name === 'gatewayReset') {
        const refreshed = await refreshStatus(true);
        setGatewayResetRefreshPending(!refreshed);
        showMessage(
          t(refreshed ? 'settings.socks5.monitor.gatewayStatsReset' : 'settings.socks5.monitor.resetRefreshFailed'),
          refreshed ? 'success' : 'warning'
        );
      }
      if (name !== 'gatewayReset') {
        setRevision((current) => current + 1);
        if (name !== 'nodeReset') void refreshStatus(true);
      }
    } catch (failure) {
      showMessage(failure.message, 'error');
    } finally {
      setAction('');
    }
  };

  const confirmReset = () => {
    const name = confirmation;
    setConfirmation('');
    void runAction(name, name === 'gatewayReset' ? resetProxyGatewayStats : resetSocks5RuntimeStats);
  };

  const changeSort = (field) => {
    setNodePage(0);
    if (sortBy === field) {
      setSortOrder((previous) => (previous === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortBy(field);
      setSortOrder('asc');
    }
  };

  const stats = status?.stats;
  const routing = routingRead?.query === load ? routingRead.data : null;
  const routingItems = Array.isArray(routing?.items) ? routing.items : [];
  const routingSummary = routing?.summary || {};
  const gatewayRefreshFailed = gatewayResetRefreshPending && Boolean(statusError);
  const formatMetric = (field, formatter = formatCount) => (stats && !gatewayRefreshFailed ? formatter(stats[field] ?? 0) : '—');
  const sortableHeader = (field, label, align = 'left') => (
    <TableCell align={align} sortDirection={sortBy === field ? sortOrder : false}>
      <TableSortLabel active={sortBy === field} direction={sortBy === field ? sortOrder : 'asc'} onClick={() => changeSort(field)}>
        {label}
      </TableSortLabel>
    </TableCell>
  );

  return (
    <Card variant="outlined">
      <CardHeader
        title={label('title')}
        subheader={label('subheader')}
        action={
          <Button
            variant="outlined"
            size="small"
            disabled={Boolean(action) || !status?.running}
            onClick={() => setConfirmation('gatewayReset')}
          >
            {label(action === 'gatewayReset' ? 'resettingStats' : 'resetGatewayStats')}
          </Button>
        }
        sx={{
          alignItems: { xs: 'flex-start', sm: 'center' },
          flexWrap: 'wrap',
          gap: 1,
          '& .MuiCardHeader-action': { alignSelf: 'center', m: 0 }
        }}
      />
      <CardContent sx={{ pt: 0 }}>
        {gatewayRefreshFailed && (
          <Alert
            severity="warning"
            sx={{ mb: 2 }}
            action={
              <Button color="inherit" size="small" onClick={() => refreshStatus(true)}>
                {t('settings.socks5.page.retry')}
              </Button>
            }
          >
            {label('resetRefreshFailed')}
          </Alert>
        )}
        <Box
          component="dl"
          sx={{
            display: 'grid',
            gridTemplateColumns: { xs: 'repeat(2, 1fr)', sm: 'repeat(3, 1fr)', lg: 'repeat(6, 1fr)' },
            gap: 1,
            m: 0
          }}
        >
          {[
            ['activeConnections', 'primary.main'],
            ['totalConnections'],
            ['successfulConnections', 'success.main'],
            ['failedConnections', 'error.main'],
            ['uploadBytes', undefined, formatBytes],
            ['downloadBytes', undefined, formatBytes]
          ].map(([field, color, formatter]) => (
            <Box key={field} sx={{ minWidth: 0, px: 1.5, py: 1, border: 1, borderColor: 'divider', borderRadius: 1 }}>
              <Typography component="dt" variant="caption" color="text.secondary">
                {label(`summaryLabels.${field}`)}
              </Typography>
              <Typography component="dd" variant="h4" color={color} sx={{ m: 0, mt: 0.25, fontVariantNumeric: 'tabular-nums' }}>
                {formatMetric(field, formatter)}
              </Typography>
            </Box>
          ))}
        </Box>
        <Tabs
          value={activeView}
          onChange={(_, next) => {
            setError('');
            setView(next);
          }}
          aria-label={t('settings.socks5.monitor.title')}
          sx={{ mt: 1.5, borderBottom: 1, borderColor: 'divider', '& .MuiTab-root': { textTransform: 'none', px: 1.5 } }}
        >
          {['connections', 'nodes'].map((name) => (
            <Tab
              key={name}
              value={name}
              id={`gateway-monitor-tab-${name}`}
              aria-controls={`gateway-monitor-panel-${name}`}
              label={t(`settings.socks5.page.monitorViews.${name}`)}
            />
          ))}
        </Tabs>
        {error && (
          <Alert
            severity="error"
            sx={{ mt: 2 }}
            action={
              <Button color="inherit" size="small" onClick={() => setRevision((value) => value + 1)}>
                {t('settings.socks5.page.retry')}
              </Button>
            }
          >
            {nodeResetRefreshPending && activeView === 'nodes' ? label('resetRefreshFailed') : error}
          </Alert>
        )}
        {(!loaded[activeView] || (activeView === 'nodes' && status?.running && !routing)) && !error && (
          <Typography sx={{ mt: 2 }} color="text.secondary" role="status">
            {t('settings.socks5.page.loading')}
          </Typography>
        )}
        <Box
          role="tabpanel"
          id="gateway-monitor-panel-connections"
          aria-labelledby="gateway-monitor-tab-connections"
          hidden={activeView !== 'connections'}
        >
          {activeView === 'connections' && (
            <>
              <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end' }}>
                <Button
                  color="error"
                  disabled={!connections.length || Boolean(action)}
                  onClick={() => {
                    if (window.confirm(t('settings.socks5.page.disconnectAllConfirm'))) {
                      void runAction('disconnectAll', closeAllSocks5Connections);
                    }
                  }}
                >
                  {t('settings.socks5.monitor.disconnectAll')}
                </Button>
              </Box>
              {loaded.connections && (
                <Box sx={{ mt: 2 }}>
                  {connections.length
                    ? connections.map((connection) => (
                        <Stack
                          key={connection.id}
                          direction={{ xs: 'column', md: 'row' }}
                          spacing={1}
                          alignItems={{ md: 'center' }}
                          sx={{ py: 1, borderBottom: 1, borderColor: 'divider' }}
                        >
                          <Typography sx={{ flex: 1, overflowWrap: 'anywhere' }}>
                            {connection.target || t('settings.socks5.monitor.connecting')} ·{' '}
                            {connection.nodeName || t('settings.socks5.monitor.waitingNode')}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {connection.profileName || connection.profileId || 'Default'}
                            {connection.account ? ` · ${connection.account}` : ''}
                            {connection.listenerId ? ` · ${listenerLabel(connection.listenerId)}` : ''}
                          </Typography>
                          <Typography variant="caption">
                            {t(
                              `settings.socks5.monitor.protocols.${connection.protocol === 'http' && connection.method === 'CONNECT' ? 'httpConnect' : connection.protocol || 'socks5'}`
                            )}{' '}
                            · {connection.clientAddress} ·{' '}
                            {t(`settings.socks5.monitor.phases.${connection.phase}`, { defaultValue: connection.phase })} · ↑{' '}
                            {formatBytes(connection.uploadBytes ?? 0)} · ↓ {formatBytes(connection.downloadBytes ?? 0)} ·{' '}
                            {t('settings.socks5.monitor.duration', { value: formatElapsed(connection.startedAt) })} ·{' '}
                            {t('settings.socks5.monitor.idle', { value: formatElapsed(connection.lastActivity) })}
                          </Typography>
                          <Button
                            size="small"
                            color="error"
                            disabled={Boolean(action)}
                            onClick={() => runAction(connection.id, () => closeSocks5Connection(connection.id))}
                          >
                            {t('settings.socks5.monitor.disconnect')}
                          </Button>
                        </Stack>
                      ))
                    : !error && <Typography color="text.secondary">{t('settings.socks5.monitor.empty')}</Typography>}
                </Box>
              )}
            </>
          )}
        </Box>
        <Box role="tabpanel" id="gateway-monitor-panel-nodes" aria-labelledby="gateway-monitor-tab-nodes" hidden={activeView !== 'nodes'}>
          {activeView === 'nodes' && (
            <>
              <Stack
                direction={{ xs: 'column', sm: 'row' }}
                justifyContent="space-between"
                alignItems={{ sm: 'flex-start' }}
                gap={1}
                sx={{ mt: 2 }}
              >
                <Box>
                  <Typography variant="subtitle1">{label('nodeRouting')}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {label('nodeRoutingHelper')}
                  </Typography>
                </Box>
                <Stack direction="row" useFlexGap flexWrap="wrap" spacing={1} sx={{ flexShrink: 0 }}>
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() =>
                      runAction('probe', () =>
                        probeSocks5Health({ listenerId: listenerID || undefined, profileId: selectedProfileID || undefined })
                      )
                    }
                    disabled={Boolean(action) || routing?.probeRunning || !status?.running}
                  >
                    {label(action === 'probe' || routing?.probeRunning ? 'probing' : 'probeNow')}
                  </Button>
                  <Button size="small" onClick={() => setConfirmation('nodeReset')} disabled={Boolean(action) || !status?.running}>
                    {label(resettingNodes ? 'resettingStats' : 'resetStats')}
                  </Button>
                </Stack>
              </Stack>
              <Box sx={{ mt: 1.5 }}>
                {routing && (
                  <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" sx={{ mb: 1.5 }}>
                    <Chip label={t('settings.socks5.monitor.routingTotalNodes', { count: formatCount(routingSummary.totalNodes) })} />
                    <Chip
                      label={t('settings.socks5.monitor.routingActiveNodes', { count: formatCount(routingSummary.activeNodes) })}
                      color="primary"
                    />
                    <Chip
                      label={t('settings.socks5.monitor.routingActiveConnections', {
                        count: formatCount(routingSummary.activeConnections)
                      })}
                    />
                    <Chip
                      label={t('settings.socks5.monitor.routingCoolingNodes', { count: formatCount(routingSummary.coolingNodes) })}
                      color="warning"
                    />
                    <Chip
                      label={t('settings.socks5.monitor.routingSuccesses', { count: formatCount(routingSummary.successfulConnections) })}
                      color="success"
                    />
                    <Chip
                      label={t('settings.socks5.monitor.routingFailures', { count: formatCount(routingSummary.failedConnections) })}
                      color="error"
                    />
                  </Stack>
                )}
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} sx={{ mb: 1.5 }}>
                  <FormControl size="small" sx={{ minWidth: { sm: 150 } }}>
                    <InputLabel id="gateway-monitor-listener-label">{label('listenerFilter')}</InputLabel>
                    <Select
                      labelId="gateway-monitor-listener-label"
                      label={label('listenerFilter')}
                      value={listenerID || '*'}
                      onChange={(event) => {
                        setSelectedListenerID(event.target.value === '*' ? '' : event.target.value);
                        setNodePage(0);
                      }}
                    >
                      <MenuItem value="*">{label('allListeners')}</MenuItem>
                      {runningListeners.map((listener) => (
                        <MenuItem key={listener.id} value={listener.id}>
                          {listener.name || listener.id}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <FormControl size="small" sx={{ minWidth: { sm: 180 } }}>
                    <InputLabel id="gateway-monitor-profile-label">{t('settings.socks5.monitor.profileFilter')}</InputLabel>
                    <Select
                      labelId="gateway-monitor-profile-label"
                      value={selectedProfileID || '*'}
                      label={t('settings.socks5.monitor.profileFilter')}
                      onChange={(event) => {
                        setSelectedProfileID(event.target.value === '*' ? '' : event.target.value);
                        setNodePage(0);
                      }}
                    >
                      <MenuItem value="*">{label('allProfiles')}</MenuItem>
                      {!profiles.some((profile) => profile.enabled && profile.id === 'default') && (
                        <MenuItem value="default">{t('settings.socks5.profiles.default')}</MenuItem>
                      )}
                      {profiles
                        .filter((profile) => profile.enabled)
                        .map((profile) => (
                          <MenuItem key={profile.id} value={profile.id}>
                            {profile.isDefault ? t('settings.socks5.profiles.default') : profile.name}
                          </MenuItem>
                        ))}
                    </Select>
                  </FormControl>
                  <TextField
                    size="small"
                    fullWidth
                    label={t('settings.socks5.monitor.nodeSearch')}
                    placeholder={t('settings.socks5.monitor.nodeSearchPlaceholder')}
                    value={nodeKeyword}
                    onChange={(event) => setNodeKeyword(event.target.value)}
                  />
                  <FormControl size="small" sx={{ minWidth: { sm: 180 } }}>
                    <InputLabel id="gateway-monitor-health-label">{t('settings.socks5.monitor.nodeStatusFilter')}</InputLabel>
                    <Select
                      labelId="gateway-monitor-health-label"
                      value={nodeStatus}
                      label={t('settings.socks5.monitor.nodeStatusFilter')}
                      onChange={(event) => {
                        setNodeStatus(event.target.value);
                        setNodePage(0);
                      }}
                    >
                      {['all', 'healthy', 'unhealthy', 'checking', 'cooling', 'unknown'].map((value) => (
                        <MenuItem key={value} value={value}>
                          {t(`settings.socks5.monitor.healthStatus.${value}`)}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Stack>
                {routing && (
                  <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" sx={{ mb: 1.5 }}>
                    <Chip size="small" label={label('healthyNodes', { count: routingSummary.healthyNodes ?? 0 })} color="success" />
                    <Chip size="small" label={label('unhealthyNodes', { count: routingSummary.unhealthyNodes ?? 0 })} color="error" />
                    <Chip size="small" label={label('checkingNodes', { count: routingSummary.checkingNodes ?? 0 })} color="warning" />
                    <Chip size="small" label={label('unknownNodes', { count: routingSummary.unknownNodes ?? 0 })} />
                  </Stack>
                )}
                {status && !status.running && <Alert severity="info">{t('settings.socks5.monitor.routingStopped')}</Alert>}
                {status?.running && routing && !error && routingItems.length === 0 && (
                  <Alert severity="info">{t('settings.socks5.monitor.noRoutingData')}</Alert>
                )}
                {routingItems.length > 0 && (
                  <TableContainer sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
                    <Table size="small" sx={{ minWidth: 820 }}>
                      <TableHead>
                        <TableRow>
                          {sortableHeader('nodeName', t('settings.socks5.monitor.columns.node'))}
                          {sortableHeader('activeConnections', t('settings.socks5.monitor.columns.active'), 'right')}
                          {sortableHeader('successfulConnections', t('settings.socks5.monitor.columns.successful'), 'right')}
                          {sortableHeader('failedConnections', t('settings.socks5.monitor.columns.failed'), 'right')}
                          {sortableHeader('status', t('settings.socks5.monitor.columns.status'))}
                          {sortableHeader('latencyMs', t('settings.socks5.monitor.columns.latency'), 'right')}
                          {sortableHeader('smartScore', t('settings.socks5.monitor.columns.score'), 'right')}
                          {sortableHeader('lastSelectedAt', t('settings.socks5.monitor.columns.lastSelected'))}
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {routingItems.map((node) => (
                          <TableRow key={`${node.nodeId}-${node.nodeName}`} hover>
                            <TableCell sx={{ minWidth: 200 }}>
                              <Typography variant="body2">{node.nodeName || `#${node.nodeId}`}</Typography>
                              <Typography variant="caption" color="text.secondary">
                                {[node.group, node.source, node.protocol, node.country].filter(Boolean).join(' · ') || `#${node.nodeId}`}
                              </Typography>
                            </TableCell>
                            <TableCell align="right" sx={{ fontVariantNumeric: 'tabular-nums' }}>
                              {formatCount(node.activeConnections)}
                            </TableCell>
                            <TableCell align="right" sx={{ fontVariantNumeric: 'tabular-nums' }}>
                              {formatCount(node.successfulConnections)}
                            </TableCell>
                            <TableCell align="right" sx={{ fontVariantNumeric: 'tabular-nums' }}>
                              {formatCount(node.failedConnections)}
                            </TableCell>
                            <TableCell sx={{ minWidth: 145 }}>
                              <Chip
                                size="small"
                                color={healthColor(node.status)}
                                label={t(`settings.socks5.monitor.healthStatus.${node.status || 'unknown'}`)}
                              />
                              {node.excludedFromRouting && (
                                <Typography variant="caption" color="error" display="block">
                                  {t('settings.socks5.monitor.excludedFromRouting')}
                                </Typography>
                              )}
                              {!listenerID && node.healthListenerId && (
                                <Typography variant="caption" color="text.secondary" display="block">
                                  {label('healthSource', { listener: listenerLabel(node.healthListenerId) })}
                                </Typography>
                              )}
                              {(Boolean(node.consecutiveFailures) || node.cooldownUntil || node.lastError) && (
                                <Box component="details" sx={{ mt: 0.5 }}>
                                  <Typography component="summary" variant="caption" color="primary" sx={{ cursor: 'pointer' }}>
                                    {label('healthDetails')}
                                  </Typography>
                                  {Boolean(node.consecutiveFailures) && (
                                    <Typography variant="caption" display="block" color="error">
                                      {label('consecutiveFailuresShort', { count: node.consecutiveFailures })}
                                    </Typography>
                                  )}
                                  {node.cooldownUntil && (
                                    <Typography variant="caption" display="block" color="warning.main">
                                      {label('cooldownUntil', { value: new Date(node.cooldownUntil).toLocaleString() })}
                                    </Typography>
                                  )}
                                  {node.lastError && (
                                    <Typography
                                      variant="caption"
                                      display="block"
                                      color="error"
                                      sx={{ maxWidth: 240, overflowWrap: 'anywhere' }}
                                    >
                                      {node.lastError}
                                    </Typography>
                                  )}
                                </Box>
                              )}
                            </TableCell>
                            <TableCell align="right" sx={{ whiteSpace: 'nowrap' }}>
                              <Typography variant="body2">{node.latencyMs > 0 ? `${node.latencyMs} ms` : '-'}</Typography>
                              <Typography variant="caption" color="text.secondary">
                                {t(`settings.socks5.monitor.latencySource.${node.latencySource || 'fallback'}`)}
                              </Typography>
                            </TableCell>
                            <TableCell align="right">{Number(node.smartScore ?? 0).toFixed(1)}</TableCell>
                            <TableCell sx={{ minWidth: 180 }}>
                              <Typography variant="body2">
                                {node.lastSelectedAt ? new Date(node.lastSelectedAt).toLocaleString() : '-'}
                              </Typography>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                )}
                <TablePagination
                  component="div"
                  count={routing?.total ?? 0}
                  page={Math.min(nodePage, Math.max(0, Math.ceil((routing?.total ?? 0) / nodePageSize) - 1))}
                  onPageChange={(_, nextPage) => setNodePage(nextPage)}
                  rowsPerPage={nodePageSize}
                  onRowsPerPageChange={(event) => {
                    setNodePageSize(Number(event.target.value));
                    setNodePage(0);
                  }}
                  rowsPerPageOptions={[10, 25, 50, 100]}
                  labelRowsPerPage={t('settings.socks5.monitor.rowsPerPage')}
                />
              </Box>
              <Box sx={{ mt: 1.5 }}>{probeSettings}</Box>
            </>
          )}
        </Box>
      </CardContent>
      <Dialog
        open={Boolean(confirmation)}
        onClose={() => setConfirmation('')}
        aria-labelledby="gateway-monitor-reset-title"
        aria-describedby="gateway-monitor-reset-description"
      >
        <DialogTitle id="gateway-monitor-reset-title">
          {label(confirmation === 'gatewayReset' ? 'resetGatewayStats' : 'resetStats')}
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="gateway-monitor-reset-description">
            {label(confirmation === 'gatewayReset' ? 'resetGatewayConfirm' : 'resetNodesConfirm')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button autoFocus onClick={() => setConfirmation('')}>
            {t('common.cancel')}
          </Button>
          <Button variant="contained" onClick={confirmReset}>
            {label('confirmReset')}
          </Button>
        </DialogActions>
      </Dialog>
    </Card>
  );
}
