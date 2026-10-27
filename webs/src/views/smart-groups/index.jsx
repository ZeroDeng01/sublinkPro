import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTaskCompletion } from 'contexts/TaskProgressContext';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Chip,
  CircularProgress,
  Collapse,
  Divider,
  FormControlLabel,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  MenuItem,
  Paper,
  Stack,
  Switch,
  TextField,
  Typography
} from '@mui/material';
import MainCard from 'ui-component/cards/MainCard';
import { getNodeCountries, getNodeGroups } from 'api/nodes';
import { getNodeCheckProfiles } from 'api/nodeCheck';
import {
  checkSmartGroup,
  createSmartGroup,
  deleteSmartGroup,
  getSmartGroupMembers,
  getSmartGroups,
  updateSmartGroup
} from 'api/smartGroups';
import { formatCountry } from 'utils/countryDisplay';
import CloseIcon from '@mui/icons-material/Close';
import CronExpressionGenerator from 'components/CronExpressionGenerator';
import { formatDateTime } from 'i18n/locales';

const emptyForm = {
  name: '',
  countries: '',
  keyword: '',
  sourceGroups: [],
  maxDelay: 0,
  minSpeed: 0,
  maxAgeHours: 72,
  checkProfileId: 0,
  scheduleEnabled: false,
  cronExpr: '0 3 * * *',
  checkAfterAirportUpdate: false
};
const isCountryCode = (value) => typeof value === 'string' && /^[A-Z]{2}$/i.test(value);
const parseCountries = (value) => (value || '').split(',').filter(Boolean);
// Intl's region names provide a full locale-aware country/territory catalog rather than only countries already found on nodes.
const regionCodes = Array.from({ length: 26 * 26 }, (_, index) => String.fromCharCode(65 + Math.floor(index / 26), 65 + (index % 26)));
const nonCountryRegions = new Set(['EU', 'UN', 'EZ', 'QO', 'ZZ']);
const exclusionKeys = ['delayUnusable', 'delayOverLimit', 'delayStale', 'speedUnusable', 'speedBelowMin', 'speedStale'];

