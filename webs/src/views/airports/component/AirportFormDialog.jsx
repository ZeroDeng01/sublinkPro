import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { useTheme } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Link from '@mui/material/Link';
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import InfoIcon from '@mui/icons-material/Info';

import SearchableNodeSelect from 'components/SearchableNodeSelect';
import CronExpressionGenerator from 'components/CronExpressionGenerator';
import LogoPicker from 'components/LogoPicker';
import NodeNameFilter from 'components/NodeNameFilter';
import NodeNamePreprocessor from 'components/NodeNamePreprocessor';
import NodeProtocolFilter from 'components/NodeProtocolFilter';
import NodeNameUniquifyConfig from 'components/NodeNameUniquifyConfig';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import AirportDeduplicationConfig from './AirportDeduplicationConfig';
import AirportDialogSection from './AirportDialogSection';

import { USER_AGENT_OPTIONS } from '../utils';

const createEmptyRequestHeader = () => ({ key: '', value: '' });

const getRequestHeaderRowError = (requestHeader, t) => {
  const key = `${requestHeader?.key ?? ''}`.trim();
  const value = `${requestHeader?.value ?? ''}`.trim();

  if (!key && !value) {
    return '';
  }

  if (!key && value) {
    return t('airports.form.requestHeaders.errors.missingKey');
  }

  if (key.toLowerCase() === 'user-agent') {
    return t('airports.form.requestHeaders.errors.userAgentDedicated');
  }

  return '';
};

