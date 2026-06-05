import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

// material-ui
import { alpha, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Autocomplete from '@mui/material/Autocomplete';
import Switch from '@mui/material/Switch';

// icons
import AddIcon from '@mui/icons-material/Add';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import TransformIcon from '@mui/icons-material/Transform';
import UnfoldMoreIcon from '@mui/icons-material/UnfoldMore';
import FullscreenIcon from '@mui/icons-material/Fullscreen';
import FullscreenExitIcon from '@mui/icons-material/FullscreenExit';
import CircularProgress from '@mui/material/CircularProgress';
import CompareArrowsIcon from '@mui/icons-material/CompareArrows';
import CheckIcon from '@mui/icons-material/Check';
import UndoIcon from '@mui/icons-material/Undo';

import MainCard from 'ui-component/cards/MainCard';
import Pagination from 'components/Pagination';
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import {
  getTemplates,
  addTemplate,
  updateTemplate,
  deleteTemplate,
  getTemplateUsage,
  getACL4SSRPresets,
  convertRules,
  generateTemplateAICandidateStream
} from 'api/templates';
import { getAISettings, getBaseTemplates, updateBaseTemplate } from 'api/settings';
import { getNodes } from 'api/nodes';
import { withAlpha } from 'utils/colorUtils';

// Monaco Editor
import Editor, { DiffEditor } from '@monaco-editor/react';

const createEmptyTemplateAIAssistant = () => ({
  summary: '',
  warnings: [],
  candidateText: '',
  revisionHash: '',
  validation: null,
  finishReason: '',
  usage: null,
  sourceText: '',
  sourceFilename: '',
  sourceCategory: '',
  sourceRuleSource: '',
  sourceUseProxy: false,
  sourceProxyLink: '',
  sourceEnableIncludeAll: false
});

const normalizeMessages = (messages) => (Array.isArray(messages) ? messages.filter(Boolean) : []);

const JSON_ESCAPE_CHAR_MAP = {
  '"': '"',
  '\\': '\\',
  '/': '/',
  b: '\b',
  f: '\f',
  n: '\n',
  r: '\r',
  t: '\t'
};

const AI_SETUP_ERROR_MARKERS = [
  'AI \u52a9\u624b\u672a\u542f\u7528',
  'AI \u8bbe\u7f6e\u4e0d\u5b8c\u6574\uff0c\u8bf7\u5148\u914d\u7f6e Base URL\u3001\u6a21\u578b\u548c API Key'
];

const createTemplateAISourceSnapshot = (formData, useProxy, proxyLink) => ({
  sourceText: formData.text,
  sourceFilename: formData.filename.trim(),
  sourceCategory: formData.category,
  sourceRuleSource: formData.ruleSource,
  sourceUseProxy: useProxy,
  sourceProxyLink: proxyLink,
  sourceEnableIncludeAll: formData.enableIncludeAll
});

const buildTemplateAIAssistantState = (payload, sourceSnapshot, fallbackCandidateText = '') => ({
  summary: payload?.summary || '',
  warnings: normalizeMessages(payload?.warnings),
  candidateText: payload?.candidateText || fallbackCandidateText,
  revisionHash: payload?.revisionHash || '',
  validation: payload?.validation || null,
  finishReason: payload?.finishReason || '',
  usage: payload?.usage || null,
  ...sourceSnapshot
});

const extractResponseUsage = (eventData) => {
  if (!eventData || typeof eventData !== 'object' || Array.isArray(eventData)) {
    return null;
  }

  const response = eventData.response;
  if (!response || typeof response !== 'object' || Array.isArray(response)) {
    return null;
  }

  return response.usage && typeof response.usage === 'object' && !Array.isArray(response.usage) ? response.usage : null;
};

const extractResponseFinishReason = (eventData) => {
  if (!eventData || typeof eventData !== 'object' || Array.isArray(eventData)) {
    return '';
  }

  const response = eventData.response;
  if (!response || typeof response !== 'object' || Array.isArray(response)) {
    return '';
  }

  return typeof response.status === 'string' ? response.status : '';
};

const getUsageNumber = (container, key) => {
  if (!container || typeof container !== 'object' || Array.isArray(container) || !Object.prototype.hasOwnProperty.call(container, key)) {
    return null;
  }

  const value = container[key];

  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === 'string' && value.trim() !== '') {
    const parsedValue = Number(value);

    if (Number.isFinite(parsedValue)) {
      return parsedValue;
    }
  }

  return null;
};

const buildTemplateAIUsageItems = (usage, t) => {
  if (!usage || typeof usage !== 'object' || Array.isArray(usage)) {
    return [];
  }

  const inputTokens = getUsageNumber(usage, 'input_tokens') ?? getUsageNumber(usage, 'prompt_tokens');
  const outputTokens = getUsageNumber(usage, 'output_tokens') ?? getUsageNumber(usage, 'completion_tokens');
  const inputTokenDetails =
    usage.input_tokens_details && typeof usage.input_tokens_details === 'object' && !Array.isArray(usage.input_tokens_details)
      ? usage.input_tokens_details
      : null;
  const cacheTokens =
    getUsageNumber(inputTokenDetails, 'cached_tokens') ??
    getUsageNumber(usage, 'cached_tokens') ??
    getUsageNumber(usage, 'cache_tokens') ??
    getUsageNumber(usage, 'cached_input_tokens');

  return [
    inputTokens !== null ? { key: 'input', label: t('templates.ai.usage.input'), value: inputTokens } : null,
    outputTokens !== null ? { key: 'output', label: t('templates.ai.usage.output'), value: outputTokens } : null,
    cacheTokens !== null ? { key: 'cache', label: t('templates.ai.usage.cache'), value: cacheTokens } : null
  ].filter(Boolean);
};

const decodePartialJSONString = (value, startIndex) => {
  let decoded = '';

  for (let index = startIndex; index < value.length; index += 1) {
    const currentChar = value[index];

    if (currentChar === '"') {
      break;
    }

    if (currentChar !== '\\') {
      decoded += currentChar;
      continue;
    }

    if (index + 1 >= value.length) {
      break;
    }

    const nextChar = value[index + 1];

    if (nextChar === 'u') {
      const unicodeHex = value.slice(index + 2, index + 6);
      if (unicodeHex.length < 4 || !/^[0-9a-fA-F]{4}$/.test(unicodeHex)) {
        break;
      }
      decoded += String.fromCharCode(parseInt(unicodeHex, 16));
      index += 5;
      continue;
    }

    decoded += JSON_ESCAPE_CHAR_MAP[nextChar] ?? nextChar;
    index += 1;
  }

  return decoded;
};

const extractCandidatePreviewFromStream = (streamBuffer) => {
  if (!streamBuffer) {
    return '';
  }

  const keyMatch = /"candidateText"\s*:\s*"/.exec(streamBuffer);
  if (!keyMatch) {
    return '';
  }

  return decodePartialJSONString(streamBuffer, keyMatch.index + keyMatch[0].length);
};

