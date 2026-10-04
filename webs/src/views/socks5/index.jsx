import { useCallback, useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Snackbar from '@mui/material/Snackbar';
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

import { useAuth } from 'contexts/AuthContext';
import { useTranslation } from 'react-i18next';
import { formatBytes } from 'views/airports/utils';
import RoutingProfiles from './components/RoutingProfiles';
import Socks5Accounts from './components/Socks5Accounts';
import Socks5Listeners from './components/Socks5Listeners';
import Socks5Settings from './components/Socks5Settings';
import {
  closeAllSocks5Connections,
  closeSocks5Connection,
  getSocks5Connections,
  getSocks5RoutingProfiles,
  getSocks5RoutingSnapshot,
  getSocks5Status,
  probeSocks5Health,
  resetSocks5RuntimeStats
} from 'api/settings';

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

function Socks5Monitor({ showMessage }) {
  const { t } = useTranslation();
  const [status, setStatus] = useState(null);
  const [connections, setConnections] = useState([]);
  const [routing, setRouting] = useState({ items: [], summary: {}, total: 0, page: 1, pageSize: 25 });
  const [profiles, setProfiles] = useState([]);
  const [selectedProfileID, setSelectedProfileID] = useState('default');
  const [probing, setProbing] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [nodeKeyword, setNodeKeyword] = useState('');
  const [debouncedKeyword, setDebouncedKeyword] = useState('');
  const [nodeStatus, setNodeStatus] = useState('all');
  const [nodePage, setNodePage] = useState(0);
  const [nodePageSize, setNodePageSize] = useState(25);
  const [sortBy, setSortBy] = useState('smartScore');
  const [sortOrder, setSortOrder] = useState('asc');

  useEffect(() => {
    const timer = setTimeout(() => {
      setNodePage(0);
      setDebouncedKeyword(nodeKeyword.trim());
    }, 300);
    return () => clearTimeout(timer);
  }, [nodeKeyword]);

  const load = useCallback(async () => {
    try {
      const [statusResponse, connectionsResponse, routingResponse] = await Promise.all([
        getSocks5Status(),
        getSocks5Connections(),
        getSocks5RoutingSnapshot({
          profileId: selectedProfileID,
          keyword: debouncedKeyword,
          status: nodeStatus === 'all' ? '' : nodeStatus,
          sortBy,
          sortOrder,
          page: nodePage + 1,
          pageSize: nodePageSize
        })
      ]);
      setStatus(statusResponse.data);
      setConnections(Array.isArray(connectionsResponse.data) ? connectionsResponse.data : []);
      setRouting(routingResponse.data || { items: [], summary: {}, total: 0, page: 1, pageSize: nodePageSize });
    } catch (error) {
      if (error.response?.status !== 404) showMessage(error.response?.data?.msg || error.message, 'error');
    }
  }, [debouncedKeyword, nodePage, nodePageSize, nodeStatus, selectedProfileID, showMessage, sortBy, sortOrder]);

  useEffect(() => {
    let active = true;
    getSocks5RoutingProfiles()
      .then((response) => {
        if (active) setProfiles(Array.isArray(response.data) ? response.data : []);
      })
      .catch((error) => {
        if (active && error.response?.status !== 404) showMessage(error.response?.data?.msg || error.message, 'error');
      });
    return () => {
      active = false;
    };
  }, [showMessage]);

  useEffect(() => {
    load();
    const timer = setInterval(() => {
      if (document.visibilityState === 'visible') load();
    }, 3000);
    return () => clearInterval(timer);
  }, [load]);

  useEffect(() => {
    const resolvedPage = Math.max(0, Number(routing.page || 1) - 1);
    if (resolvedPage !== nodePage) setNodePage(resolvedPage);
  }, [nodePage, routing.page]);

  useEffect(() => {
    if (profiles.length && !profiles.some((profile) => profile.id === selectedProfileID)) {
      setSelectedProfileID('default');
      setNodePage(0);
    }
  }, [profiles, selectedProfileID]);

  const disconnect = async (id) => {
    try {
      await closeSocks5Connection(id);
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    }
  };

  const probeHealth = async () => {
    setProbing(true);
    try {
      const response = await probeSocks5Health();
      showMessage(
        t(response.data?.started === false ? 'settings.socks5.messages.healthProbeRunning' : 'settings.socks5.messages.healthProbeStarted')
      );
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    } finally {
      setProbing(false);
    }
  };

  const resetRuntimeStats = async () => {
    setResetting(true);
    try {
      await resetSocks5RuntimeStats();
      showMessage(t('settings.socks5.messages.runtimeStatsReset'));
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    } finally {
      setResetting(false);
    }
  };

  const disconnectAll = async () => {
    try {
      await closeAllSocks5Connections();
      await load();
    } catch (error) {
      showMessage(error.message, 'error');
    }
  };

  const changeSort = (field) => {
    setNodePage(0);
    if (sortBy === field) {
      setSortOrder((previous) => (previous === 'asc' ? 'desc' : 'asc'));
      return;
    }
    setSortBy(field);
    setSortOrder('asc');
  };

  const stats = status?.stats || {};
  const health = status?.health || { nodes: [] };
  const routingItems = Array.isArray(routing.items) ? routing.items : [];
  const routingSummary = routing.summary || {};
  const sortableHeader = (field, label, align = 'left') => (
    <TableCell align={align} sortDirection={sortBy === field ? sortOrder : false}>
      <TableSortLabel active={sortBy === field} direction={sortBy === field ? sortOrder : 'asc'} onClick={() => changeSort(field)}>
        {label}
      </TableSortLabel>
    </TableCell>
  );

  return (
    <Card variant="outlined" sx={{ mt: 2 }}>
      <CardHeader
        title={t('settings.socks5.monitor.title')}
        subheader={t('settings.socks5.monitor.subheader')}
        action={
          <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" justifyContent="flex-end">
            <Button onClick={probeHealth} disabled={probing || health.running || !status?.config?.running}>
              {probing || health.running ? t('settings.socks5.monitor.probing') : t('settings.socks5.monitor.probeNow')}
            </Button>
            <Button onClick={resetRuntimeStats} disabled={resetting || !status?.config?.running}>
              {resetting ? t('settings.socks5.monitor.resettingStats') : t('settings.socks5.monitor.resetStats')}
            </Button>
            <Button color="error" onClick={disconnectAll} disabled={!connections.length}>
              {t('settings.socks5.monitor.disconnectAll')}
            </Button>
          </Stack>
        }
      />
      <CardContent>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
          <Chip label={t('settings.socks5.monitor.activeConnections', { count: stats.activeConnections ?? 0 })} />
          <Chip label={t('settings.socks5.monitor.totalConnections', { count: stats.totalConnections ?? 0 })} />
          <Chip label={t('settings.socks5.monitor.successfulConnections', { count: stats.successfulConnections ?? 0 })} color="success" />
          <Chip label={t('settings.socks5.monitor.failedConnections', { count: stats.failedConnections ?? 0 })} color="error" />
          <Chip label={t('settings.socks5.monitor.uploadBytes', { value: formatBytes(stats.uploadBytes ?? 0) })} />
          <Chip label={t('settings.socks5.monitor.downloadBytes', { value: formatBytes(stats.downloadBytes ?? 0) })} />
        </Stack>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" sx={{ mt: 1.5 }}>
          <Chip label={t('settings.socks5.monitor.healthyNodes', { count: health.healthyNodes ?? 0 })} color="success" />
          <Chip label={t('settings.socks5.monitor.unhealthyNodes', { count: health.unhealthyNodes ?? 0 })} color="error" />
          <Chip label={t('settings.socks5.monitor.checkingNodes', { count: health.checkingNodes ?? 0 })} color="warning" />
          <Chip label={t('settings.socks5.monitor.unknownNodes', { count: health.unknownNodes ?? 0 })} />
        </Stack>
        <Box sx={{ mt: 2 }}>
          {connections.length ? (
            connections.map((connection) => (
              <Stack
                key={connection.id}
                direction={{ xs: 'column', md: 'row' }}
                spacing={1}
                alignItems={{ md: 'center' }}
                sx={{ py: 1, borderBottom: 1, borderColor: 'divider' }}
              >
                <Typography sx={{ flex: 1 }}>
                  {connection.target || t('settings.socks5.monitor.connecting')} ·{' '}
                  {connection.nodeName || t('settings.socks5.monitor.waitingNode')}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {connection.profileName || connection.profileId || 'Default'}
                  {connection.account ? ` · ${connection.account}` : ''}
                  {connection.listenerId ? ` · ${connection.listenerId}` : ''}
                </Typography>
                <Typography variant="caption">
                  {connection.clientAddress} · {connection.phase} · ↑ {formatBytes(connection.uploadBytes ?? 0)} · ↓{' '}
                  {formatBytes(connection.downloadBytes ?? 0)} ·{' '}
                  {t('settings.socks5.monitor.duration', { value: formatElapsed(connection.startedAt) })} ·{' '}
                  {t('settings.socks5.monitor.idle', { value: formatElapsed(connection.lastActivity) })}
                </Typography>
                <Button size="small" color="error" onClick={() => disconnect(connection.id)}>
                  {t('settings.socks5.monitor.disconnect')}
                </Button>
              </Stack>
            ))
          ) : (
            <Typography color="text.secondary">{t('settings.socks5.monitor.empty')}</Typography>
          )}
        </Box>

        <Box sx={{ mt: 3 }}>
          <Typography variant="subtitle1">{t('settings.socks5.monitor.nodeRouting')}</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
            {t('settings.socks5.monitor.nodeRoutingHelper')}
          </Typography>
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
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} sx={{ mb: 1.5 }}>
            <FormControl size="small" sx={{ minWidth: 180 }}>
              <InputLabel>{t('settings.socks5.monitor.profileFilter')}</InputLabel>
              <Select
                value={selectedProfileID}
                label={t('settings.socks5.monitor.profileFilter')}
                onChange={(event) => {
                  setSelectedProfileID(event.target.value);
                  setNodePage(0);
                }}
              >
                {profiles
                  .filter((profile) => profile.enabled)
                  .map((profile) => (
                    <MenuItem key={profile.id} value={profile.id}>
                      {profile.name}
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
            <FormControl size="small" sx={{ minWidth: 180 }}>
              <InputLabel>{t('settings.socks5.monitor.nodeStatusFilter')}</InputLabel>
              <Select
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
          {!status?.config?.running && <Alert severity="info">{t('settings.socks5.monitor.routingStopped')}</Alert>}
          {status?.config?.running && routingItems.length === 0 && (
            <Alert severity="info">{t('settings.socks5.monitor.noRoutingData')}</Alert>
          )}
          {routingItems.length > 0 && (
            <TableContainer sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    {sortableHeader('nodeName', t('settings.socks5.monitor.columns.node'))}
                    {sortableHeader('status', t('settings.socks5.monitor.columns.status'))}
                    {sortableHeader('latencyMs', t('settings.socks5.monitor.columns.latency'), 'right')}
                    {sortableHeader('activeConnections', t('settings.socks5.monitor.columns.active'), 'right')}
                    {sortableHeader('successfulConnections', t('settings.socks5.monitor.columns.successful'), 'right')}
                    {sortableHeader('failedConnections', t('settings.socks5.monitor.columns.failed'), 'right')}
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
                      <TableCell>
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
                      </TableCell>
                      <TableCell align="right" sx={{ whiteSpace: 'nowrap' }}>
                        <Typography variant="body2">{node.latencyMs > 0 ? `${node.latencyMs} ms` : '-'}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          {t(`settings.socks5.monitor.latencySource.${node.latencySource || 'fallback'}`)}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">{formatCount(node.activeConnections)}</TableCell>
                      <TableCell align="right">{formatCount(node.successfulConnections)}</TableCell>
                      <TableCell align="right">
                        <Typography variant="body2">{formatCount(node.failedConnections)}</Typography>
                        <Typography variant="caption" color={node.consecutiveFailures ? 'error' : 'text.secondary'}>
                          {t('settings.socks5.monitor.consecutiveFailuresShort', { count: node.consecutiveFailures ?? 0 })}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">{Number(node.smartScore ?? 0).toFixed(1)}</TableCell>
                      <TableCell sx={{ minWidth: 180 }}>
                        <Typography variant="body2">
                          {node.lastSelectedAt ? new Date(node.lastSelectedAt).toLocaleString() : '-'}
                        </Typography>
                        {node.cooldownUntil && (
                          <Typography variant="caption" color="warning.main" display="block">
                            {t('settings.socks5.monitor.cooldownUntil', {
                              value: new Date(node.cooldownUntil).toLocaleString()
                            })}
                          </Typography>
                        )}
                        {node.lastError && (
                          <Typography variant="caption" color="error" display="block" noWrap title={node.lastError} sx={{ maxWidth: 240 }}>
                            {node.lastError}
                          </Typography>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
          <TablePagination
            component="div"
            count={routing.total ?? 0}
            page={nodePage}
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
      </CardContent>
    </Card>
  );
}

export default function Socks5GatewayPage() {
  const { user } = useAuth();
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [profileRevision, setProfileRevision] = useState(0);
  const isAdmin = [user?.role, ...(user?.roles || [])].some((role) => String(role || '').toLowerCase() === 'admin');

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  const handleSnackbarClose = useCallback(() => {
    setSnackbar((previous) => ({ ...previous, open: false }));
  }, []);

  if (!isAdmin) {
    return <Navigate to="/system/settings" replace />;
  }

  return (
    <>
      <Socks5Settings showMessage={showMessage} onChanged={() => setProfileRevision((value) => value + 1)} />
      <RoutingProfiles
        key={`profiles-${profileRevision}`}
        showMessage={showMessage}
        onChanged={() => setProfileRevision((value) => value + 1)}
      />
      <Socks5Accounts
        key={`accounts-${profileRevision}`}
        showMessage={showMessage}
        onChanged={() => setProfileRevision((value) => value + 1)}
      />
      <Socks5Listeners
        key={`listeners-${profileRevision}`}
        showMessage={showMessage}
        onChanged={() => setProfileRevision((value) => value + 1)}
      />
      <Socks5Monitor key={`monitor-${profileRevision}`} showMessage={showMessage} />
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleSnackbarClose}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>
    </>
  );
}
