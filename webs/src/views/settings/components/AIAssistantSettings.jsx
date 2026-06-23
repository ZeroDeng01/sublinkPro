import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import FormControlLabel from '@mui/material/FormControlLabel';
import Grid from '@mui/material/Grid';
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import CachedIcon from '@mui/icons-material/Cached';
import PsychologyIcon from '@mui/icons-material/Psychology';
import SaveIcon from '@mui/icons-material/Save';
import ScienceIcon from '@mui/icons-material/Science';

import { getAISettings, listAIModels, testAISettings, updateAISettings } from 'api/settings';

const DEFAULT_AI_MAX_TOKENS = 400000;

const AI_REQUEST_TYPES = [
  { value: 'responses', labelKey: 'settings.aiAssistantPanel.requestTypes.responses' },
  { value: 'chat_completions', labelKey: 'settings.aiAssistantPanel.requestTypes.chatCompletions' }
];

const dedupeModelOptions = (models = [], currentModel = '') => {
  const seen = new Set();
  const options = [];

  [currentModel, ...models].forEach((model) => {
    const value = String(model || '').trim();
    if (!value || seen.has(value)) {
      return;
    }
    seen.add(value);
    options.push(value);
  });

  return options;
};

const formatAIUsage = (usage) => {
  if (!usage || typeof usage !== 'object' || Array.isArray(usage)) {
    return '';
  }

  return JSON.stringify(usage, null, 2);
};