export default function AirportFormDialog({
  open,
  isEdit,
  airportForm,
  setAirportForm,
  groupOptions,
  proxyNodeOptions,
  loadingProxyNodes,
  protocolOptions,
  nodeCheckProfiles,
  onClose,
  onSubmit,
  onFetchProxyNodes
}) {
  const theme = useTheme();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);

  const controlRowSx = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 2,
    px: 1.5,
    py: 1.25,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.paper, 0.2) : withAlpha(palette.background.paper, 0.92),
    border: '1px solid',
    borderColor: panelBorder
  };

  const requestHeaders = Array.isArray(airportForm.requestHeaders) ? airportForm.requestHeaders : [];
  const hasRequestHeaderRows = requestHeaders.length > 0;
  const updateAfterDetectProfileId = airportForm.updateAfterDetectProfileId || '';
  const hasUpdateAfterDetectProfiles = Array.isArray(nodeCheckProfiles) && nodeCheckProfiles.length > 0;

  const requestHeaderPanelSx = {
    p: 1.5,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.default, 0.88) : withAlpha(palette.background.default, 0.56),
    border: '1px solid',
    borderColor: panelBorder
  };

  const requestHeaderRowSx = {
    p: 1.25,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.paper, 0.22) : withAlpha(palette.background.paper, 0.96),
    border: '1px solid',
    borderColor: panelBorder
  };

  const addRequestHeaderButtonSx = {
    minWidth: 0,
    alignSelf: { xs: 'flex-start', sm: 'center' },
    px: 1,
    py: 0.5,
    borderRadius: 1.5,
    fontSize: '0.8125rem',
    fontWeight: 600,
    lineHeight: 1.2,
    color: isDark ? palette.primary.light : palette.primary.main,
    bgcolor: withAlpha(palette.primary.main, isDark ? 0.08 : 0.04),
    borderColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.16),
    whiteSpace: 'nowrap',
    flexShrink: 0,
    '& .MuiButton-startIcon': {
      mr: 0.5,
      ml: 0,
      '& > *:nth-of-type(1)': {
        fontSize: '1rem'
      }
    },
    '&:hover': {
      borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.22),
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.08)
    }
  };

  const updateRequestHeaders = (updater) => {
    const nextRequestHeaders = typeof updater === 'function' ? updater(requestHeaders) : updater;
    setAirportForm({ ...airportForm, requestHeaders: nextRequestHeaders });
  };

  const handleRequestHeaderChange = (index, field, value) => {
    updateRequestHeaders((currentHeaders) =>
      currentHeaders.map((header, headerIndex) => (headerIndex === index ? { ...header, [field]: value } : header))
    );
  };

  const handleAddRequestHeader = () => {
    updateRequestHeaders((currentHeaders) => [...currentHeaders, createEmptyRequestHeader()]);
  };

  const handleRemoveRequestHeader = (index) => {
    updateRequestHeaders((currentHeaders) => currentHeaders.filter((_, headerIndex) => headerIndex !== index));
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          maxHeight: '90vh',
          borderRadius: 2.5,
          border: '1px solid',
          borderColor: panelBorder,
          bgcolor: dialogSurface,
          backgroundImage: dialogSurfaceGradient
        }
      }}
    >
      <DialogTitle
        sx={{
          pb: 1.5,
          color: primaryText,
          bgcolor: mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: panelBorder
        }}
      >
        {isEdit ? t('airports.form.title.edit') : t('airports.form.title.add')}
      </DialogTitle>
      <DialogContent dividers sx={{ pt: 2.5, pb: 2, bgcolor: 'transparent', borderColor: panelBorder }}>
        <Stack spacing={2.5}>
          <AirportDialogSection
            title={t('airports.form.sections.basic')}
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <Stack spacing={2}>
              <TextField
                fullWidth
                size="small"
                label={t('airports.form.fields.name')}
                value={airportForm.name}
                helperText={t('airports.form.helpers.name')}
                onChange={(e) => setAirportForm({ ...airportForm, name: e.target.value })}
              />
              <Box>
                <Typography variant="body2" sx={{ mb: 1, color: secondaryText }}>
                  {t('airports.form.fields.logoOptional')}
                </Typography>
                <LogoPicker
                  value={airportForm.logo || ''}
                  onChange={(logo) => setAirportForm({ ...airportForm, logo })}
                  name={airportForm.name}
                />
              </Box>
              <TextField
                fullWidth
                size="small"
                label={t('airports.form.fields.subscriptionUrl')}
                value={airportForm.url}
                helperText={t('airports.form.helpers.subscriptionUrl')}
                onChange={(e) => setAirportForm({ ...airportForm, url: e.target.value })}
              />
              <Autocomplete
                freeSolo
                size="small"
                options={groupOptions}
                value={airportForm.group}
                onChange={(_, newValue) => setAirportForm({ ...airportForm, group: newValue || '' })}
                onInputChange={(_, newValue) => setAirportForm({ ...airportForm, group: newValue || '' })}
                renderInput={(params) => (
                  <TextField {...params} label={t('airports.form.fields.nodeGroup')} helperText={t('airports.form.helpers.nodeGroup')} />
                )}
              />
              <TextField
                fullWidth
                size="small"
                label={t('airports.form.fields.remark')}
                value={airportForm.remark}
                placeholder={t('airports.form.placeholders.remark')}
                helperText={t('airports.form.helpers.remark')}
                multiline
                minRows={2}
                maxRows={4}
                onChange={(e) => setAirportForm({ ...airportForm, remark: e.target.value })}
              />
            </Stack>
          </AirportDialogSection>

          <AirportDialogSection
            title={t('airports.form.sections.schedule')}
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <Stack spacing={2}>
              <Box sx={controlRowSx}>
                <Box>
                  <Typography variant="body2" sx={{ color: primaryText }}>
                    {t('airports.form.schedule.enable')}
                  </Typography>
                  <Typography variant="caption" sx={{ color: secondaryText }}>
                    {t('airports.form.schedule.disableDescription')}
                  </Typography>
                </Box>
                <Switch checked={airportForm.enabled} onChange={(e) => setAirportForm({ ...airportForm, enabled: e.target.checked })} />
              </Box>
              <Collapse in={airportForm.enabled}>
                <CronExpressionGenerator
                  value={airportForm.cronExpr}
                  onChange={(value) => setAirportForm({ ...airportForm, cronExpr: value })}
                  label=""
                />
              </Collapse>
            </Stack>
          </AirportDialogSection>

          <AirportDialogSection
            title={t('airports.form.sections.request')}
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <Stack spacing={2}>
              <Autocomplete
                freeSolo
                size="small"
                options={USER_AGENT_OPTIONS}
                getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
                value={airportForm.userAgent}
                onChange={(_, newValue) => {
                  const value = typeof newValue === 'string' ? newValue : (newValue?.value ?? '');
                  setAirportForm({ ...airportForm, userAgent: value });
                }}
                onInputChange={(_, newValue) => setAirportForm({ ...airportForm, userAgent: newValue ?? '' })}
                renderOption={(props, option) => (
                  <Box component="li" {...props} key={option.value}>
                    <Box>
                      <Typography variant="body2" sx={{ color: primaryText }}>
                        {option.labelKey ? t(option.labelKey) : option.label}
                      </Typography>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        {option.value}
                      </Typography>
                    </Box>
                  </Box>
                )}
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label="User-Agent"
                    placeholder={t('airports.form.placeholders.userAgent')}
                    helperText={t('airports.form.helpers.userAgent')}
                  />
                )}
              />

              <Box sx={requestHeaderPanelSx}>
                <Stack spacing={1.5}>
                  <Stack
                    direction={{ xs: 'column', sm: 'row' }}
                    spacing={1}
                    justifyContent="space-between"
                    alignItems={{ xs: 'flex-start', sm: 'flex-start' }}
                  >
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Typography variant="body2" sx={{ color: primaryText, fontWeight: 500 }}>
                        {t('airports.form.requestHeaders.title')}
                      </Typography>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        {t('airports.form.requestHeaders.description')}
                      </Typography>
                    </Box>
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<AddIcon />}
                      onClick={handleAddRequestHeader}
                      sx={addRequestHeaderButtonSx}
                    >
                      {t('airports.form.requestHeaders.add')}
                    </Button>
                  </Stack>

                  {hasRequestHeaderRows ? (
                    <Stack spacing={1}>
                      {requestHeaders.map((header, index) => {
                        const rowError = getRequestHeaderRowError(header, t);

                        return (
                          <Box key={`request-header-${index}`} sx={requestHeaderRowSx}>
                            <Stack spacing={1}>
                              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ xs: 'stretch', sm: 'flex-start' }}>
                                <TextField
                                  fullWidth
                                  size="small"
                                  label={t('airports.form.requestHeaders.key')}
                                  placeholder={t('airports.form.requestHeaders.keyPlaceholder')}
                                  value={header.key}
                                  error={Boolean(rowError)}
                                  onChange={(e) => handleRequestHeaderChange(index, 'key', e.target.value)}
                                />
                                <TextField
                                  fullWidth
                                  size="small"
                                  label={t('airports.form.requestHeaders.value')}
                                  placeholder={t('airports.form.requestHeaders.valuePlaceholder')}
                                  value={header.value}
                                  onChange={(e) => handleRequestHeaderChange(index, 'value', e.target.value)}
                                />
                                <IconButton
                                  aria-label={t('airports.form.requestHeaders.deleteRow', { index: index + 1 })}
                                  color="error"
                                  onClick={() => handleRemoveRequestHeader(index)}
                                  sx={{
                                    alignSelf: { xs: 'flex-end', sm: 'center' },
                                    border: '1px solid',
                                    borderColor: withAlpha(palette.error.main, isDark ? 0.32 : 0.22),
                                    bgcolor: withAlpha(palette.error.main, isDark ? 0.12 : 0.04),
                                    '&:hover': {
                                      bgcolor: withAlpha(palette.error.main, isDark ? 0.18 : 0.08)
                                    }
                                  }}
                                >
                                  <DeleteIcon fontSize="small" />
                                </IconButton>
                              </Stack>
                              {rowError ? (
                                <Typography variant="caption" color="error">
                                  {rowError}
                                </Typography>
                              ) : (
                                <Typography variant="caption" sx={{ color: secondaryText }}>
                                  {t('airports.form.requestHeaders.emptyRowIgnored')}
                                </Typography>
                              )}
                            </Stack>
                          </Box>
                        );
                      })}
                    </Stack>
                  ) : (
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      {t('airports.form.requestHeaders.empty')}
                    </Typography>
                  )}
                </Stack>
              </Box>

              <Box>
                <Box sx={{ ...controlRowSx, mb: airportForm.downloadWithProxy ? 1.5 : 0 }}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      {t('airports.form.proxyDownload.enable')}
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      {t('airports.form.proxyDownload.description')}
                    </Typography>
                  </Box>
                  <Switch
                    checked={airportForm.downloadWithProxy}
                    onChange={(e) => {
                      const checked = e.target.checked;
                      setAirportForm({ ...airportForm, downloadWithProxy: checked });
                      if (checked) {
                        onFetchProxyNodes();
                      }
                    }}
                  />
                </Box>
                <Collapse in={airportForm.downloadWithProxy}>
                  <SearchableNodeSelect
                    nodes={proxyNodeOptions}
                    loading={loadingProxyNodes}
                    value={
                      proxyNodeOptions.find((n) => n.Link === airportForm.proxyLink) ||
                      (airportForm.proxyLink ? { Link: airportForm.proxyLink, Name: '', ID: 0 } : null)
                    }
                    onChange={(newValue) =>
                      setAirportForm({ ...airportForm, proxyLink: typeof newValue === 'string' ? newValue : newValue?.Link || '' })
                    }
                    displayField="Name"
                    valueField="Link"
                    label={t('airports.form.proxyDownload.proxyNode')}
                    placeholder={t('airports.form.proxyDownload.placeholder')}
                    helperText={t('airports.form.proxyDownload.helper')}
                    freeSolo={true}
                    limit={50}
                    size="small"
                  />
                </Collapse>
              </Box>
            </Stack>
          </AirportDialogSection>

          <AirportDialogSection
            title={t('airports.form.sections.advanced')}
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <Stack spacing={1}>
              <Box>
                <Box sx={controlRowSx}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      {t('airports.form.advanced.fetchUsage')}
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      {t('airports.form.advanced.fetchUsageDescription')}
                    </Typography>
                  </Box>
                  <Switch
                    checked={airportForm.fetchUsageInfo || false}
                    onChange={(e) => setAirportForm({ ...airportForm, fetchUsageInfo: e.target.checked })}
                  />
                </Box>
                <Collapse in={airportForm.fetchUsageInfo}>
                  <Alert severity="info" sx={{ mt: 1 }} icon={false}>
                    <Typography variant="caption">{t('airports.form.advanced.fetchUsageAlert')}</Typography>
                  </Alert>
                </Collapse>
              </Box>

              <Divider sx={{ my: 0.5, borderColor: panelBorder }} />

              <Box>
                <Box sx={controlRowSx}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      {t('airports.form.advanced.skipTls')}
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      {t('airports.form.advanced.skipTlsDescription')}
                    </Typography>
                  </Box>
                  <Switch
                    checked={airportForm.skipTLSVerify || false}
                    onChange={(e) => setAirportForm({ ...airportForm, skipTLSVerify: e.target.checked })}
                  />
                </Box>
                <Collapse in={airportForm.skipTLSVerify}>
                  <Alert severity="warning" sx={{ mt: 1 }} icon={false}>
                    <Typography variant="caption">{t('airports.form.advanced.skipTlsAlert')}</Typography>
                  </Alert>
                </Collapse>
              </Box>

              <Divider sx={{ my: 0.5, borderColor: panelBorder }} />

              <Box>
                <Box sx={controlRowSx}>
                  <Box>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      {t('airports.form.advanced.updateAfterDetect')}
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      {t('airports.form.advanced.updateAfterDetectDescription')}
                    </Typography>
                  </Box>
                  <Switch
                    checked={airportForm.updateAfterDetect || false}
                    onChange={(e) => setAirportForm({ ...airportForm, updateAfterDetect: e.target.checked })}
                  />
                </Box>
                <Collapse in={airportForm.updateAfterDetect}>
                  <Stack spacing={1} sx={{ mt: 1 }}>
                    <TextField
                      select
                      fullWidth
                      size="small"
                      label={t('airports.form.advanced.detectProfile')}
                      value={updateAfterDetectProfileId}
                      onChange={(e) =>
                        setAirportForm({
                          ...airportForm,
                          updateAfterDetectProfileId: Number(e.target.value) || 0
                        })
                      }
                      disabled={!hasUpdateAfterDetectProfiles}
                      helperText={
                        hasUpdateAfterDetectProfiles
                          ? t('airports.form.advanced.detectProfileHelper')
                          : t('airports.form.advanced.noDetectProfiles')
                      }
                    >
                      <MenuItem value="">{t('airports.form.advanced.selectProfile')}</MenuItem>
                      {nodeCheckProfiles.map((profile) => (
                        <MenuItem key={profile.id} value={profile.id}>
                          {profile.name}
                        </MenuItem>
                      ))}
                    </TextField>
                    <Box sx={controlRowSx}>
                      <Box>
                        <Typography variant="body2" sx={{ color: primaryText }}>
                          {t('airports.form.advanced.changedOnly')}
                        </Typography>
                        <Typography variant="caption" sx={{ color: secondaryText }}>
                          {t('airports.form.advanced.changedOnlyDescription')}
                        </Typography>
                      </Box>
                      <Switch
                        checked={airportForm.updateAfterDetectChangedOnly || false}
                        onChange={(e) => setAirportForm({ ...airportForm, updateAfterDetectChangedOnly: e.target.checked })}
                      />
                    </Box>
                  </Stack>
                </Collapse>
              </Box>
            </Stack>
          </AirportDialogSection>

          <AirportDialogSection
            title={t('airports.form.sections.nodeProcessing')}
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <Stack spacing={2}>
              <Alert severity="info" icon={<InfoIcon />}>
                <Typography variant="caption">
                  {t('airports.form.nodeProcessing.alert')}{' '}
                  <Link
                    component="button"
                    variant="caption"
                    onClick={(e) => {
                      e.preventDefault();
                      navigate('/system/settings', { state: { targetTab: 'globalNodeProcessing' } });
                    }}
                    sx={{ cursor: 'pointer', textDecoration: 'underline' }}
                  >
                    {t('airports.form.nodeProcessing.globalRulesLink')}
                  </Link>
                </Typography>
              </Alert>
              <NodeNameFilter
                whitelistValue={airportForm.nodeNameWhitelist || ''}
                blacklistValue={airportForm.nodeNameBlacklist || ''}
                onWhitelistChange={(rules) => setAirportForm({ ...airportForm, nodeNameWhitelist: rules })}
                onBlacklistChange={(rules) => setAirportForm({ ...airportForm, nodeNameBlacklist: rules })}
              />
              <NodeProtocolFilter
                protocolOptions={protocolOptions}
                whitelistValue={airportForm.protocolWhitelist || ''}
                blacklistValue={airportForm.protocolBlacklist || ''}
                onWhitelistChange={(protocols) => setAirportForm({ ...airportForm, protocolWhitelist: protocols })}
                onBlacklistChange={(protocols) => setAirportForm({ ...airportForm, protocolBlacklist: protocols })}
              />
              <AirportDeduplicationConfig
                value={airportForm.deduplicationRule || ''}
                onChange={(rule) => setAirportForm({ ...airportForm, deduplicationRule: rule })}
              />
              <NodeNamePreprocessor
                value={airportForm.nodeNamePreprocess || ''}
                onChange={(rules) => setAirportForm({ ...airportForm, nodeNamePreprocess: rules })}
              />
              <NodeNameUniquifyConfig
                enabled={airportForm.nodeNameUniquify || false}
                prefix={airportForm.nodeNamePrefix || ''}
                intraUniquify={airportForm.nodeNameIntraUniquify || false}
                airportId={airportForm.id || 0}
                onChange={({ enabled, prefix, intraUniquify }) =>
                  setAirportForm({
                    ...airportForm,
                    nodeNameUniquify: enabled,
                    nodeNamePrefix: prefix,
                    nodeNameIntraUniquify: intraUniquify
                  })
                }
              />
            </Stack>
          </AirportDialogSection>
        </Stack>
      </DialogContent>
      <DialogActions
        sx={{
          px: 3,
          py: 2,
          bgcolor: mutedPanelSurface,
          borderTop: '1px solid',
          borderColor: panelBorder
        }}
      >
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={onSubmit}>
          {t('common.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

AirportFormDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  isEdit: PropTypes.bool.isRequired,
  airportForm: PropTypes.shape({
    id: PropTypes.number,
    name: PropTypes.string,
    url: PropTypes.string,
    cronExpr: PropTypes.string,
    enabled: PropTypes.bool,
    group: PropTypes.string,
    downloadWithProxy: PropTypes.bool,
    proxyLink: PropTypes.string,
    userAgent: PropTypes.string,
    requestHeaders: PropTypes.arrayOf(
      PropTypes.shape({
        key: PropTypes.string,
        value: PropTypes.string
      })
    ),
    fetchUsageInfo: PropTypes.bool,
    skipTLSVerify: PropTypes.bool,
    updateAfterDetect: PropTypes.bool,
    updateAfterDetectProfileId: PropTypes.number,
    updateAfterDetectChangedOnly: PropTypes.bool,
    remark: PropTypes.string,
    logo: PropTypes.string,
    nodeNameWhitelist: PropTypes.string,
    nodeNameBlacklist: PropTypes.string,
    protocolWhitelist: PropTypes.string,
    protocolBlacklist: PropTypes.string,
    nodeNamePreprocess: PropTypes.string,
    deduplicationRule: PropTypes.string,
    nodeNameUniquify: PropTypes.bool,
    nodeNamePrefix: PropTypes.string,
    nodeNameIntraUniquify: PropTypes.bool
  }).isRequired,
  setAirportForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  protocolOptions: PropTypes.array,
  nodeCheckProfiles: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