export default function SmartGroupsPage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { registerOnComplete, unregisterOnComplete } = useTaskCompletion();
  const [groups, setGroups] = useState([]);
  const [countries, setCountries] = useState([]);
  const [sourceGroups, setSourceGroups] = useState([]);
  const [profiles, setProfiles] = useState([]);
  const [profileId, setProfileId] = useState('');
  const [editing, setEditing] = useState(null);
  const [form, setForm] = useState(emptyForm);
  const [view, setView] = useState(null);
  const memberRequest = useRef(0);
  const closeMembers = () => {
    memberRequest.current += 1;
    setView(null);
  };
  const [members, setMembers] = useState(null);
  const [memberPage, setMemberPage] = useState(1);
  const [loadingMembers, setLoadingMembers] = useState(false);
  const [checkMessage, setCheckMessage] = useState('');
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');
  const displayNames = useMemo(
    () => new Intl.DisplayNames([i18n.resolvedLanguage || i18n.language], { type: 'region' }),
    [i18n.resolvedLanguage, i18n.language]
  );
  const countryOptions = useMemo(
    () =>
      [
        ...new Set([
          ...countries.filter(isCountryCode).map((code) => code.toUpperCase()),
          ...regionCodes.filter((code) => !nonCountryRegions.has(code) && displayNames.of(code) !== code)
        ])
      ].sort(),
    [countries, displayNames]
  );
  const groupOptions = useMemo(
    () => [...new Set([...sourceGroups, ...(form.sourceGroups || [])])].sort(),
    [sourceGroups, form.sourceGroups]
  );
  const countryLabel = (code) => {
    // freeSolo labels can be arbitrary input before selection; Intl rejects non-region values.
    if (!isCountryCode(code)) return code || '';
    const name = displayNames.of(code);
    return name && name !== code ? `${formatCountry(code)} — ${name}` : formatCountry(code);
  };

  const refresh = useCallback(async () => {
    try {
      const response = await getSmartGroups();
      setGroups(response.data || []);
    } catch (error) {
      setMessage(error.message || t('smartGroups.loadFailed'));
    }
  }, [t]);

  useEffect(() => {
    void refresh();
    void getNodeCountries()
      .then((response) => setCountries(response.data || []))
      .catch(() => {});
    void getNodeCheckProfiles()
      .then((response) => {
        const items = response.data || [];
        setProfiles(items);
        setProfileId((current) => current || String(items[0]?.id ?? items[0]?.ID ?? ''));
      })
      .catch(() => {});
    void getNodeGroups()
      .then((response) => setSourceGroups(response.data || []))
      .catch(() => {});
  }, [refresh]);

  const automaticChecks = Boolean(form.scheduleEnabled || form.checkAfterAirportUpdate);
  const selectedAutoProfile = profiles.find((profile) => String(profile.id ?? profile.ID) === String(form.checkProfileId));
  const autoProfileMissing = automaticChecks && !selectedAutoProfile;
  const autoProfileMismatch = Number(form.minSpeed) > 0 && selectedAutoProfile?.mode === 'tcp';
  const cronInvalid = Boolean(form.scheduleEnabled) && (form.cronExpr || '').trim().split(/\s+/).length !== 5;
  const manualProfileMismatch =
    Number(view?.minSpeed) > 0 && profiles.find((profile) => String(profile.id ?? profile.ID) === String(profileId))?.mode === 'tcp';
  const hasAutomaticGroups = groups.some((group) => group.scheduleEnabled || group.checkAfterAirportUpdate || group.checkRunning);
  const formatCheckTime = (value) =>
    value
      ? formatDateTime(value, i18n.resolvedLanguage || i18n.language, {
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit'
        })
      : t('smartGroups.notTriggered');

  const save = async () => {
    setBusy(true);
    try {
      if (editing?.id) await updateSmartGroup(editing.id, form);
      else await createSmartGroup(form);
      setEditing(null);
      await refresh();
    } catch (error) {
      setMessage(error.message || t('smartGroups.saveFailed'));
    } finally {
      setBusy(false);
    }
  };

  const showMembers = useCallback(
    async (group, page = 1) => {
      const request = ++memberRequest.current;
      setView(group);
      setMemberPage(page);
      setLoadingMembers(true);
      try {
        const response = await getSmartGroupMembers(group.id, page);
        if (request !== memberRequest.current) return;
        setMembers(response.data);
        if (response.data?.checkInfo) setView(response.data.checkInfo);
      } catch (error) {
        if (request === memberRequest.current) setCheckMessage(error.message || t('smartGroups.loadFailed'));
      } finally {
        if (request === memberRequest.current) setLoadingMembers(false);
      }
    },
    [t]
  );

  useEffect(() => {
    const onComplete = ({ taskType }) => {
      if (taskType === 'speed_test') {
        void refresh();
        if (view) void showMembers(view, memberPage);
      }
    };
    registerOnComplete(onComplete);
    return () => unregisterOnComplete(onComplete);
  }, [view, memberPage, showMembers, refresh, registerOnComplete, unregisterOnComplete]);

  useEffect(() => {
    if (!hasAutomaticGroups && !view?.checkRunning) return undefined;
    const interval = setInterval(() => {
      void refresh();
      if (view) void showMembers(view, memberPage);
    }, 15000);
    return () => clearInterval(interval);
  }, [hasAutomaticGroups, view, memberPage, refresh, showMembers]);

  const runCheck = async (group, nodeId = 0) => {
    if (!profileId) {
      setCheckMessage(t('smartGroups.chooseProfile'));
      return;
    }
    setBusy(true);
    setCheckMessage('');
    try {
      const response = await checkSmartGroup(group.id, Number(profileId), nodeId);
      setCheckMessage(t('smartGroups.started', { count: response.data?.count || 0 }));
      setView((current) => (current?.id === group.id ? { ...current, checkRunning: true } : current));
      await refresh();
    } catch (error) {
      setCheckMessage(error.message || t('smartGroups.checkFailed'));
    } finally {
      setBusy(false);
    }
  };

  const remove = async (group) => {
    if (!window.confirm(t('smartGroups.confirmDelete', { name: group.name }))) return;
    setBusy(true);
    try {
      await deleteSmartGroup(group.id);
      if (view?.id === group.id) closeMembers();
      await refresh();
    } catch (error) {
      setMessage(error.message || t('smartGroups.deleteFailed'));
    } finally {
      setBusy(false);
    }
  };

  return (
    <MainCard title={t('smartGroups.title')}>
      <Stack spacing={2}>
        {message && (
          <Alert severity="info" onClose={() => setMessage('')}>
            {message}
          </Alert>
        )}
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
          <Button
            variant="contained"
            onClick={() => {
              setForm({ ...emptyForm });
              setEditing({});
            }}
          >
            {t('smartGroups.add')}
          </Button>
          <Button variant="outlined" onClick={() => void refresh()}>
            {t('smartGroups.refresh')}
          </Button>
        </Stack>
        {groups.map((group) => (
          <Paper key={group.id} variant="outlined" sx={{ p: 2 }}>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} alignItems={{ md: 'center' }} justifyContent="space-between">
              <Box>
                <Typography variant="h4">{group.name}</Typography>
                <Typography variant="body2">
                  {parseCountries(group.countries).map(formatCountry).join(' / ')} ·{' '}
                  {t('smartGroups.rule', {
                    delay: group.maxDelay || '∞',
                    speed: group.minSpeed || 0,
                    age: group.maxAgeHours || '∞'
                  })}
                  {group.keyword && ` · ${t('smartGroups.keywordAlternative')}: ${group.keyword}`}
                  {group.sourceGroups?.length > 0 && ` · ${t('smartGroups.sourceGroups')}: ${group.sourceGroups.join(' / ')}`}
                </Typography>
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" sx={{ mt: 1 }}>
                  {group.scheduleEnabled && (
                    <Chip size="small" variant="outlined" label={t('smartGroups.scheduleSummary', { cron: group.cronExpr })} />
                  )}
                  {group.checkAfterAirportUpdate && <Chip size="small" variant="outlined" label={t('smartGroups.afterAirportUpdate')} />}
                  {group.checkRunning && <Chip size="small" color="info" label={t('smartGroups.checkRunning')} />}
                </Stack>
                {(group.checkProfileId > 0 || group.lastCheckAt) && (
                  <Typography variant="caption" color="text.secondary" display="block" sx={{ mt: 0.5 }}>
                    {t('smartGroups.lastTriggered')}: {formatCheckTime(group.lastCheckAt)} · {t('smartGroups.nextCheck')}:{' '}
                    {group.nextCheckAt ? formatCheckTime(group.nextCheckAt) : t('smartGroups.notScheduled')}
                  </Typography>
                )}
              </Box>
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                <Button
                  disabled={busy}
                  onClick={() => {
                    setMembers(null);
                    setCheckMessage('');
                    if (group.checkProfileId) setProfileId(String(group.checkProfileId));
                    void showMembers(group);
                  }}
                >
                  {t('smartGroups.members')}
                </Button>
                <Button disabled={busy} onClick={() => navigate(`/subscription/subs?smartGroupId=${group.id}`)}>
                  {t('smartGroups.createSubscription')}
                </Button>
                <Button
                  disabled={busy}
                  onClick={() => {
                    setForm({
                      ...emptyForm,
                      ...group,
                      sourceGroups: group.sourceGroups || [],
                      cronExpr: group.cronExpr || emptyForm.cronExpr
                    });
                    setEditing(group);
                  }}
                >
                  {t('smartGroups.edit')}
                </Button>
                <Button color="error" disabled={busy} onClick={() => void remove(group)}>
                  {t('smartGroups.delete')}
                </Button>
              </Stack>
            </Stack>
          </Paper>
        ))}
      </Stack>
      <Dialog
        open={view !== null}
        onClose={closeMembers}
        fullWidth
        maxWidth="lg"
        fullScreen={isMobile}
        PaperProps={{ sx: { bgcolor: 'background.paper', maxHeight: isMobile ? '100dvh' : '90dvh' } }}
      >
        <DialogTitle component="div" sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
          <Box sx={{ minWidth: 0 }}>
            <Typography component="h2" variant="h3" noWrap>
              {view?.name}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {t('smartGroups.membersTitle')}
            </Typography>
          </Box>
          <IconButton aria-label={t('smartGroups.close')} onClick={closeMembers} size="small">
            <CloseIcon fontSize="small" />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers sx={{ px: { xs: 2, sm: 3 }, bgcolor: 'background.default' }}>
          <Stack spacing={2} sx={{ mt: 1.5 }}>
            <Paper variant="outlined" sx={{ p: { xs: 1.5, sm: 2 }, borderRadius: 2, bgcolor: 'background.paper' }}>
              <Stack spacing={1.25}>
                {members ? (
                  <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                    <Chip variant="outlined" label={t('smartGroups.candidatesCount', { count: members.candidateCount })} />
                    <Chip color="success" variant="outlined" label={t('smartGroups.healthyCount', { count: members.count })} />
                    {members.statusCounts &&
                      exclusionKeys
                        .filter((key) => members.statusCounts[key] > 0)
                        .map((key) => (
                          <Chip
                            key={key}
                            size="small"
                            variant="outlined"
                            label={t(`smartGroups.exclusions.${key}`, { count: members.statusCounts[key] })}
                          />
                        ))}
                  </Stack>
                ) : (
                  <Typography variant="body2" color="text.secondary">
                    {t('smartGroups.loadingCandidates')}
                  </Typography>
                )}
              </Stack>
            </Paper>
            {checkMessage && (
              <Alert severity="info" onClose={() => setCheckMessage('')}>
                {checkMessage}
              </Alert>
            )}
            <Paper variant="outlined" sx={{ p: 1.5, borderRadius: 2, bgcolor: 'background.paper' }}>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ sm: 'center' }}>
                <TextField
                  select
                  size="small"
                  label={t('smartGroups.profile')}
                  value={profileId}
                  onChange={(event) => setProfileId(event.target.value)}
                  sx={{ minWidth: { sm: 220 }, width: { xs: '100%', sm: 'auto' } }}
                >
                  {profiles.map((profile) => (
                    <MenuItem key={profile.id ?? profile.ID} value={profile.id ?? profile.ID}>
                      {profile.name ?? profile.Name}
                    </MenuItem>
                  ))}
                </TextField>
                <Stack direction="row" spacing={1} sx={{ width: { xs: '100%', sm: 'auto' } }}>
                  <Button
                    variant="contained"
                    disabled={busy || view?.checkRunning || manualProfileMismatch || !members?.candidateCount}
                    onClick={() => void runCheck(view)}
                    sx={{ flex: { xs: 1, sm: 'none' } }}
                  >
                    {t('smartGroups.check')}
                  </Button>
                  <Button variant="outlined" disabled={loadingMembers} onClick={() => void showMembers(view, memberPage)}>
                    {t('smartGroups.refresh')}
                  </Button>
                </Stack>
              </Stack>
            </Paper>
            {view?.checkRunning && <Alert severity="info">{t('smartGroups.checkRunningHint')}</Alert>}
            {manualProfileMismatch && <Alert severity="warning">{t('smartGroups.tcpWarning')}</Alert>}
            {loadingMembers && <CircularProgress size={24} />}
            {!loadingMembers && members?.candidateCount === 0 && <Alert severity="info">{t('smartGroups.noCandidatesHint')}</Alert>}
            {members?.candidateCount > 0 && (
              <Stack spacing={1}>
                <Typography variant="subtitle1" fontWeight={600}>
                  {t('smartGroups.candidateList')}
                </Typography>
                {(members.candidateNodes || []).map((node) => (
                  <Paper
                    key={node.id}
                    variant="outlined"
                    sx={{
                      px: { xs: 1.5, sm: 2 },
                      py: 1.25,
                      borderRadius: 2,
                      bgcolor: 'background.paper',
                      '&:hover': { bgcolor: 'action.hover' }
                    }}
                  >
                    <Box
                      sx={{
                        display: 'grid',
                        alignItems: 'center',
                        gap: { xs: 1, md: 2 },
                        gridTemplateColumns: {
                          xs: 'minmax(0, 1fr) auto',
                          md: 'minmax(0, 2.4fr) minmax(0, 1.5fr) minmax(100px, 1fr) auto'
                        }
                      }}
                    >
                      <Box sx={{ minWidth: 0, gridColumn: { xs: '1 / -1', md: 'auto' } }}>
                        <Typography variant="subtitle2" fontWeight={600} sx={{ overflowWrap: 'anywhere' }}>
                          {node.name}
                        </Typography>
                        <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap alignItems="center" sx={{ mt: 0.5 }}>
                          <Typography variant="caption" color="text.secondary">
                            #{node.id}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {formatCountry(node.country) || '—'}
                          </Typography>
                          <Chip
                            size="small"
                            variant="outlined"
                            color={node.includedBySmartGroup ? 'success' : 'default'}
                            label={t(node.includedBySmartGroup ? 'smartGroups.eligibleShort' : 'smartGroups.excludedShort')}
                          />
                        </Stack>
                        {(node.reason || node.matchSource !== 'landingCountry' || node.countrySource === 'name') && (
                          <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 0.5 }}>
                            {[
                              node.reason && t(`smartGroups.reasons.${node.reason}`),
                              node.matchSource &&
                                node.matchSource !== 'landingCountry' &&
                                t(`smartGroups.matchSources.${node.matchSource}`),
                              node.countrySource === 'name' && t('smartGroups.nameInferred')
                            ]
                              .filter(Boolean)
                              .join(' · ')}
                          </Typography>
                        )}
                      </Box>
                      <Box sx={{ minWidth: 0, gridColumn: { xs: '1 / -1', md: 'auto' } }}>
                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                          {t('smartGroups.sourceGroup')}:{' '}
                          <Box component="span" sx={{ color: 'text.primary' }}>
                            {node.group || '—'}
                          </Box>
                        </Typography>
                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', overflowWrap: 'anywhere' }}>
                          {t('smartGroups.sourceAirport')}:{' '}
                          <Box component="span" sx={{ color: 'text.primary' }}>
                            {node.airportName ||
                              (node.airportId > 0 ? t('smartGroups.deletedAirport', { id: node.airportId }) : t('smartGroups.manualNode'))}
                          </Box>
                        </Typography>
                      </Box>
                      <Box sx={{ minWidth: 0 }}>
                        <Typography variant="body2" noWrap>
                          {t('smartGroups.delay')}: {node.delay > 0 ? `${node.delay} ms` : '—'}
                        </Typography>
                        <Typography variant="body2" noWrap>
                          {t('smartGroups.speed')}: {node.speed > 0 ? `${node.speed} MB/s` : '—'}
                        </Typography>
                      </Box>
                      <Button
                        size="small"
                        variant="outlined"
                        disabled={busy || view?.checkRunning || manualProfileMismatch || !profileId}
                        onClick={() => void runCheck(view, node.id)}
                        sx={{ whiteSpace: 'nowrap' }}
                      >
                        {t('smartGroups.checkOne')}
                      </Button>
                    </Box>
                  </Paper>
                ))}
              </Stack>
            )}
          </Stack>
        </DialogContent>
        <DialogActions
          sx={{
            px: { xs: 2, sm: 3 },
            py: 1.25,
            borderTop: '1px solid',
            borderColor: 'divider',
            justifyContent: 'space-between'
          }}
        >
          <Typography variant="caption" color="text.secondary">
            {members
              ? t('smartGroups.page', {
                  page: memberPage,
                  total: Math.max(1, Math.ceil(members.candidateCount / (members.pageSize || 30)))
                })
              : '—'}
          </Typography>
          <Stack direction="row" spacing={1}>
            <Button size="small" disabled={loadingMembers || memberPage <= 1} onClick={() => void showMembers(view, memberPage - 1)}>
              {t('smartGroups.previous')}
            </Button>
            <Button
              size="small"
              disabled={loadingMembers || !members || memberPage * (members.pageSize || 30) >= members.candidateCount}
              onClick={() => void showMembers(view, memberPage + 1)}
            >
              {t('smartGroups.next')}
            </Button>
          </Stack>
        </DialogActions>
      </Dialog>
      <Dialog open={editing !== null} onClose={() => setEditing(null)} fullWidth maxWidth="sm">
        <DialogTitle>{editing?.id ? t('smartGroups.edit') : t('smartGroups.add')}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label={t('smartGroups.name')}
              required
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
            />
            <Autocomplete
              multiple
              freeSolo
              options={countryOptions}
              value={parseCountries(form.countries)}
              onChange={(_event, values) =>
                setForm({
                  ...form,
                  countries: values
                    .map((value) => value.trim().toUpperCase())
                    .filter(isCountryCode)
                    .join(',')
                })
              }
              getOptionLabel={countryLabel}
              renderInput={(params) => (
                <TextField {...params} required label={t('smartGroups.country')} helperText={t('smartGroups.countryHint')} />
              )}
            />
            <TextField
              label={t('smartGroups.keyword')}
              value={form.keyword || ''}
              onChange={(event) => setForm({ ...form, keyword: event.target.value })}
              helperText={t('smartGroups.keywordHint')}
            />
            <Autocomplete
              multiple
              options={groupOptions}
              value={form.sourceGroups || []}
              onChange={(_event, values) => setForm({ ...form, sourceGroups: values })}
              renderInput={(params) => (
                <TextField {...params} label={t('smartGroups.sourceGroups')} helperText={t('smartGroups.sourceGroupsHint')} />
              )}
            />
            <TextField
              label={t('smartGroups.maxDelay')}
              type="number"
              value={form.maxDelay}
              onChange={(event) => setForm({ ...form, maxDelay: Number(event.target.value) })}
              helperText={t('smartGroups.zeroUnlimited')}
            />
            <TextField
              label={t('smartGroups.minSpeed')}
              type="number"
              value={form.minSpeed}
              onChange={(event) => setForm({ ...form, minSpeed: Number(event.target.value) })}
              helperText={t('smartGroups.speedHint')}
            />
            <TextField
              label={t('smartGroups.maxAge')}
              type="number"
              value={form.maxAgeHours}
              onChange={(event) => setForm({ ...form, maxAgeHours: Number(event.target.value) })}
              helperText={t('smartGroups.zeroUnlimited')}
            />
            <Divider />
            <Typography variant="h5">{t('smartGroups.automaticChecks')}</Typography>
            <TextField
              select
              fullWidth
              label={t('smartGroups.defaultProfile')}
              value={form.checkProfileId || 0}
              onChange={(event) => setForm({ ...form, checkProfileId: Number(event.target.value) })}
              error={autoProfileMissing || autoProfileMismatch}
            >
              <MenuItem value={0}>{t('smartGroups.noDefaultProfile')}</MenuItem>
              {profiles.map((profile) => (
                <MenuItem key={profile.id ?? profile.ID} value={profile.id ?? profile.ID}>
                  {profile.name ?? profile.Name} · {profile.mode === 'tcp' ? 'TCP' : 'mihomo'}
                </MenuItem>
              ))}
            </TextField>
            {autoProfileMismatch && <Alert severity="warning">{t('smartGroups.tcpWarning')}</Alert>}
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.scheduleEnabled)}
                  onChange={(event) =>
                    setForm({
                      ...form,
                      scheduleEnabled: event.target.checked,
                      cronExpr: form.cronExpr || emptyForm.cronExpr
                    })
                  }
                />
              }
              label={t('smartGroups.scheduleEnabled')}
            />
            <Collapse in={Boolean(form.scheduleEnabled)}>
              <Stack spacing={1.5}>
                <CronExpressionGenerator
                  value={form.cronExpr || ''}
                  onChange={(value) => setForm({ ...form, cronExpr: value })}
                  label={t('smartGroups.cronExpr')}
                  error={cronInvalid}
                />
                <Typography variant="caption" color="text.secondary">
                  {t('smartGroups.scheduleTimezoneHint', {
                    timezone: form.checkTimezone || groups[0]?.checkTimezone || t('smartGroups.serverTimezone')
                  })}
                </Typography>
                {cronInvalid && <Alert severity="warning">{t('smartGroups.invalidCron')}</Alert>}
              </Stack>
            </Collapse>
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(form.checkAfterAirportUpdate)}
                  onChange={(event) => setForm({ ...form, checkAfterAirportUpdate: event.target.checked })}
                />
              }
              label={t('smartGroups.afterAirportUpdate')}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditing(null)}>{t('smartGroups.cancel')}</Button>
          <Button
            disabled={busy || cronInvalid || autoProfileMissing || autoProfileMismatch}
            variant="contained"
            onClick={() => void save()}
          >
            {t('smartGroups.save')}
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