export default function AIAssistantSettings({ showMessage, loading, setLoading }) {
  const { t } = useTranslation();
  const [aiSettingsLoading, setAISettingsLoading] = useState(false);
  const [aiAction, setAIAction] = useState('');
  const [aiHeadersText, setAIHeadersText] = useState('{}');
  const [aiTestResult, setAITestResult] = useState(null);
  const [aiTestError, setAITestError] = useState('');
  const [aiModelOptions, setAIModelOptions] = useState([]);
  const [aiModelsFetched, setAIModelsFetched] = useState(false);
  const [aiForm, setAIForm] = useState({
    enabled: false,
    baseUrl: '',
    model: '',
    apiKey: '',
    maskedKey: '',
    hasKey: false,
    configured: false,
    providerType: 'openai_compatible',
    requestType: 'responses',
    temperature: 0.2,
    maxTokens: DEFAULT_AI_MAX_TOKENS
  });

  const setAIField = (field, value) => {
    setAIForm((prev) => ({ ...prev, [field]: value }));
    if (field === 'baseUrl' || field === 'apiKey') {
      setAIModelsFetched(false);
    }
  };

  const applyAISettingsData = useCallback((data = {}) => {
    setAIForm((prev) => ({
      ...prev,
      enabled: Boolean(data.enabled),
      baseUrl: data.baseUrl || '',
      model: data.model || '',
      apiKey: '',
      maskedKey: data.maskedKey || '',
      hasKey: Boolean(data.hasKey),
      configured: Boolean(data.configured),
      providerType: data.providerType || 'openai_compatible',
      requestType: data.requestType || 'responses',
      temperature: data.temperature ?? 0.2,
      maxTokens: data.maxTokens ?? DEFAULT_AI_MAX_TOKENS
    }));
    setAIModelOptions((prev) => dedupeModelOptions(prev, data.model || ''));
    setAIHeadersText(data.extraHeaders && Object.keys(data.extraHeaders).length > 0 ? JSON.stringify(data.extraHeaders, null, 2) : '{}');
  }, []);

  const fetchAISettings = useCallback(async () => {
    setAISettingsLoading(true);
    try {
      const response = await getAISettings();
      const data = response.data || {};
      applyAISettingsData(data);
      setAITestResult(null);
      setAITestError('');
    } catch (error) {
      console.error('获取 AI 设置失败:', error);
      showMessage(t('settings.aiAssistantPanel.messages.loadFailed', { message: error.response?.data?.message || error.message }), 'error');
    } finally {
      setAISettingsLoading(false);
    }
  }, [applyAISettingsData, showMessage, t]);

  useEffect(() => {
    fetchAISettings();
  }, [fetchAISettings]);

  const parseAIExtraHeaders = () => {
    const trimmed = aiHeadersText.trim();
    if (!trimmed) {
      return {};
    }

    let parsed;
    try {
      parsed = JSON.parse(trimmed);
    } catch {
      throw new Error(t('settings.aiAssistantPanel.messages.extraHeadersExample'));
    }

    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      throw new Error(t('settings.aiAssistantPanel.messages.extraHeadersObject'));
    }

    return Object.entries(parsed).reduce((acc, [key, value]) => {
      const headerKey = key.trim();
      if (headerKey) {
        acc[headerKey] = value == null ? '' : String(value);
      }
      return acc;
    }, {});
  };

  const buildAISettingsPayload = () => ({
    enabled: aiForm.enabled,
    baseUrl: aiForm.baseUrl.trim(),
    model: aiForm.model.trim(),
    requestType: aiForm.requestType,
    apiKey: aiForm.apiKey.trim(),
    temperature: aiForm.temperature === '' ? 0.2 : Number(aiForm.temperature),
    maxTokens: aiForm.maxTokens === '' ? 0 : Number(aiForm.maxTokens),
    extraHeaders: parseAIExtraHeaders()
  });

  const handleFetchAIModels = async () => {
    if (!aiForm.baseUrl.trim() && !aiForm.configured) {
      showMessage(t('settings.aiAssistantPanel.messages.baseUrlRequired'), 'warning');
      return;
    }
    if (!aiForm.apiKey.trim() && !aiForm.hasKey) {
      showMessage(t('settings.aiAssistantPanel.messages.apiKeyRequired'), 'warning');
      return;
    }

    let payload;
    try {
      payload = buildAISettingsPayload();
    } catch (error) {
      showMessage(error.message, 'warning');
      return;
    }

    setAIAction('models');
    setLoading(true);
    try {
      const response = await listAIModels(payload);
      const models = response.data?.models || [];
      setAIModelOptions(dedupeModelOptions(models, aiForm.model));
      setAIModelsFetched(true);
      showMessage(
        models.length > 0 ? t('settings.aiAssistantPanel.messages.modelsLoaded') : t('settings.aiAssistantPanel.messages.noModels'),
        models.length > 0 ? 'success' : 'info'
      );
    } catch (error) {
      showMessage(
        t('settings.aiAssistantPanel.messages.modelsFailed', { message: error.response?.data?.message || error.message }),
        'error'
      );
    } finally {
      setLoading(false);
      setAIAction('');
    }
  };

  const validateAISettingsPayload = (payload) => {
    if (payload.enabled && !payload.baseUrl) {
      throw new Error(t('settings.aiAssistantPanel.messages.enabledBaseUrlRequired'));
    }
    if (payload.enabled && !payload.model) {
      throw new Error(t('settings.aiAssistantPanel.messages.enabledModelRequired'));
    }
    if (payload.enabled && !payload.apiKey && !aiForm.hasKey) {
      throw new Error(t('settings.aiAssistantPanel.messages.enabledApiKeyRequired'));
    }
  };

  const handleTestAISettings = async () => {
    let payload;
    try {
      payload = buildAISettingsPayload();
      validateAISettingsPayload({ ...payload, enabled: true });
    } catch (error) {
      showMessage(error.message, 'warning');
      return;
    }

    setAIAction('test');
    setAITestResult(null);
    setAITestError('');
    setLoading(true);
    try {
      const response = await testAISettings(payload);
      setAITestResult(response.data || null);
      showMessage(t('settings.aiAssistantPanel.messages.testSuccess'));
    } catch (error) {
      setAITestResult(null);
      const message = error.response?.data?.message || error.message;
      setAITestError(message);
      showMessage(t('settings.aiAssistantPanel.messages.testFailed', { message }), 'error');
    } finally {
      setLoading(false);
      setAIAction('');
    }
  };

  const actionButtonSx = {
    minHeight: 44,
    px: 2.5,
    alignSelf: { xs: 'stretch', sm: 'center' }
  };

  const modelButtonSx = {
    height: 56,
    whiteSpace: 'nowrap'
  };

  const handleSaveAISettings = async () => {
    let payload;
    try {
      payload = buildAISettingsPayload();
      validateAISettingsPayload(payload);
    } catch (error) {
      showMessage(error.message, 'warning');
      return;
    }

    setAIAction('save');
    setLoading(true);
    try {
      const response = await updateAISettings(payload);
      applyAISettingsData(response.data || {});
      showMessage(t('settings.aiAssistantPanel.messages.saveSuccess'));
    } catch (error) {
      showMessage(t('settings.aiAssistantPanel.messages.saveFailed', { message: error.response?.data?.message || error.message }), 'error');
    } finally {
      setLoading(false);
      setAIAction('');
    }
  };

  const runWithoutPageJump = async (event, action) => {
    event?.preventDefault();

    const position = { left: window.scrollX, top: window.scrollY };
    const restorePosition = () => window.scrollTo({ ...position, behavior: 'auto' });

    requestAnimationFrame(restorePosition);
    try {
      await action();
    } finally {
      requestAnimationFrame(restorePosition);
      window.setTimeout(restorePosition, 0);
      window.setTimeout(restorePosition, 120);
    }
  };

  const aiUsageText = formatAIUsage(aiTestResult?.usage);
  const selectedRequestTypeLabel = t(
    AI_REQUEST_TYPES.find((type) => type.value === aiForm.requestType)?.labelKey || 'settings.aiAssistantPanel.requestTypes.responses'
  );

  return (
    <Card variant="outlined">
      <CardHeader
        title={t('settings.aiAssistantPanel.title')}
        subheader={t('settings.aiAssistantPanel.subheader')}
        avatar={<PsychologyIcon color="primary" />}
        action={
          <FormControlLabel
            sx={{ mr: 0 }}
            control={<Switch checked={aiForm.enabled} onChange={(e) => setAIField('enabled', e.target.checked)} />}
            label={aiForm.enabled ? t('common.enabled') : t('common.disabled')}
          />
        }
      />
      <CardContent>
        {aiSettingsLoading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress size={28} />
          </Box>
        ) : (
          <Stack spacing={2.5}>
            <Alert severity="info">{t('settings.aiAssistantPanel.alerts.systemConfig')}</Alert>

            <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2 }}>
              <Stack spacing={2}>
                <Box>
                  <Typography variant="subtitle2">{t('settings.aiAssistantPanel.enableTitle')}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {t('settings.aiAssistantPanel.providerType', { type: selectedRequestTypeLabel })}
                  </Typography>
                </Box>

                <TextField
                  select
                  fullWidth
                  label={t('settings.aiAssistantPanel.fields.requestType')}
                  value={aiForm.requestType}
                  onChange={(e) => setAIField('requestType', e.target.value)}
                  helperText={t('settings.aiAssistantPanel.fields.requestTypeHelper')}
                >
                  {AI_REQUEST_TYPES.map((type) => (
                    <MenuItem key={type.value} value={type.value}>
                      {t(type.labelKey)}
                    </MenuItem>
                  ))}
                </TextField>

                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  <Chip
                    size="small"
                    color={aiForm.configured ? 'info' : 'default'}
                    variant={aiForm.configured ? 'filled' : 'outlined'}
                    label={
                      aiForm.configured
                        ? t('settings.aiAssistantPanel.status.configured')
                        : t('settings.aiAssistantPanel.status.notConfigured')
                    }
                  />
                  <Chip
                    size="small"
                    color={aiForm.hasKey ? 'success' : 'default'}
                    variant="outlined"
                    label={
                      aiForm.hasKey
                        ? t('settings.aiAssistantPanel.status.savedApiKey', {
                            key: aiForm.maskedKey || t('settings.aiAssistantPanel.status.hidden')
                          })
                        : t('settings.aiAssistantPanel.status.noApiKey')
                    }
                  />
                </Stack>

                <TextField
                  fullWidth
                  label="AI Base URL"
                  value={aiForm.baseUrl}
                  onChange={(e) => setAIField('baseUrl', e.target.value)}
                  placeholder="https://api.openai.com/v1"
                  helperText={t('settings.aiAssistantPanel.fields.baseUrlHelper')}
                />

                <TextField
                  fullWidth
                  type="password"
                  label={aiForm.hasKey ? t('settings.aiAssistantPanel.fields.replaceApiKey') : 'API Key'}
                  value={aiForm.apiKey}
                  onChange={(e) => setAIField('apiKey', e.target.value)}
                  autoComplete="off"
                />

                <Grid container spacing={2} alignItems="flex-start">
                  <Grid item xs={12} md={8}>
                    <Autocomplete
                      freeSolo
                      options={aiModelOptions}
                      value={aiForm.model}
                      inputValue={aiForm.model}
                      onInputChange={(_event, value) => setAIField('model', value || '')}
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label={t('settings.aiAssistantPanel.fields.model')}
                          placeholder="gpt-4.1-mini"
                          helperText={
                            aiModelsFetched
                              ? t('settings.aiAssistantPanel.fields.modelHelperFetched')
                              : t('settings.aiAssistantPanel.fields.modelHelper')
                          }
                        />
                      )}
                    />
                  </Grid>
                  <Grid item xs={12} md={4}>
                    <Stack spacing={0.5}>
                      <Button
                        type="button"
                        fullWidth
                        variant="outlined"
                        disabled={loading || aiSettingsLoading}
                        startIcon={loading && aiAction === 'models' ? <CircularProgress size={18} /> : <CachedIcon />}
                        onClick={(event) => runWithoutPageJump(event, handleFetchAIModels)}
                        sx={modelButtonSx}
                      >
                        {t('settings.aiAssistantPanel.actions.fetchModels')}
                      </Button>
                      <Typography variant="caption" color="text.secondary" sx={{ px: 1.75, lineHeight: 1.66 }}>
                        {t('settings.aiAssistantPanel.actions.fetchModelsHelper')}
                      </Typography>
                    </Stack>
                  </Grid>
                </Grid>

                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems="flex-start" flexWrap="wrap" useFlexGap>
                  <TextField
                    type="number"
                    label="Temperature"
                    value={aiForm.temperature}
                    onChange={(e) => setAIField('temperature', e.target.value)}
                    slotProps={{ htmlInput: { min: 0, max: 2, step: 0.1 } }}
                    sx={{ width: { xs: '100%', sm: 180 } }}
                  />
                  <TextField
                    type="number"
                    label="Max Tokens"
                    value={aiForm.maxTokens}
                    onChange={(e) => setAIField('maxTokens', e.target.value)}
                    slotProps={{ htmlInput: { min: 0, step: 100 } }}
                    helperText={t('settings.aiAssistantPanel.fields.maxTokensHelper')}
                    sx={{ width: { xs: '100%', sm: 280 } }}
                  />
                </Stack>

                <TextField
                  fullWidth
                  multiline
                  minRows={3}
                  label={t('settings.aiAssistantPanel.fields.extraHeaders')}
                  value={aiHeadersText}
                  onChange={(e) => setAIHeadersText(e.target.value)}
                  helperText={t('settings.aiAssistantPanel.fields.extraHeadersHelper')}
                />
              </Stack>
            </Box>

            <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 2, bgcolor: 'background.default' }}>
              <Stack spacing={1.5}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <ScienceIcon color="primary" fontSize="small" />
                  <Typography variant="subtitle2">{t('settings.aiAssistantPanel.testResult.title')}</Typography>
                </Stack>

                {loading && aiAction === 'test' ? (
                  <Alert severity="info" icon={<CircularProgress size={18} />}>
                    {t('settings.aiAssistantPanel.testResult.testing')}
                  </Alert>
                ) : aiTestError ? (
                  <Alert severity="error">{aiTestError}</Alert>
                ) : aiTestResult ? (
                  <Stack spacing={1.75}>
                    <Alert severity="success">{t('settings.aiAssistantPanel.testResult.success')}</Alert>
                    <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, bgcolor: 'background.paper', p: 1.5 }}>
                      <Typography variant="subtitle2" sx={{ mb: 0.75 }}>
                        {t('settings.aiAssistantPanel.testResult.response')}
                      </Typography>
                      <Typography variant="body2" color="text.secondary" sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                        {aiTestResult.message || '-'}
                      </Typography>
                    </Box>
                    <Grid container spacing={2}>
                      <Grid item xs={12} sm={6}>
                        <Typography variant="subtitle2">{t('settings.aiAssistantPanel.fields.model')}</Typography>
                        <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-word' }}>
                          {aiTestResult.model || aiForm.model || '-'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} sm={6}>
                        <Typography variant="subtitle2">{t('settings.aiAssistantPanel.testResult.latency')}</Typography>
                        <Typography variant="body2" color="text.secondary">
                          {typeof aiTestResult.latencyMs === 'number' ? `${aiTestResult.latencyMs} ms` : '-'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} sm={6}>
                        <Typography variant="subtitle2">{t('settings.aiAssistantPanel.testResult.finishReason')}</Typography>
                        <Typography variant="body2" color="text.secondary">
                          {aiTestResult.finishReason || '-'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} sm={6}>
                        <Typography variant="subtitle2">Base URL</Typography>
                        <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-word' }}>
                          {aiTestResult.baseUrl || aiForm.baseUrl || '-'}
                        </Typography>
                      </Grid>
                    </Grid>
                    {aiUsageText && (
                      <Box>
                        <Typography variant="subtitle2" sx={{ mb: 0.75 }}>
                          {t('settings.aiAssistantPanel.testResult.usage')}
                        </Typography>
                        <Box
                          component="pre"
                          sx={{
                            m: 0,
                            border: '1px solid',
                            borderColor: 'divider',
                            borderRadius: 1,
                            bgcolor: 'background.paper',
                            color: 'text.primary',
                            fontFamily: 'monospace',
                            fontSize: '0.8125rem',
                            overflowX: 'auto',
                            p: 1.5,
                            whiteSpace: 'pre-wrap'
                          }}
                        >
                          {aiUsageText}
                        </Box>
                      </Box>
                    )}
                  </Stack>
                ) : (
                  <Typography variant="body2" color="text.secondary">
                    {t('settings.aiAssistantPanel.testResult.empty')}
                  </Typography>
                )}
              </Stack>
            </Box>

            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} alignItems={{ xs: 'stretch', sm: 'center' }}>
              <Button
                type="button"
                variant="outlined"
                startIcon={loading && aiAction === 'test' ? <CircularProgress size={18} /> : <ScienceIcon />}
                onClick={(event) => runWithoutPageJump(event, handleTestAISettings)}
                disabled={loading}
                sx={actionButtonSx}
              >
                {t('settings.aiAssistantPanel.actions.test')}
              </Button>
              <Button
                type="button"
                variant="contained"
                startIcon={loading && aiAction === 'save' ? <CircularProgress size={18} /> : <SaveIcon />}
                onClick={(event) => runWithoutPageJump(event, handleSaveAISettings)}
                disabled={loading || aiSettingsLoading}
                sx={actionButtonSx}
              >
                {t('settings.aiAssistantPanel.actions.save')}
              </Button>
            </Stack>
          </Stack>
        )}
      </CardContent>
    </Card>
  );
}