export default function TemplateList() {
  const { t } = useTranslation();
  const theme = useTheme();
  const palette = theme.vars?.palette || theme.palette;
  const isDark = theme.palette.mode === 'dark';
  const aiPromptPrimaryLight = theme.palette.primary.light;
  const aiPromptPrimaryMain = theme.palette.primary.main;
  const aiPromptSurface = theme.palette.grey[900];
  const aiPromptCollapsedSurface = theme.palette.grey[800];
  const navigate = useNavigate();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));
  const aiGenerationAbortRef = useRef(null);
  const aiStreamBufferRef = useRef('');

  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [currentTemplate, setCurrentTemplate] = useState(null);
  const [formData, setFormData] = useState({ filename: '', text: '', category: 'clash', ruleSource: '', enableIncludeAll: false });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [aclPresets, setAclPresets] = useState([]);
  const [converting, setConverting] = useState(false);
  const [editorFullscreen, setEditorFullscreen] = useState(false);
  const [templateEditorMode, setTemplateEditorMode] = useState('edit');
  const [aiPrompt, setAIPrompt] = useState('');
  const [aiGenerating, setAIGenerating] = useState(false);
  const [aiAssistant, setAIAssistant] = useState(createEmptyTemplateAIAssistant);
  const [aiGenerationError, setAIGenerationError] = useState('');
  const [aiLocalAcceptSnapshot, setAILocalAcceptSnapshot] = useState(null);
  const [isAIEnabled, setIsAIEnabled] = useState(false);
  const [aiCommandOpen, setAICommandOpen] = useState(false);
  const [aiDisabledPromptOpen, setAIDisabledPromptOpen] = useState(false);
  const [errorDialog, setErrorDialog] = useState({ open: false, title: '', message: '' });
  const [usageDialog, setUsageDialog] = useState({ open: false, title: '', message: '', subscriptions: [], action: null });
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('templates_rowsPerPage');
    return saved ? parseInt(saved, 10) : 10;
  });
  const [totalItems, setTotalItems] = useState(0);

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({
    title: '',
    content: '',
    action: null
  });

  const [baseTemplateDialogOpen, setBaseTemplateDialogOpen] = useState(false);
  const [baseTemplateCategory, setBaseTemplateCategory] = useState('clash');
  const [baseTemplateContent, setBaseTemplateContent] = useState('');
  const [baseTemplateLoading, setBaseTemplateLoading] = useState(false);
  const [baseTemplateSaving, setBaseTemplateSaving] = useState(false);

  const [useProxy, setUseProxy] = useState(false);
  const [proxyLink, setProxyLink] = useState('');
  const [proxyNodeOptions, setProxyNodeOptions] = useState([]);
  const [loadingProxyNodes, setLoadingProxyNodes] = useState(false);

  const openConfirm = (title, content, action) => {
    setConfirmInfo({ title, content, action });
    setConfirmOpen(true);
  };

  const abortAIGeneration = () => {
    if (aiGenerationAbortRef.current) {
      aiGenerationAbortRef.current.abort();
      aiGenerationAbortRef.current = null;
    }
    aiStreamBufferRef.current = '';
  };

  const resetTemplateAIAssistant = () => {
    abortAIGeneration();
    setTemplateEditorMode('edit');
    setAIPrompt('');
    setAICommandOpen(false);
    setAIAssistant(createEmptyTemplateAIAssistant());
    setAIGenerationError('');
    setAIGenerating(false);
    setAILocalAcceptSnapshot(null);
  };

  const handleConfirmClose = () => {
    setConfirmOpen(false);
  };

  const handleConfirmAction = async () => {
    if (confirmInfo.action) {
      await confirmInfo.action();
    }
    setConfirmOpen(false);
  };

  const fetchTemplates = async (currentPage, currentPageSize) => {
    setLoading(true);
    try {
      const response = await getTemplates({ page: currentPage + 1, pageSize: currentPageSize });
      if (response.data && response.data.items !== undefined) {
        setTemplates(response.data.items || []);
        setTotalItems(response.data.total || 0);
      } else {
        setTemplates(response.data || []);
        setTotalItems((response.data || []).length);
      }
    } catch (error) {
      console.log(error);
      showMessage(error.message || t('templates.messages.loadFailed'), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    fetchTemplates(page, rowsPerPage);
  };

  useEffect(() => {
    fetchTemplates(0, rowsPerPage);
    getAISettings()
      .then((res) => {
        setIsAIEnabled(Boolean(res.data?.enabled));
      })
      .catch((err) => console.log('Failed to load AI settings:', err));
    getACL4SSRPresets()
      .then((res) => {
        if (res.data) {
          setAclPresets(res.data);
        }
      })
      .catch((err) => console.log('Failed to load preset list:', err));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleAdd = () => {
    setIsEdit(false);
    setCurrentTemplate(null);
    setFormData({ filename: '', text: '', category: 'clash', ruleSource: '', enableIncludeAll: false });
    setUseProxy(false);
    setProxyLink('');
    setEditorFullscreen(false);
    resetTemplateAIAssistant();
    setDialogOpen(true);
  };

  const handleEdit = (template) => {
    setIsEdit(true);
    setCurrentTemplate(template);
    setEditorFullscreen(false);
    setFormData({
      filename: template.file,
      text: template.text,
      category: template.category || 'clash',
      ruleSource: template.ruleSource || '',
      enableIncludeAll: template.enableIncludeAll || false
    });
    setUseProxy(template.useProxy || false);
    setProxyLink(template.proxyLink || '');
    if (template.useProxy) {
      fetchProxyNodes();
    }
    resetTemplateAIAssistant();
    setDialogOpen(true);
  };

  const handleDelete = async (template) => {
    let usedSubscriptions = [];

    try {
      const response = await getTemplateUsage({ filename: template.file });
      usedSubscriptions = response.data?.subscriptions || [];
    } catch (error) {
      console.log(error);
      showMessage(error.message || t('templates.messages.usageFailed'), 'error');
      return;
    }

    const deleteAction = async () => {
      try {
        await deleteTemplate({ filename: template.file });
        showMessage(t('templates.messages.deleteSuccess'));
        fetchTemplates(page, rowsPerPage);
      } catch (error) {
        console.log(error);
        showMessage(error.message || t('templates.messages.deleteFailed'), 'error');
      }
    };

    if (usedSubscriptions.length > 0) {
      setUsageDialog({
        open: true,
        title: t('templates.usage.title'),
        message: t('templates.usage.message', { name: template.file }),
        subscriptions: usedSubscriptions,
        action: deleteAction
      });
      return;
    }

    openConfirm(t('templates.delete.title'), t('templates.delete.confirm', { name: template.file }), deleteAction);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditorFullscreen(false);
    resetTemplateAIAssistant();
  };

  useEffect(() => () => abortAIGeneration(), []);

  const handleGenerateWithAI = async () => {
    if (!aiPrompt.trim()) {
      showMessage(t('templates.ai.messages.promptRequired'), 'warning');
      return;
    }

    abortAIGeneration();
    const sourceSnapshot = createTemplateAISourceSnapshot(formData, useProxy, proxyLink);
    const controller = new AbortController();
    aiGenerationAbortRef.current = controller;
    aiStreamBufferRef.current = '';
    let latestCandidatePreview = '';

    setAIGenerating(true);
    setAIGenerationError('');
    setTemplateEditorMode('edit');
    setAILocalAcceptSnapshot(null);
    setAIAssistant({
      ...createEmptyTemplateAIAssistant(),
      ...sourceSnapshot
    });

    try {
      const data = await generateTemplateAICandidateStream(
        {
          filename: formData.filename.trim(),
          category: formData.category,
          currentText: formData.text,
          userPrompt: aiPrompt.trim(),
          ruleSource: formData.ruleSource,
          useProxy,
          proxyLink,
          enableIncludeAll: formData.enableIncludeAll
        },
        {
          signal: controller.signal,
          onStart: () => {
            aiStreamBufferRef.current = '';
          },
          onDelta: (eventData) => {
            const deltaText = typeof eventData === 'string' ? eventData : eventData?.delta || '';
            if (!deltaText) {
              return;
            }

            aiStreamBufferRef.current += deltaText;
            const nextCandidatePreview = extractCandidatePreviewFromStream(aiStreamBufferRef.current);
            latestCandidatePreview = nextCandidatePreview || latestCandidatePreview;

            setAIAssistant((prev) => ({
              ...prev,
              candidateText: nextCandidatePreview || prev.candidateText
            }));
          },
          onComplete: (eventData) => {
            if (!eventData || typeof eventData !== 'object') {
              return;
            }
            setAIAssistant((prev) => ({
              ...prev,
              finishReason: extractResponseFinishReason(eventData) || prev.finishReason,
              usage: extractResponseUsage(eventData) || prev.usage,
              candidateText: latestCandidatePreview || prev.candidateText
            }));
          },
          onFinal: (eventData) => {
            if (!eventData || typeof eventData !== 'object') {
              return;
            }

            const nextAssistantState = buildTemplateAIAssistantState(eventData, sourceSnapshot, latestCandidatePreview);
            latestCandidatePreview = nextAssistantState.candidateText;
            setAIAssistant(nextAssistantState);
          }
        }
      );

      const finalAssistantState = buildTemplateAIAssistantState(data, sourceSnapshot, latestCandidatePreview);
      setAIGenerationError('');

      if (finalAssistantState.candidateText) {
        setTemplateEditorMode('diff');
        showMessage(t('templates.ai.messages.generated'));
      } else {
        showMessage(t('templates.ai.messages.emptyCandidate'), 'warning');
      }
    } catch (error) {
      if (error.name === 'AbortError') {
        return;
      }

      const errorMessage = error.response?.data?.message || error.message || t('templates.ai.messages.generateFailed');
      const friendlyErrorMessage = AI_SETUP_ERROR_MARKERS.some((marker) => errorMessage.includes(marker))
        ? t('templates.ai.messages.setupUnavailable')
        : errorMessage;
      setAIGenerationError(errorMessage);
      showMessage(friendlyErrorMessage, 'error');
    } finally {
      if (aiGenerationAbortRef.current === controller) {
        aiGenerationAbortRef.current = null;
      }
      setAIGenerating(false);
    }
  };

  const aiCandidateMatchesEditor = Boolean(aiAssistant.candidateText) && aiAssistant.candidateText === formData.text;
  const aiCandidateOutdated =
    Boolean(aiAssistant.candidateText) &&
    !aiCandidateMatchesEditor &&
    (aiAssistant.sourceText !== formData.text ||
      aiAssistant.sourceFilename !== formData.filename.trim() ||
      aiAssistant.sourceCategory !== formData.category ||
      aiAssistant.sourceRuleSource !== formData.ruleSource ||
      aiAssistant.sourceUseProxy !== useProxy ||
      aiAssistant.sourceProxyLink !== proxyLink ||
      aiAssistant.sourceEnableIncludeAll !== formData.enableIncludeAll);
  const canReviewAICandidate = Boolean(aiAssistant.candidateText) && !aiCandidateOutdated && !aiCandidateMatchesEditor;
  const isDiffMode = templateEditorMode === 'diff';
  const showDiffReview = isDiffMode && canReviewAICandidate;
  const canAcceptAICandidateLocally = Boolean(aiAssistant.candidateText) && !aiCandidateOutdated && !aiCandidateMatchesEditor;
  const canRevertLocalAIAccept = Boolean(aiLocalAcceptSnapshot);
  const canSwitchToDiffMode = canReviewAICandidate;

  useEffect(() => {
    if (templateEditorMode === 'diff' && (!canSwitchToDiffMode || !isAIEnabled)) {
      setTemplateEditorMode('edit');
    }
  }, [templateEditorMode, canSwitchToDiffMode, isAIEnabled]);

  const handleAcceptAICandidateLocally = () => {
    if (!aiAssistant.candidateText) {
      showMessage(t('templates.ai.messages.generateFirst'), 'warning');
      return;
    }

    if (aiCandidateMatchesEditor) {
      showMessage(t('templates.ai.messages.alreadyApplied'), 'info');
      return;
    }

    if (aiCandidateOutdated) {
      showMessage(t('templates.ai.messages.outdatedAccept'), 'warning');
      return;
    }

    setAILocalAcceptSnapshot({ text: formData.text });
    setTemplateEditorMode('edit');
    setFormData((prev) => ({
      ...prev,
      text: aiAssistant.candidateText
    }));
    showMessage(t('templates.ai.messages.accepted'));
  };

  const handleRevertLastLocalAIAccept = () => {
    if (!aiLocalAcceptSnapshot) {
      showMessage(t('templates.ai.messages.noRevert'), 'warning');
      return;
    }

    setFormData((prev) => ({
      ...prev,
      text: aiLocalAcceptSnapshot.text || ''
    }));
    setTemplateEditorMode('edit');
    setAILocalAcceptSnapshot(null);
    showMessage(t('templates.ai.messages.reverted'));
  };

  const handleConvertTemplate = async (expand) => {
    setConverting(true);
    try {
      const res = await convertRules({
        ruleSource: formData.ruleSource,
        category: formData.category,
        expand,
        template: formData.text,
        useProxy: useProxy,
        proxyLink: proxyLink,
        enableIncludeAll: formData.enableIncludeAll
      });
      if (res.code === 200 && res.data && res.data.content) {
        setFormData({ ...formData, text: res.data.content });
        showMessage(expand ? t('templates.messages.convertExpandSuccess') : t('templates.messages.convertSuccess'));
      } else {
        setErrorDialog({
          open: true,
          title: t('templates.messages.convertFailed'),
          message: res.msg || t('templates.messages.convertError')
        });
      }
    } catch (error) {
      console.error(error);
      const errorMsg = error.response?.data?.msg || error.message || t('templates.messages.convertFailed');
      setErrorDialog({
        open: true,
        title: t('templates.messages.convertFailed'),
        message: errorMsg
      });
    } finally {
      setConverting(false);
    }
  };

  const handleSubmit = async () => {
    try {
      if (templateEditorMode === 'diff') {
        showMessage(t('templates.ai.messages.diffSaveBlocked'), 'warning');
        return;
      }

      if (isEdit) {
        await updateTemplate({
          oldname: currentTemplate.file,
          filename: formData.filename,
          text: formData.text,
          category: formData.category,
          ruleSource: formData.ruleSource,
          useProxy: useProxy,
          proxyLink: proxyLink,
          enableIncludeAll: formData.enableIncludeAll
        });
        showMessage(t('templates.messages.updateSuccess'));
      } else {
        await addTemplate({
          filename: formData.filename,
          text: formData.text,
          category: formData.category,
          ruleSource: formData.ruleSource,
          useProxy: useProxy,
          proxyLink: proxyLink,
          enableIncludeAll: formData.enableIncludeAll
        });
        showMessage(t('templates.messages.addSuccess'));
      }
      setEditorFullscreen(false);
      setDialogOpen(false);
      resetTemplateAIAssistant();
      fetchTemplates(page, rowsPerPage);
    } catch (error) {
      console.log(error);
      showMessage(error.message || (isEdit ? t('templates.messages.updateFailed') : t('templates.messages.addFailed')), 'error');
    }
  };

  const handleOpenBaseTemplate = async (category) => {
    setBaseTemplateCategory(category);
    setBaseTemplateDialogOpen(true);
    setBaseTemplateLoading(true);
    try {
      const res = await getBaseTemplates();
      if (res.data) {
        const content = category === 'clash' ? res.data.clashTemplate : res.data.surgeTemplate;
        setBaseTemplateContent(content || '');
      }
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('templates.messages.baseTemplateLoadFailed'), 'error');
    } finally {
      setBaseTemplateLoading(false);
    }
  };

  const handleSaveBaseTemplate = async () => {
    setBaseTemplateSaving(true);
    try {
      await updateBaseTemplate(baseTemplateCategory, baseTemplateContent);
      showMessage(t('templates.messages.baseTemplateSaveSuccess', { category: baseTemplateCategory === 'clash' ? 'Clash' : 'Surge' }));
      setBaseTemplateDialogOpen(false);
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('templates.messages.baseTemplateSaveFailed'), 'error');
    } finally {
      setBaseTemplateSaving(false);
    }
  };

  const fetchProxyNodes = async () => {
    setLoadingProxyNodes(true);
    try {
      const res = await getNodes({ pageSize: 100 });
      if (res.data) {
        const items = res.data.items || res.data || [];
        setProxyNodeOptions(items);
      }
    } catch (error) {
      console.error('Failed to load proxy nodes:', error);
    } finally {
      setLoadingProxyNodes(false);
    }
  };

  const outlinedLabelFixSx = {
    '& .MuiInputLabel-shrink': {
      px: 0.5,
      backgroundColor: 'background.paper',
      lineHeight: 1.35,
      transform: 'translate(14px, -4px) scale(0.75)'
    }
  };

  const aiWorkspacePanelSx = {
    border: 1,
    borderColor: 'divider',
    borderRadius: 1,
    bgcolor: 'background.paper'
  };

  const isEditMode = templateEditorMode === 'edit';

  const aiStatusText = aiGenerating
    ? t('templates.ai.status.generating')
    : aiGenerationError
      ? aiGenerationError
      : aiCandidateOutdated
        ? t('templates.ai.status.outdated')
        : !isEdit && aiAssistant.candidateText
          ? t('templates.ai.status.unsavedTemplate')
          : showDiffReview
            ? t('templates.ai.status.diffReadonly')
            : aiCandidateMatchesEditor
              ? t('templates.ai.status.applied')
              : canRevertLocalAIAccept
                ? t('templates.ai.status.revertAvailable')
                : aiAssistant.candidateText
                  ? t('templates.ai.status.ready')
                  : t('templates.ai.status.prompt');
  const aiStatusColor = aiGenerationError
    ? 'error.main'
    : aiCandidateOutdated
      ? 'warning.main'
      : showDiffReview
        ? alpha(theme.palette.common.white, 0.92)
        : aiCandidateMatchesEditor
          ? 'success.main'
          : alpha(theme.palette.common.white, 0.88);
  const isAISetupIssue = AI_SETUP_ERROR_MARKERS.some((marker) => aiGenerationError.includes(marker));
  const aiSetupGuidanceText = isAISetupIssue ? t('templates.ai.setupGuidance') : '';
  const aiFriendlyGenerationError = isAISetupIssue ? t('templates.ai.setupUnavailable') : aiGenerationError;
  const aiUsageItems = buildTemplateAIUsageItems(aiAssistant.usage, t);

  const configureTemplateMonacoTheme = (monaco) => {
    monaco.editor.defineTheme('template-ai-editor', {
      base: 'vs-dark',
      inherit: true,
      rules: [],
      colors: {
        'editor.rangeHighlightBackground': '#00000000'
      }
    });
  };

  const aiStateChips = (
    <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
      <Chip
        size="small"
        variant="filled"
        label={isEditMode ? t('templates.ai.mode.edit') : t('templates.ai.mode.diff')}
        color={isEditMode ? 'primary' : 'default'}
        sx={{
          color: 'common.white',
          bgcolor: isEditMode ? undefined : alpha(theme.palette.common.white, 0.14),
          '& .MuiChip-label': {
            fontWeight: 600
          }
        }}
      />
      {aiGenerating ? (
        <Chip
          size="small"
          variant="outlined"
          color="primary"
          label={t('templates.ai.statusChip.generating')}
          sx={{ color: 'common.white' }}
        />
      ) : null}
      {!aiGenerating ? (
        <Chip
          size="small"
          variant="outlined"
          color={
            aiGenerationError
              ? 'error'
              : aiCandidateOutdated
                ? 'warning'
                : aiCandidateMatchesEditor
                  ? 'success'
                  : aiAssistant.candidateText
                    ? 'info'
                    : 'default'
          }
          label={
            aiGenerationError
              ? t('templates.ai.statusChip.failed')
              : aiCandidateOutdated
                ? t('templates.ai.statusChip.outdated')
                : aiCandidateMatchesEditor
                  ? t('templates.ai.statusChip.applied')
                  : aiAssistant.candidateText
                    ? t('templates.ai.statusChip.ready')
                    : t('templates.ai.statusChip.notGenerated')
          }
          sx={{
            color:
              aiGenerationError || aiCandidateOutdated || aiCandidateMatchesEditor || aiAssistant.candidateText
                ? 'common.white'
                : alpha(theme.palette.common.white, 0.92),
            borderColor:
              !aiGenerationError && !aiCandidateOutdated && !aiCandidateMatchesEditor && !aiAssistant.candidateText
                ? alpha(theme.palette.common.white, 0.22)
                : undefined
          }}
        />
      ) : null}
      {canRevertLocalAIAccept ? (
        <Chip size="small" variant="outlined" color="info" label={t('templates.ai.statusChip.canRevert')} sx={{ color: 'common.white' }} />
      ) : null}
      {!isEdit && aiAssistant.candidateText ? (
        <Chip
          size="small"
          variant="outlined"
          label={t('templates.ai.statusChip.unsavedTemplate')}
          sx={{ color: 'common.white', borderColor: alpha(theme.palette.common.white, 0.22) }}
        />
      ) : null}
    </Stack>
  );

  const renderAIControlPanel = ({ compact = false, minimal = false } = {}) => {
    if (!isAIEnabled) return null;
    const dense = compact || minimal;

    return (
      <Box
        sx={{
          display: 'flex',
          flexWrap: 'wrap',
          alignItems: 'center',
          gap: dense ? 0.75 : 1,
          justifyContent: 'flex-end'
        }}
      >
        <Box
          sx={{
            display: 'inline-flex',
            alignItems: 'center',
            border: 1,
            borderColor: 'divider',
            borderRadius: 1,
            overflow: 'hidden',
            bgcolor: alpha(theme.palette.background.default, 0.4),
            flexShrink: 0
          }}
        >
          <Button
            variant={isEditMode ? 'contained' : 'text'}
            size="small"
            color={isEditMode ? 'primary' : 'inherit'}
            startIcon={<EditIcon fontSize="small" />}
            disabled={aiGenerating}
            onClick={() => setTemplateEditorMode('edit')}
            sx={{
              borderRadius: 0,
              minWidth: dense ? 78 : 86,
              px: 1.25,
              ...(isEditMode
                ? {}
                : {
                    color: 'text.secondary'
                  })
            }}
          >
            {t('templates.ai.mode.edit')}
          </Button>
          <Divider orientation="vertical" flexItem />
          <Button
            variant={showDiffReview ? 'contained' : 'text'}
            size="small"
            color={showDiffReview ? 'primary' : 'inherit'}
            startIcon={<CompareArrowsIcon />}
            disabled={!canSwitchToDiffMode || aiGenerating}
            onClick={() => setTemplateEditorMode('diff')}
            sx={{
              borderRadius: 0,
              minWidth: dense ? 78 : 86,
              px: 1.25,
              ...(showDiffReview
                ? {}
                : {
                    color: 'text.secondary'
                  })
            }}
          >
            {t('templates.ai.mode.diff')}
          </Button>
        </Box>
      </Box>
    );
  };

  const renderAIFloatingCommandBar = ({ fullscreen = false } = {}) => {
    if (!isAIEnabled) return null;
    return (
      <Box
        sx={{
          position: 'absolute',
          top: fullscreen ? 18 : 12,
          left: '50%',
          transform: 'translateX(-50%)',
          width: '100%',
          maxWidth: '100%',
          zIndex: 6,
          display: 'flex',
          justifyContent: 'center'
        }}
      >
        <Box
          role={aiCommandOpen ? undefined : 'button'}
          tabIndex={aiCommandOpen ? undefined : 0}
          aria-label={aiCommandOpen ? undefined : t('templates.ai.aria.expandCommand')}
          onClick={aiCommandOpen ? undefined : () => setAICommandOpen(true)}
          onKeyDown={
            aiCommandOpen
              ? undefined
              : (event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault();
                    setAICommandOpen(true);
                  }
                }
          }
          sx={{
            width: aiCommandOpen
              ? {
                  xs: 'calc(100% - 32px)',
                  sm: fullscreen ? 'min(560px, calc(100% - 84px))' : 'min(500px, calc(100% - 64px))'
                }
              : 38,
            minHeight: 38,
            display: 'flex',
            alignItems: 'center',
            gap: 0.75,
            px: aiCommandOpen ? 0.75 : 0,
            py: aiCommandOpen ? 0.5 : 0,
            justifyContent: aiCommandOpen ? 'flex-start' : 'center',
            borderRadius: 999,
            border: 1,
            borderColor: alpha(aiPromptPrimaryLight, aiCommandOpen ? 0.34 : 0.42),
            bgcolor: alpha(aiCommandOpen ? aiPromptSurface : aiPromptCollapsedSurface, aiCommandOpen ? 0.82 : 0.72),
            boxShadow: `0 ${aiCommandOpen ? 10 : 8}px ${aiCommandOpen ? 26 : 20}px ${alpha(theme.palette.common.black, aiCommandOpen ? 0.24 : 0.18)}`,
            backdropFilter: 'blur(10px)',
            overflow: 'hidden',
            cursor: aiCommandOpen ? 'default' : 'pointer',
            transition: theme.transitions.create(
              ['width', 'min-height', 'padding', 'border-radius', 'background-color', 'border-color', 'box-shadow'],
              {
                duration: theme.transitions.duration.shorter,
                easing: theme.transitions.easing.easeOut
              }
            ),
            '&:hover': aiCommandOpen
              ? undefined
              : {
                  borderColor: alpha(aiPromptPrimaryMain, 0.62),
                  bgcolor: alpha(aiPromptSurface, 0.9),
                  boxShadow: `0 12px 28px ${alpha(aiPromptPrimaryMain, 0.18)}`
                },
            '&:focus-visible': {
              outline: `2px solid ${alpha(aiPromptPrimaryLight, 0.68)}`,
              outlineOffset: 3
            }
          }}
        >
          <IconButton
            component={aiCommandOpen ? 'button' : 'div'}
            size="small"
            aria-label={aiCommandOpen ? t('templates.ai.aria.collapseCommand') : undefined}
            onClick={
              aiCommandOpen
                ? (e) => {
                    e.stopPropagation();
                    setAICommandOpen(false);
                  }
                : undefined
            }
            disabled={aiCommandOpen && aiGenerating}
            disableRipple={!aiCommandOpen}
            sx={{
              width: aiCommandOpen ? 32 : 36,
              height: aiCommandOpen ? 32 : 36,
              flexShrink: 0,
              color: aiCommandOpen ? alpha(theme.palette.common.white, aiGenerating ? 0.48 : 0.94) : alpha(aiPromptPrimaryLight, 0.95),
              transition: theme.transitions.create(['width', 'height', 'color', 'background-color']),
              '&:hover': aiCommandOpen
                ? {
                    bgcolor: alpha(theme.palette.common.white, 0.08)
                  }
                : {
                    bgcolor: 'transparent'
                  },
              '&.Mui-disabled': {
                color: alpha(theme.palette.common.white, 0.42)
              }
            }}
          >
            <AutoAwesomeIcon fontSize="small" />
          </IconButton>
          {aiCommandOpen ? (
            <>
              <TextField
                fullWidth
                size="small"
                value={aiPrompt}
                onChange={(e) => setAIPrompt(e.target.value)}
                disabled={aiGenerating}
                placeholder={t('templates.ai.promptPlaceholder')}
                inputProps={{ 'aria-label': t('templates.ai.promptAria') }}
                sx={{
                  minWidth: 0,
                  '& .MuiOutlinedInput-root': {
                    bgcolor: 'transparent',
                    color: alpha(theme.palette.common.white, 0.96),
                    height: 34,
                    pr: 0.25,
                    borderRadius: 999,
                    '&.Mui-disabled': {
                      color: alpha(theme.palette.common.white, 0.72),
                      WebkitTextFillColor: alpha(theme.palette.common.white, 0.72)
                    },
                    '& fieldset': {
                      borderColor: 'transparent'
                    },
                    '&:hover fieldset': {
                      borderColor: 'transparent'
                    },
                    '&.Mui-focused fieldset': {
                      borderColor: alpha(theme.palette.primary.main, 0.6)
                    }
                  },
                  '& .MuiInputBase-input': {
                    color: alpha(theme.palette.common.white, 0.96)
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: alpha(theme.palette.common.white, 0.72)
                  },
                  '& .MuiInputBase-input::placeholder': {
                    color: alpha(theme.palette.common.white, 0.64),
                    opacity: 1
                  }
                }}
              />
              <Button
                variant="contained"
                size="small"
                startIcon={aiGenerating ? <CircularProgress size={16} sx={{ color: 'common.white' }} /> : <AutoAwesomeIcon />}
                disabled={aiGenerating}
                onClick={handleGenerateWithAI}
                sx={{
                  flexShrink: 0,
                  minWidth: 92,
                  borderRadius: 999,
                  color: 'common.white',
                  boxShadow: 'none',
                  '&.Mui-disabled': {
                    color: 'common.white',
                    bgcolor: alpha(theme.palette.primary.main, 0.5)
                  }
                }}
              >
                {aiGenerating ? t('templates.ai.generating') : t('templates.ai.generate')}
              </Button>
              <IconButton
                size="small"
                disabled={!canAcceptAICandidateLocally || aiGenerating}
                onClick={handleAcceptAICandidateLocally}
                sx={{
                  flexShrink: 0,
                  borderRadius: 1,
                  bgcolor: alpha(theme.palette.common.white, 0.06),
                  color:
                    canAcceptAICandidateLocally && !aiGenerating
                      ? alpha(theme.palette.common.white, 0.96)
                      : alpha(theme.palette.common.white, 0.42),
                  '&.Mui-disabled': {
                    bgcolor: alpha(theme.palette.common.white, 0.04),
                    color: alpha(theme.palette.common.white, 0.34)
                  }
                }}
              >
                <CheckIcon fontSize="small" />
              </IconButton>
              {isEditMode ? (
                <IconButton
                  size="small"
                  disabled={!canRevertLocalAIAccept || aiGenerating}
                  onClick={handleRevertLastLocalAIAccept}
                  sx={{
                    flexShrink: 0,
                    borderRadius: 1,
                    bgcolor: alpha(theme.palette.common.white, 0.06),
                    color:
                      canRevertLocalAIAccept && !aiGenerating
                        ? alpha(theme.palette.common.white, 0.92)
                        : alpha(theme.palette.common.white, 0.4),
                    '&.Mui-disabled': {
                      bgcolor: alpha(theme.palette.common.white, 0.04),
                      color: alpha(theme.palette.common.white, 0.32)
                    }
                  }}
                >
                  <UndoIcon fontSize="small" />
                </IconButton>
              ) : null}
              {isAISetupIssue ? (
                <Button
                  size="small"
                  variant="text"
                  disabled={aiGenerating}
                  onClick={() => navigate('/settings?tab=ai')}
                  sx={{
                    flexShrink: 0,
                    minWidth: 'auto',
                    px: 0.75,
                    color: alpha(theme.palette.common.white, 0.92),
                    textDecoration: 'underline',
                    textUnderlineOffset: '2px',
                    '&.Mui-disabled': {
                      color: alpha(theme.palette.common.white, 0.5)
                    }
                  }}
                >
                  {t('templates.ai.goSettings')}
                </Button>
              ) : null}
            </>
          ) : null}
        </Box>
      </Box>
    );
  };

  const renderTemplateEditor = ({ fullscreen = false } = {}) => (
    <Box
      className="template-ai-editor-shell"
      sx={{
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        minHeight: fullscreen ? 0 : 350,
        flex: fullscreen ? 1 : '0 0 auto',
        ...(fullscreen
          ? {
              height: '100%',
              borderRadius: 1,
              overflow: 'hidden'
            }
          : null),
        '& .monaco-editor, & .monaco-diff-editor': {
          '--vscode-editorGutter-addedBackground': theme.palette.success.main,
          '--vscode-editorGutter-modifiedBackground': theme.palette.primary.main,
          '--vscode-editorGutter-deletedBackground': theme.palette.warning.main,
          '--vscode-diffEditor-insertedTextBackground': alpha(theme.palette.success.main, 0.2),
          '--vscode-diffEditor-removedTextBackground': alpha(theme.palette.warning.main, 0.16),
          '--vscode-diffEditor-insertedLineBackground': alpha(theme.palette.success.main, 0.08),
          '--vscode-diffEditor-removedLineBackground': alpha(theme.palette.warning.main, 0.08)
        }
      }}
    >
      {renderAIFloatingCommandBar({ fullscreen })}
      {converting && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            bgcolor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 10,
            borderRadius: 1
          }}
        >
          <Stack alignItems="center" spacing={1}>
            <CircularProgress />
            <Typography color="white">{t('templates.messages.converting')}</Typography>
          </Stack>
        </Box>
      )}
      {showDiffReview ? (
        <DiffEditor
          height={fullscreen ? '100%' : '350px'}
          language={formData.category === 'surge' ? 'ini' : 'yaml'}
          original={aiAssistant.sourceText || ''}
          modified={aiAssistant.candidateText || ''}
          theme="template-ai-editor"
          beforeMount={configureTemplateMonacoTheme}
          options={{
            renderSideBySide: true,
            readOnly: true,
            originalEditable: false,
            minimap: { enabled: !matchDownMd },
            fontSize: matchDownMd ? 12 : 14,
            wordWrap: 'on',
            contextmenu: true,
            automaticLayout: true,
            scrollBeyondLastLine: false,
            lineNumbers: matchDownMd ? 'off' : 'on',
            renderOverviewRuler: !matchDownMd,
            diffWordWrap: 'on'
          }}
        />
      ) : (
        <Editor
          height={fullscreen ? '100%' : '350px'}
          language={formData.category === 'surge' ? 'ini' : 'yaml'}
          value={formData.text}
          onChange={(value) => {
            setFormData({ ...formData, text: value || '' });
          }}
          theme="template-ai-editor"
          beforeMount={configureTemplateMonacoTheme}
          options={{
            minimap: { enabled: !matchDownMd },
            fontSize: matchDownMd ? 12 : 14,
            readOnly: converting,
            wordWrap: 'on',
            contextmenu: true,
            selectOnLineNumbers: true,
            automaticLayout: true,
            scrollBeyondLastLine: false,
            lineNumbers: matchDownMd ? 'off' : 'on'
          }}
        />
      )}
      {isAIEnabled && aiCommandOpen ? (
        <Box
          sx={{
            position: 'absolute',
            right: { xs: 24, sm: 32 },
            bottom: 16,
            maxWidth: { xs: 'calc(100% - 48px)', sm: 380 },
            px: 1.25,
            py: 0.75,
            borderRadius: 1,
            bgcolor: alpha(theme.palette.grey[900], 0.76),
            backdropFilter: 'blur(8px)',
            border: 1,
            borderColor: alpha(theme.palette.common.white, 0.12),
            boxShadow: `0 8px 24px ${alpha(theme.palette.common.black, 0.22)}`,
            zIndex: 5,
            pointerEvents: 'none'
          }}
        >
          <Stack spacing={0.75} sx={{ minWidth: 0 }}>
            {aiStateChips}
            <Typography
              variant="caption"
              sx={{
                color: aiCandidateMatchesEditor ? 'common.white' : isAISetupIssue ? alpha(theme.palette.common.white, 0.94) : aiStatusColor,
                display: 'block',
                lineHeight: 1.45,
                textShadow: aiCandidateMatchesEditor ? `0 1px 2px ${alpha(theme.palette.common.black, 0.45)}` : 'none'
              }}
            >
              {isAISetupIssue ? aiFriendlyGenerationError : aiStatusText}
            </Typography>
            {isAISetupIssue ? (
              <Typography variant="caption" sx={{ color: alpha(theme.palette.common.white, 0.76), display: 'block', lineHeight: 1.4 }}>
                {aiSetupGuidanceText}
              </Typography>
            ) : null}
            {aiUsageItems.length > 0 ? (
              <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
                {aiUsageItems.map((item) => (
                  <Chip
                    key={item.key}
                    size="small"
                    variant="outlined"
                    label={`${item.label} ${item.value}`}
                    sx={{
                      color: alpha(theme.palette.common.white, 0.92),
                      borderColor: alpha(theme.palette.common.white, 0.18),
                      bgcolor: alpha(theme.palette.common.white, 0.04),
                      '& .MuiChip-label': {
                        px: 1,
                        fontWeight: 500
                      }
                    }}
                  />
                ))}
              </Stack>
            ) : null}
          </Stack>
        </Box>
      ) : !isAIEnabled ? (
        <Box
          sx={{
            position: 'absolute',
            right: { xs: 14, sm: 18 },
            bottom: 16,
            zIndex: 5,
            display: 'flex',
            justifyContent: 'flex-end',
            maxWidth: { xs: 'calc(100% - 28px)', sm: 320 },
            pointerEvents: 'auto'
          }}
        >
          <Box
            component="button"
            type="button"
            aria-label={aiDisabledPromptOpen ? t('templates.ai.aria.collapseDisabledPrompt') : t('templates.ai.aria.viewDisabledPrompt')}
            onClick={() => setAIDisabledPromptOpen((open) => !open)}
            sx={{
              width: aiDisabledPromptOpen ? { xs: 248, sm: 292 } : 38,
              minHeight: aiCommandOpen ? 44 : 38,
              p: 0,
              border: 1,
              borderColor: alpha(aiPromptPrimaryLight, aiDisabledPromptOpen ? 0.46 : 0.34),
              borderRadius: 999,
              bgcolor: aiDisabledPromptOpen ? alpha(aiPromptSurface, 0.86) : alpha(aiPromptCollapsedSurface, 0.68),
              color: alpha(theme.palette.common.white, 0.92),
              backdropFilter: 'blur(10px)',
              boxShadow: `0 10px 24px ${alpha(theme.palette.common.black, aiDisabledPromptOpen ? 0.28 : 0.18)}`,
              cursor: 'pointer',
              overflow: 'hidden',
              display: 'flex',
              alignItems: 'center',
              justifyContent: aiDisabledPromptOpen ? 'flex-start' : 'center',
              transition: theme.transitions.create(['width', 'background-color', 'border-color', 'box-shadow', 'transform'], {
                duration: theme.transitions.duration.shorter,
                easing: theme.transitions.easing.easeOut
              }),
              animation: aiDisabledPromptOpen ? 'none' : 'template-ai-disabled-pulse 2.8s ease-in-out infinite',
              '@keyframes template-ai-disabled-pulse': {
                '0%, 100%': { transform: 'scale(1)' },
                '50%': { transform: 'scale(1.06)' }
              },
              '&:hover': {
                borderColor: alpha(aiPromptPrimaryMain, 0.58),
                bgcolor: alpha(aiPromptSurface, 0.9),
                boxShadow: `0 12px 28px ${alpha(aiPromptPrimaryMain, 0.18)}`
              }
            }}
          >
            <Box
              sx={{
                width: 36,
                height: 36,
                borderRadius: '50%',
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
                color: alpha(aiPromptPrimaryLight, 0.95)
              }}
            >
              <AutoAwesomeIcon fontSize="small" />
            </Box>
            <Stack
              direction="row"
              spacing={1}
              alignItems="center"
              sx={{
                minWidth: 0,
                pr: 1.25,
                opacity: aiDisabledPromptOpen ? 1 : 0,
                transform: aiDisabledPromptOpen ? 'translateX(0)' : 'translateX(8px)',
                transition: theme.transitions.create(['opacity', 'transform'], { duration: theme.transitions.duration.shorter }),
                whiteSpace: 'nowrap'
              }}
            >
              <Typography variant="caption" sx={{ color: alpha(theme.palette.common.white, 0.82) }}>
                {t('templates.ai.disabled')}
              </Typography>
              <Typography
                component="span"
                variant="caption"
                onClick={(event) => {
                  event.stopPropagation();
                  navigate('/settings?tab=ai');
                }}
                sx={{ color: aiPromptPrimaryLight, fontWeight: 700, textDecoration: 'underline', textUnderlineOffset: 3 }}
              >
                {t('templates.ai.goSettingsShort')}
              </Typography>
            </Stack>
          </Box>
        </Box>
      ) : null}
    </Box>
  );

  const getCategoryChipSx = (category) => {
    const semanticColor = category === 'surge' ? palette.secondary : palette.primary;

    return {
      bgcolor: withAlpha(semanticColor.main, isDark ? 0.12 : 0.08),
      color: isDark ? semanticColor.main : semanticColor.dark,
      borderColor: withAlpha(semanticColor.main, isDark ? 0.28 : 0.22),
      borderWidth: 1,
      borderStyle: 'solid',
      fontWeight: 600
    };
  };

  return (
    <MainCard
      title={t('templates.title')}
      secondary={
        matchDownMd ? (
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
            {t('common.add')}
          </Button>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="outlined" size="small" onClick={() => handleOpenBaseTemplate('clash')}>
              {t('templates.baseTemplate.button', { category: 'Clash' })}
            </Button>
            <Button variant="outlined" size="small" color="secondary" onClick={() => handleOpenBaseTemplate('surge')}>
              {t('templates.baseTemplate.button', { category: 'Surge' })}
            </Button>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              {t('templates.actions.addTemplate')}
            </Button>
            <IconButton onClick={handleRefresh} disabled={loading}>
              <RefreshIcon
                sx={
                  loading
                    ? {
                        animation: 'spin 1s linear infinite',
                        '@keyframes spin': { from: { transform: 'rotate(0deg)' }, to: { transform: 'rotate(360deg)' } }
                      }
                    : {}
                }
              />
            </IconButton>
          </Stack>
        )
      }
    >
      {matchDownMd && (
        <Stack direction="row" justifyContent="flex-end" sx={{ mb: 2 }}>
          <IconButton onClick={handleRefresh} disabled={loading} size="small">
            <RefreshIcon
              sx={
                loading
                  ? {
                      animation: 'spin 1s linear infinite',
                      '@keyframes spin': { from: { transform: 'rotate(0deg)' }, to: { transform: 'rotate(360deg)' } }
                    }
                  : {}
              }
            />
          </IconButton>
        </Stack>
      )}

      {matchDownMd ? (
        <Stack spacing={2}>
          {templates.map((template) => (
            <MainCard key={template.file} content={false} border shadow={theme.shadows[1]}>
              <Box p={2}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1}>
                  <Chip label={template.file} color="success" variant="outlined" />
                  <Typography variant="caption" color="textSecondary">
                    {template.create_date || '-'}
                  </Typography>
                </Stack>

                <Divider sx={{ my: 1 }} />

                <Stack direction="row" justifyContent="flex-end" spacing={1}>
                  <IconButton size="small" onClick={() => handleEdit(template)}>
                    <EditIcon fontSize="small" />
                  </IconButton>
                  <IconButton size="small" color="error" onClick={() => handleDelete(template)}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Stack>
              </Box>
            </MainCard>
          ))}
        </Stack>
      ) : (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>{t('templates.fields.filename')}</TableCell>
                <TableCell>{t('templates.fields.category')}</TableCell>
                <TableCell>{t('templates.fields.ruleSource')}</TableCell>
                <TableCell>{t('templates.fields.createdAt')}</TableCell>
                <TableCell align="right">{t('templates.fields.actions')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {templates.map((template) => (
                <TableRow key={template.file} hover>
                  <TableCell>
                    <Chip label={template.file} color="success" variant="outlined" size="small" />
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={template.category === 'surge' ? 'Surge' : 'Clash'}
                      size="small"
                      sx={getCategoryChipSx(template.category)}
                    />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {template.ruleSource || '-'}
                    </Typography>
                  </TableCell>
                  <TableCell>{template.create_date || '-'}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" onClick={() => handleEdit(template)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => handleDelete(template)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Pagination
        page={page}
        pageSize={rowsPerPage}
        totalItems={totalItems}
        onPageChange={(_, newPage) => {
          setPage(newPage);
          fetchTemplates(newPage, rowsPerPage);
        }}
        onPageSizeChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('templates_rowsPerPage', newValue);
          setPage(0);
          fetchTemplates(0, newValue);
        }}
        pageSizeOptions={[10, 20, 50, 100]}
      />

      <Dialog
        open={dialogOpen}
        onClose={handleCloseDialog}
        maxWidth={editorFullscreen ? false : 'lg'}
        fullWidth
        fullScreen={editorFullscreen}
        PaperProps={{
          sx: editorFullscreen
            ? {
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
                height: '100vh',
                maxHeight: '100vh',
                m: 0
              }
            : undefined
        }}
      >
        <DialogTitle
          sx={
            editorFullscreen
              ? {
                  flexShrink: 0,
                  pb: 1,
                  display: 'flex',
                  flexDirection: { xs: 'column', md: 'row' },
                  alignItems: { xs: 'stretch', md: 'center' },
                  justifyContent: 'space-between',
                  gap: 1.5
                }
              : undefined
          }
        >
          <Stack spacing={0.5}>
            <Typography variant="h4">{isEdit ? t('templates.dialog.editTitle') : t('templates.dialog.addTitle')}</Typography>
          </Stack>
          {editorFullscreen && (
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap justifyContent="flex-end" alignItems="center">
              {renderAIControlPanel({ compact: true })}
              <Button variant="outlined" size="small" startIcon={<FullscreenExitIcon />} onClick={() => setEditorFullscreen(false)}>
                {t('templates.actions.exitFullscreen')}
              </Button>
            </Stack>
          )}
        </DialogTitle>
        <DialogContent
          sx={
            editorFullscreen
              ? {
                  display: 'flex',
                  flexDirection: 'column',
                  flex: 1,
                  minHeight: 0,
                  overflow: 'hidden',
                  overflowX: 'hidden',
                  pt: 1,
                  pb: 1.5
                }
              : undefined
          }
        >
          <Stack
            spacing={1.5}
            sx={
              editorFullscreen
                ? {
                    flex: 1,
                    minHeight: 0
                  }
                : { mt: 0.5 }
            }
          >
            {!editorFullscreen && (
              <>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                  <TextField
                    fullWidth
                    label={t('templates.fields.filename')}
                    value={formData.filename}
                    onChange={(e) => setFormData({ ...formData, filename: e.target.value })}
                    placeholder={t('templates.placeholders.filename')}
                    sx={outlinedLabelFixSx}
                  />
                  <FormControl sx={{ minWidth: 120, ...outlinedLabelFixSx }}>
                    <InputLabel>{t('templates.fields.category')}</InputLabel>
                    <Select
                      value={formData.category}
                      label={t('templates.fields.category')}
                      onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    >
                      <MenuItem value="clash">Clash</MenuItem>
                      <MenuItem value="surge">Surge</MenuItem>
                    </Select>
                  </FormControl>
                </Stack>
                <Autocomplete
                  freeSolo
                  options={aclPresets}
                  sx={outlinedLabelFixSx}
                  getOptionLabel={(option) => {
                    if (typeof option === 'string') return option;
                    return option.label || option.url || '';
                  }}
                  isOptionEqualToValue={(option, value) => {
                    if (typeof value === 'string') {
                      return option.url === value;
                    }
                    return option.url === value?.url;
                  }}
                  value={aclPresets.find((preset) => preset.url === formData.ruleSource) || formData.ruleSource}
                  onChange={(_, newValue) => {
                    if (typeof newValue === 'string') {
                      setFormData({ ...formData, ruleSource: newValue });
                    } else if (newValue && newValue.url) {
                      setFormData({ ...formData, ruleSource: newValue.url });
                    } else {
                      setFormData({ ...formData, ruleSource: '' });
                    }
                  }}
                  onInputChange={(_, newInputValue) => {
                    setFormData({ ...formData, ruleSource: newInputValue });
                  }}
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      label={t('templates.fields.remoteRuleSource')}
                      placeholder={t('templates.placeholders.ruleSource')}
                      helperText={t('templates.helpers.ruleSource')}
                    />
                  )}
                  renderOption={(props, option) => (
                    <li {...props} key={option.name}>
                      <Stack>
                        <Typography variant="body2">{option.label}</Typography>
                        <Typography variant="caption" color="textSecondary" sx={{ fontSize: '0.7rem' }}>
                          {option.url}
                        </Typography>
                      </Stack>
                    </li>
                  )}
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={useProxy}
                      onChange={(e) => {
                        const checked = e.target.checked;
                        setUseProxy(checked);
                        if (checked) {
                          fetchProxyNodes();
                        }
                      }}
                    />
                  }
                  label={t('templates.fields.useProxy')}
                />
                {useProxy && (
                  <Box>
                    <SearchableNodeSelect
                      nodes={proxyNodeOptions}
                      loading={loadingProxyNodes}
                      value={
                        proxyNodeOptions.find((n) => n.Link === proxyLink) || (proxyLink ? { Link: proxyLink, Name: '', ID: 0 } : null)
                      }
                      onChange={(newValue) => setProxyLink(typeof newValue === 'string' ? newValue : newValue?.Link || '')}
                      displayField="Name"
                      valueField="Link"
                      label={t('templates.fields.proxyNode')}
                      placeholder={t('templates.placeholders.proxyNode')}
                      helperText={t('templates.helpers.proxyNode')}
                      freeSolo={true}
                      limit={50}
                    />
                  </Box>
                )}
                <FormControlLabel
                  control={
                    <Switch
                      checked={formData.enableIncludeAll}
                      onChange={(e) => setFormData({ ...formData, enableIncludeAll: e.target.checked })}
                    />
                  }
                  label={t('templates.fields.includeAll')}
                />
                <Typography variant="caption" color="textSecondary" component="div" sx={{ ml: 6, mt: -0.5, lineHeight: 1.6 }}>
                  {t('templates.helpers.includeAllOn')}
                </Typography>
                <Typography variant="caption" color="textSecondary" component="div" sx={{ ml: 6, lineHeight: 1.6 }}>
                  {t('templates.helpers.includeAllOff')}
                </Typography>
                <Stack direction="row" spacing={1}>
                  <Button
                    variant="outlined"
                    startIcon={converting ? <CircularProgress size={18} /> : <TransformIcon />}
                    disabled={!formData.ruleSource || converting}
                    onClick={() => handleConvertTemplate(false)}
                  >
                    {t('templates.actions.convertRules')}
                  </Button>
                  <Button
                    variant="outlined"
                    startIcon={converting ? <CircularProgress size={18} /> : <UnfoldMoreIcon />}
                    disabled={!formData.ruleSource || converting}
                    onClick={() => handleConvertTemplate(true)}
                  >
                    {t('templates.actions.convertRulesExpand')}
                  </Button>
                  <Button
                    variant="outlined"
                    color="error"
                    disabled={!formData.text || converting}
                    onClick={() => {
                      openConfirm(t('templates.confirm.clearTitle'), t('templates.confirm.clearContent'), () => {
                        setFormData({ ...formData, text: '' });
                        showMessage(t('templates.messages.cleared'));
                      });
                    }}
                  >
                    {t('templates.actions.clearContent')}
                  </Button>
                </Stack>
                <Box
                  sx={{
                    ...aiWorkspacePanelSx,
                    p: { xs: 1, md: 1.25 },
                    display: 'flex',
                    flexDirection: { xs: 'column', sm: 'row' },
                    alignItems: { xs: 'stretch', sm: 'center' },
                    justifyContent: 'flex-end',
                    gap: 1.5
                  }}
                >
                  {renderAIControlPanel({ minimal: true })}
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={editorFullscreen ? <FullscreenExitIcon /> : <FullscreenIcon />}
                    onClick={() => setEditorFullscreen((prev) => !prev)}
                    sx={{ flexShrink: 0 }}
                  >
                    {editorFullscreen ? t('templates.actions.exitFullscreen') : t('templates.actions.fullscreen')}
                  </Button>
                </Box>
              </>
            )}
            <Box
              sx={
                editorFullscreen
                  ? {
                      flex: 1,
                      minHeight: 0,
                      display: 'flex',
                      flexDirection: 'column'
                    }
                  : undefined
              }
            >
              {renderTemplateEditor({ fullscreen: editorFullscreen })}
            </Box>
          </Stack>
        </DialogContent>
        {!editorFullscreen && (
          <DialogActions>
            <Button onClick={handleCloseDialog}>{t('common.cancel')}</Button>
            <Button variant="contained" disabled={templateEditorMode === 'diff'} onClick={handleSubmit}>
              {t('common.confirm')}
            </Button>
          </DialogActions>
        )}
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>

      <Dialog
        open={confirmOpen}
        onClose={handleConfirmClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">{confirmInfo.title}</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description" sx={{ color: 'text.primary' }}>
            {confirmInfo.content}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleConfirmClose}>{t('common.cancel')}</Button>
          <Button onClick={handleConfirmAction} variant="contained" color="error" autoFocus>
            {t('common.confirm')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={errorDialog.open}
        onClose={() => setErrorDialog({ ...errorDialog, open: false })}
        aria-labelledby="error-dialog-title"
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle id="error-dialog-title" sx={{ color: 'error.main' }}>
          ⚠️ {errorDialog.title}
        </DialogTitle>
        <DialogContent>
          <Alert severity="error" sx={{ mt: 1 }}>
            {errorDialog.message}
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button variant="contained" onClick={() => setErrorDialog({ ...errorDialog, open: false })} autoFocus>
            {t('common.close')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={usageDialog.open}
        onClose={() => setUsageDialog({ ...usageDialog, open: false })}
        aria-labelledby="template-usage-dialog-title"
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle id="template-usage-dialog-title">⚠️ {usageDialog.title}</DialogTitle>
        <DialogContent>
          <Alert
            severity="warning"
            variant="outlined"
            sx={{
              mt: 1,
              alignItems: 'flex-start',
              backgroundColor: alpha(theme.palette.warning.main, 0.08),
              borderColor: alpha(theme.palette.warning.main, 0.28),
              color: 'text.primary',
              '& .MuiAlert-icon': {
                color: 'warning.dark',
                mt: '2px'
              },
              '& .MuiAlert-message': {
                width: '100%'
              }
            }}
          >
            {usageDialog.message}
          </Alert>
          {usageDialog.subscriptions?.length > 0 && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                {t('templates.usage.usedSubscriptions')}
              </Typography>
              <Stack spacing={1}>
                {usageDialog.subscriptions.map((subscriptionName) => (
                  <Chip key={subscriptionName} label={subscriptionName} color="warning" variant="outlined" sx={{ width: 'fit-content' }} />
                ))}
              </Stack>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setUsageDialog({ ...usageDialog, open: false, subscriptions: [], action: null })}>
            {t('common.cancel')}
          </Button>
          <Button
            variant="contained"
            color="error"
            onClick={async () => {
              const action = usageDialog.action;
              setUsageDialog({ open: false, title: '', message: '', subscriptions: [], action: null });
              if (action) {
                await action();
              }
            }}
            autoFocus
          >
            {t('templates.actions.continueDelete')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={baseTemplateDialogOpen} onClose={() => setBaseTemplateDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>{t('templates.baseTemplate.title', { category: baseTemplateCategory === 'clash' ? 'Clash' : 'Surge' })}</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
            {t('templates.baseTemplate.description')}
          </Typography>
          {baseTemplateLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: 400 }}>
              <CircularProgress />
            </Box>
          ) : (
            <Editor
              height="400px"
              language={baseTemplateCategory === 'surge' ? 'ini' : 'yaml'}
              value={baseTemplateContent}
              onChange={(value) => setBaseTemplateContent(value || '')}
              theme="vs-dark"
              options={{
                minimap: { enabled: !matchDownMd },
                fontSize: matchDownMd ? 12 : 14,
                readOnly: baseTemplateSaving,
                wordWrap: 'on',
                contextmenu: true,
                selectOnLineNumbers: true,
                automaticLayout: true,
                scrollBeyondLastLine: false,
                lineNumbers: matchDownMd ? 'off' : 'on'
              }}
            />
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setBaseTemplateDialogOpen(false)}>{t('common.cancel')}</Button>
          <Button
            variant="contained"
            onClick={handleSaveBaseTemplate}
            disabled={baseTemplateLoading || baseTemplateSaving}
            startIcon={baseTemplateSaving ? <CircularProgress size={18} /> : null}
          >
            {t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
