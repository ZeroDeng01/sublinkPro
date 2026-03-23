import { useState, useEffect, useRef } from 'react';
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
import InputAdornment from '@mui/material/InputAdornment';

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
import { getBaseTemplates, updateBaseTemplate } from 'api/settings';
import { getNodes } from 'api/nodes';

// Monaco Editor
import Editor, { DiffEditor } from '@monaco-editor/react';

// ==============================|| 模板管理 ||============================== //

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

const buildTemplateAIUsageItems = (usage) => {
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
    inputTokens !== null ? { key: 'input', label: '输入', value: inputTokens } : null,
    outputTokens !== null ? { key: 'output', label: '输出', value: outputTokens } : null,
    cacheTokens !== null ? { key: 'cache', label: '缓存', value: cacheTokens } : null
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
  const theme = useTheme();
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
  const [errorDialog, setErrorDialog] = useState({ open: false, title: '', message: '' });
  const [usageDialog, setUsageDialog] = useState({ open: false, title: '', message: '', subscriptions: [], action: null });
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('templates_rowsPerPage');
    return saved ? parseInt(saved, 10) : 10;
  });
  const [totalItems, setTotalItems] = useState(0);

  // 确认对话框
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({
    title: '',
    content: '',
    action: null
  });

  // 基础模板编辑对话框
  const [baseTemplateDialogOpen, setBaseTemplateDialogOpen] = useState(false);
  const [baseTemplateCategory, setBaseTemplateCategory] = useState('clash');
  const [baseTemplateContent, setBaseTemplateContent] = useState('');
  const [baseTemplateLoading, setBaseTemplateLoading] = useState(false);
  const [baseTemplateSaving, setBaseTemplateSaving] = useState(false);

  // 代理设置
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
      // 处理分页响应
      if (response.data && response.data.items !== undefined) {
        setTemplates(response.data.items || []);
        setTotalItems(response.data.total || 0);
      } else {
        // 向后兼容：老格式直接返回数组
        setTemplates(response.data || []);
        setTotalItems((response.data || []).length);
      }
    } catch (error) {
      console.log(error);
      showMessage(error.message || '获取模板列表失败', 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    fetchTemplates(page, rowsPerPage);
  };

  useEffect(() => {
    fetchTemplates(0, rowsPerPage);
    // 获取 ACL4SSR 预设列表
    getACL4SSRPresets()
      .then((res) => {
        if (res.data) {
          setAclPresets(res.data);
        }
      })
      .catch((err) => console.log('获取预设列表失败:', err));
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
    // 从模板数据加载代理设置
    setUseProxy(template.useProxy || false);
    setProxyLink(template.proxyLink || '');
    // 如果之前保存了使用代理，预加载节点列表
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
      showMessage(error.message || '获取模板使用情况失败', 'error');
      return;
    }

    const deleteAction = async () => {
      try {
        await deleteTemplate({ filename: template.file });
        showMessage('删除成功');
        fetchTemplates(page, rowsPerPage);
      } catch (error) {
        console.log(error);
        showMessage(error.message || '删除失败', 'error');
      }
    };

    if (usedSubscriptions.length > 0) {
      setUsageDialog({
        open: true,
        title: '模板正在被订阅使用',
        message: `模板 "${template.file}" 当前正被以下订阅使用，删除后这些订阅可能受到影响，是否继续删除？`,
        subscriptions: usedSubscriptions,
        action: deleteAction
      });
      return;
    }

    openConfirm('删除模板', `确定要删除模板 "${template.file}" 吗？`, deleteAction);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditorFullscreen(false);
    resetTemplateAIAssistant();
  };

  useEffect(() => () => abortAIGeneration(), []);

  const handleGenerateWithAI = async () => {
    if (!aiPrompt.trim()) {
      showMessage('请输入 AI 指令', 'warning');
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
        showMessage('AI 草稿已生成，可切换并停留在并排对比模式');
      } else {
        showMessage('AI 生成完成，但未返回候选内容', 'warning');
      }
    } catch (error) {
      if (error.name === 'AbortError') {
        return;
      }

      const errorMessage = error.response?.data?.message || error.message || 'AI 生成失败';
      const friendlyErrorMessage =
        errorMessage.includes('当前用户未启用 AI 助手') || errorMessage.includes('AI 设置不完整，请先配置 Base URL、模型和 API Key')
          ? 'AI 助手当前不可用，请前往 /settings，在 个人设置 -> AI 助手 中完成配置。'
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
    if (templateEditorMode === 'diff' && !canSwitchToDiffMode) {
      setTemplateEditorMode('edit');
    }
  }, [templateEditorMode, canSwitchToDiffMode]);

  const handleAcceptAICandidateLocally = () => {
    if (!aiAssistant.candidateText) {
      showMessage('请先生成 AI 候选内容', 'warning');
      return;
    }

    if (aiCandidateMatchesEditor) {
      showMessage('当前编辑器内容已经与 AI 候选结果一致', 'info');
      return;
    }

    if (aiCandidateOutdated) {
      showMessage('模板内容或配置已变化，请重新生成或重新校验候选内容后再接受到编辑器', 'warning');
      return;
    }

    setAILocalAcceptSnapshot({ text: formData.text });
    setTemplateEditorMode('edit');
    setFormData((prev) => ({
      ...prev,
      text: aiAssistant.candidateText
    }));
    showMessage('AI 草稿已写入编辑器，可继续编辑或直接保存');
  };

  const handleRevertLastLocalAIAccept = () => {
    if (!aiLocalAcceptSnapshot) {
      showMessage('没有可回退的本地接受记录', 'warning');
      return;
    }

    setFormData((prev) => ({
      ...prev,
      text: aiLocalAcceptSnapshot.text || ''
    }));
    setTemplateEditorMode('edit');
    setAILocalAcceptSnapshot(null);
    showMessage('已恢复最近一次接受 AI 候选前的编辑器内容');
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
        showMessage(expand ? '规则生成/转换并展开成功' : '规则生成/转换成功');
      } else {
        setErrorDialog({
          open: true,
          title: '规则生成/转换失败',
          message: res.msg || '生成/转换过程中发生错误'
        });
      }
    } catch (error) {
      console.error(error);
      const errorMsg = error.response?.data?.msg || error.message || '规则生成/转换失败';
      setErrorDialog({
        open: true,
        title: '规则生成/转换失败',
        message: errorMsg
      });
    } finally {
      setConverting(false);
    }
  };

  const handleSubmit = async () => {
    try {
      if (templateEditorMode === 'diff') {
        showMessage('当前处于 AI 对比模式，请先返回编辑模式后再保存', 'warning');
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
        showMessage('更新成功');
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
        showMessage('添加成功');
      }
      setEditorFullscreen(false);
      setDialogOpen(false);
      resetTemplateAIAssistant();
      fetchTemplates(page, rowsPerPage);
    } catch (error) {
      console.log(error);
      showMessage(error.message || (isEdit ? '更新失败' : '添加失败'), 'error');
    }
  };

  // 打开基础模板编辑对话框
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
      showMessage(error.message || '获取基础模板失败', 'error');
    } finally {
      setBaseTemplateLoading(false);
    }
  };

  // 保存基础模板
  const handleSaveBaseTemplate = async () => {
    setBaseTemplateSaving(true);
    try {
      await updateBaseTemplate(baseTemplateCategory, baseTemplateContent);
      showMessage(`${baseTemplateCategory === 'clash' ? 'Clash' : 'Surge'} 基础模板保存成功`);
      setBaseTemplateDialogOpen(false);
    } catch (error) {
      console.error(error);
      showMessage(error.message || '保存基础模板失败', 'error');
    } finally {
      setBaseTemplateSaving(false);
    }
  };

  // 获取代理节点列表
  const fetchProxyNodes = async () => {
    setLoadingProxyNodes(true);
    try {
      const res = await getNodes({ minSpeed: 0.01, pageSize: 100 }); // 获取有速度的节点
      if (res.data) {
        const items = res.data.items || res.data || [];
        setProxyNodeOptions(items);
      }
    } catch (error) {
      console.error('获取代理节点失败:', error);
    } finally {
      setLoadingProxyNodes(false);
    }
  };

  const compactOutlinedFieldSx = {
    '& .MuiInputLabel-root': {
      px: 0.5,
      backgroundColor: 'background.paper',
      maxWidth: 'calc(100% - 24px)'
    },
    '& .MuiInputLabel-shrink': {
      maxWidth: 'calc(133% - 32px)'
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
    ? '正在基于当前编辑器内容生成草稿。'
    : aiGenerationError
      ? aiGenerationError
      : aiCandidateOutdated
        ? '当前内容或配置已变化，请重新生成新的草稿。'
        : !isEdit && aiAssistant.candidateText
          ? '当前模板尚未保存，可先应用到编辑器后再保存。'
          : showDiffReview
            ? '对比模式为只读，保存前请先返回编辑模式。'
            : aiCandidateMatchesEditor
              ? '当前编辑器已载入 AI 草稿。'
              : canRevertLocalAIAccept
                ? '已保留应用前快照，可在编辑模式下回退。'
                : aiAssistant.candidateText
                  ? '可对比、应用或继续编辑当前候选草稿。'
                  : '输入指令后生成候选草稿。';
  const aiStatusColor = aiGenerationError
    ? 'error.main'
    : aiCandidateOutdated
      ? 'warning.main'
      : showDiffReview
        ? alpha(theme.palette.common.white, 0.92)
        : aiCandidateMatchesEditor
          ? 'success.main'
          : alpha(theme.palette.common.white, 0.88);
  const isAISetupIssue =
    aiGenerationError.includes('当前用户未启用 AI 助手') || aiGenerationError.includes('AI 设置不完整，请先配置 Base URL、模型和 API Key');
  const aiSetupGuidanceText = isAISetupIssue ? '请前往 /settings，在 个人设置 -> AI 助手 中启用并完成配置。' : '';
  const aiFriendlyGenerationError = isAISetupIssue ? 'AI 助手当前不可用。' : aiGenerationError;
  const aiUsageItems = buildTemplateAIUsageItems(aiAssistant.usage);

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
        label={isEditMode ? '编辑' : '对比'}
        color={isEditMode ? 'primary' : 'default'}
        sx={{
          color: 'common.white',
          bgcolor: isEditMode ? undefined : alpha(theme.palette.common.white, 0.14),
          '& .MuiChip-label': {
            fontWeight: 600
          }
        }}
      />
      {aiGenerating ? <Chip size="small" variant="outlined" color="primary" label="生成中" sx={{ color: 'common.white' }} /> : null}
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
              ? '生成失败'
              : aiCandidateOutdated
                ? '草稿过期'
                : aiCandidateMatchesEditor
                  ? '已写入编辑器'
                  : aiAssistant.candidateText
                    ? '草稿可用'
                    : '未生成'
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
        <Chip size="small" variant="outlined" color="info" label="可回退上次接受" sx={{ color: 'common.white' }} />
      ) : null}
      {!isEdit && aiAssistant.candidateText ? (
        <Chip
          size="small"
          variant="outlined"
          label="未保存模板"
          sx={{ color: 'common.white', borderColor: alpha(theme.palette.common.white, 0.22) }}
        />
      ) : null}
    </Stack>
  );

  const renderAIControlPanel = ({ compact = false, minimal = false } = {}) => {
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
            编辑
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
            对比
          </Button>
        </Box>
      </Box>
    );
  };

  const renderAIFloatingCommandBar = ({ fullscreen = false } = {}) => (
    <Box
      sx={{
        position: 'absolute',
        top: fullscreen ? 18 : 12,
        left: '50%',
        transform: 'translateX(-50%)',
        width: {
          xs: 'calc(100% - 32px)',
          sm: fullscreen ? 'min(560px, calc(100% - 84px))' : 'min(500px, calc(100% - 64px))'
        },
        maxWidth: '100%',
        zIndex: 6,
        display: 'flex',
        justifyContent: 'center'
      }}
    >
      <Box
        sx={{
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          gap: 0.75,
          px: 0.75,
          py: 0.5,
          borderRadius: 1,
          border: 1,
          borderColor: alpha(theme.palette.common.white, 0.12),
          bgcolor: alpha(theme.palette.grey[900], 0.82),
          boxShadow: `0 10px 26px ${alpha(theme.palette.common.black, 0.24)}`,
          backdropFilter: 'blur(10px)'
        }}
      >
        <TextField
          fullWidth
          size="small"
          value={aiPrompt}
          onChange={(e) => setAIPrompt(e.target.value)}
          disabled={aiGenerating}
          placeholder="告诉 AI 要如何调整当前模板…"
          inputProps={{ 'aria-label': 'AI 指令' }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <AutoAwesomeIcon fontSize="small" sx={{ color: alpha(theme.palette.common.white, 0.92) }} />
              </InputAdornment>
            )
          }}
          sx={{
            '& .MuiOutlinedInput-root': {
              bgcolor: 'transparent',
              color: alpha(theme.palette.common.white, 0.96),
              height: 34,
              pr: 0.25,
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
              color: alpha(theme.palette.common.white, 0.96),
              paddingLeft: 2
            },
            '& .MuiInputBase-input.Mui-disabled': {
              WebkitTextFillColor: alpha(theme.palette.common.white, 0.72)
            },
            '& .MuiInputBase-input::placeholder': {
              color: alpha(theme.palette.common.white, 0.64),
              opacity: 1
            },
            '& .MuiInputAdornment-root': {
              mr: 0.75,
              color: alpha(theme.palette.common.white, 0.88)
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
            color: 'common.white',
            boxShadow: 'none',
            '&.Mui-disabled': {
              color: 'common.white',
              bgcolor: alpha(theme.palette.primary.main, 0.5)
            }
          }}
        >
          {aiGenerating ? '生成中' : '生成'}
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
                canRevertLocalAIAccept && !aiGenerating ? alpha(theme.palette.common.white, 0.92) : alpha(theme.palette.common.white, 0.4),
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
            onClick={() => navigate('/settings')}
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
            前往设置
          </Button>
        ) : null}
      </Box>
    </Box>
  );

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
            <Typography color="white">正在转换规则...</Typography>
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
    </Box>
  );

  return (
    <MainCard
      title="模板管理"
      secondary={
        matchDownMd ? (
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
            添加
          </Button>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="outlined" size="small" onClick={() => handleOpenBaseTemplate('clash')}>
              Clash 基础模板
            </Button>
            <Button variant="outlined" size="small" color="secondary" onClick={() => handleOpenBaseTemplate('surge')}>
              Surge 基础模板
            </Button>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              添加模板
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
                <TableCell>文件名</TableCell>
                <TableCell>类别</TableCell>
                <TableCell>规则源</TableCell>
                <TableCell>创建时间</TableCell>
                <TableCell align="right">操作</TableCell>
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
                      color={template.category === 'surge' ? 'secondary' : 'primary'}
                      size="small"
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

      {/* 添加/编辑对话框 */}
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
            <Typography variant="h4">{isEdit ? '编辑模板' : '添加模板'}</Typography>
            {editorFullscreen && (
              <Typography variant="body2" color="textSecondary">
                全屏模式同样使用编辑 / 对比双模式，避免额外的 AI 侧边工作区。
              </Typography>
            )}
          </Stack>
          {editorFullscreen && (
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap justifyContent="flex-end">
              <Button size="small" onClick={handleCloseDialog}>
                取消
              </Button>
              <Button variant="contained" size="small" disabled={templateEditorMode === 'diff'} onClick={handleSubmit}>
                保存
              </Button>
              <Button variant="outlined" size="small" startIcon={<FullscreenExitIcon />} onClick={() => setEditorFullscreen(false)}>
                退出全屏
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
            {editorFullscreen ? (
              <Stack spacing={1} sx={{ flexShrink: 0 }}>
                <Box
                  sx={{
                    px: 1,
                    py: 1.25,
                    border: 1,
                    borderColor: 'divider',
                    borderRadius: 1,
                    bgcolor: 'background.paper'
                  }}
                >
                  <Stack spacing={1.25}>
                    <Stack direction="row" spacing={1.25} useFlexGap flexWrap="wrap" alignItems={{ xs: 'stretch', lg: 'flex-start' }}>
                      <TextField
                        fullWidth
                        size="small"
                        label="文件名"
                        value={formData.filename}
                        onChange={(e) => setFormData({ ...formData, filename: e.target.value })}
                        placeholder="例如: clash.yaml"
                        InputLabelProps={{ shrink: true }}
                        sx={{
                          ...compactOutlinedFieldSx,
                          flex: { xs: '1 1 100%', sm: '1 1 240px', lg: '0 1 220px' },
                          minWidth: { xs: '100%', sm: 220, lg: 200 }
                        }}
                      />
                      <FormControl
                        size="small"
                        sx={{
                          minWidth: { xs: '100%', sm: 144, lg: 132 },
                          flex: { xs: '1 1 100%', sm: '0 1 160px', lg: '0 0 132px' }
                        }}
                      >
                        <InputLabel shrink sx={compactOutlinedFieldSx['& .MuiInputLabel-root']}>
                          类别
                        </InputLabel>
                        <Select
                          value={formData.category}
                          label="类别"
                          notched
                          onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                        >
                          <MenuItem value="clash">Clash</MenuItem>
                          <MenuItem value="surge">Surge</MenuItem>
                        </Select>
                      </FormControl>
                      <Autocomplete
                        freeSolo
                        options={aclPresets}
                        sx={{
                          flex: { xs: '1 1 100%', md: '999 1 320px' },
                          minWidth: { xs: '100%', md: 280, lg: 320 }
                        }}
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
                            size="small"
                            label="远程规则地址"
                            placeholder="输入 URL 或选择 ACL4SSR 预设"
                            InputLabelProps={{ ...params.InputLabelProps, shrink: true }}
                            sx={compactOutlinedFieldSx}
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
                    </Stack>
                    <Stack spacing={0.75}>
                      <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" alignItems="center">
                        <FormControlLabel
                          sx={{ mr: 0, flex: { xs: '1 1 100%', md: '0 1 auto' } }}
                          control={
                            <Switch
                              size="small"
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
                          label="使用代理下载远程规则"
                        />
                        <FormControlLabel
                          sx={{ mr: 0, flex: { xs: '1 1 100%', md: '0 1 auto' } }}
                          control={
                            <Switch
                              size="small"
                              checked={formData.enableIncludeAll}
                              onChange={(e) => setFormData({ ...formData, enableIncludeAll: e.target.checked })}
                            />
                          }
                          label="使用 Include-All 模式"
                        />
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={converting ? <CircularProgress size={16} /> : <TransformIcon />}
                          disabled={!formData.ruleSource || converting}
                          onClick={() => handleConvertTemplate(false)}
                        >
                          生成/转换
                        </Button>
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={converting ? <CircularProgress size={16} /> : <UnfoldMoreIcon />}
                          disabled={!formData.ruleSource || converting}
                          onClick={() => handleConvertTemplate(true)}
                        >
                          转换并展开
                        </Button>
                        <Button
                          variant="outlined"
                          size="small"
                          color="error"
                          disabled={!formData.text || converting}
                          onClick={() => {
                            openConfirm('清空内容', '确定要清空编辑器中的所有内容吗？', () => {
                              setFormData({ ...formData, text: '' });
                              showMessage('已清空内容');
                            });
                          }}
                        >
                          清空
                        </Button>
                      </Stack>
                      <Typography variant="caption" color="text.secondary">
                        全屏模式保留紧凑配置栏，并在同一编辑区内切换编辑模式与对比模式。
                      </Typography>
                    </Stack>
                    {useProxy && (
                      <SearchableNodeSelect
                        nodes={proxyNodeOptions}
                        loading={loadingProxyNodes}
                        value={
                          proxyNodeOptions.find((n) => n.Link === proxyLink) || (proxyLink ? { Link: proxyLink, Name: '', ID: 0 } : null)
                        }
                        onChange={(newValue) => setProxyLink(newValue?.Link || '')}
                        displayField="Name"
                        valueField="Link"
                        label="选择代理节点"
                        placeholder="留空则自动选择最佳节点"
                        helperText="如果未选择具体代理，系统将自动选择最佳节点"
                        freeSolo={true}
                        limit={50}
                      />
                    )}
                  </Stack>
                </Box>

                <Box
                  sx={{
                    ...aiWorkspacePanelSx,
                    p: { xs: 1, md: 1.25 },
                    flexShrink: 0
                  }}
                >
                  {renderAIControlPanel({ compact: true })}
                </Box>
              </Stack>
            ) : (
              <>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                  <TextField
                    fullWidth
                    label="文件名"
                    value={formData.filename}
                    onChange={(e) => setFormData({ ...formData, filename: e.target.value })}
                    placeholder="例如: clash.yaml"
                  />
                  <FormControl sx={{ minWidth: 120 }}>
                    <InputLabel>类别</InputLabel>
                    <Select value={formData.category} label="类别" onChange={(e) => setFormData({ ...formData, category: e.target.value })}>
                      <MenuItem value="clash">Clash</MenuItem>
                      <MenuItem value="surge">Surge</MenuItem>
                    </Select>
                  </FormControl>
                </Stack>
                <Autocomplete
                  freeSolo
                  options={aclPresets}
                  getOptionLabel={(option) => {
                    if (typeof option === 'string') return option;
                    return option.label || option.url || '';
                  }}
                  isOptionEqualToValue={(option, value) => {
                    // 如果 value 是字符串，比较 URL
                    if (typeof value === 'string') {
                      return option.url === value;
                    }
                    // 如果 value 是对象，比较 URL
                    return option.url === value?.url;
                  }}
                  value={
                    // 如果 ruleSource 匹配某个预设的 URL，返回该预设对象
                    aclPresets.find((preset) => preset.url === formData.ruleSource) || formData.ruleSource
                  }
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
                      label="远程规则地址"
                      placeholder="输入 URL 或选择 ACL4SSR 预设"
                      helperText="可填写远程 ACL 规则配置地址，生成订阅时会动态加载规则"
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
                  label="使用代理下载远程规则"
                />
                {useProxy && (
                  <Box>
                    <SearchableNodeSelect
                      nodes={proxyNodeOptions}
                      loading={loadingProxyNodes}
                      value={
                        proxyNodeOptions.find((n) => n.Link === proxyLink) || (proxyLink ? { Link: proxyLink, Name: '', ID: 0 } : null)
                      }
                      onChange={(newValue) => setProxyLink(newValue?.Link || '')}
                      displayField="Name"
                      valueField="Link"
                      label="选择代理节点"
                      placeholder="留空则自动选择最佳节点"
                      helperText="如果未选择具体代理，系统将自动选择延迟最低且速度最快的节点"
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
                  label="使用 Include-All 模式"
                />
                <Typography variant="caption" color="textSecondary" component="div" sx={{ ml: 6, mt: -0.5, lineHeight: 1.6 }}>
                  • 开启：配置更精简，使用客户端 include-all 自动匹配节点，不遵循系统排序
                </Typography>
                <Typography variant="caption" color="textSecondary" component="div" sx={{ ml: 6, lineHeight: 1.6 }}>
                  • 关闭（推荐）：由系统按顺序插入节点，遵循系统排序和过滤规则
                </Typography>
                <Stack direction="row" spacing={1}>
                  <Button
                    variant="outlined"
                    startIcon={converting ? <CircularProgress size={18} /> : <TransformIcon />}
                    disabled={!formData.ruleSource || converting}
                    onClick={() => handleConvertTemplate(false)}
                  >
                    规则生成/转换
                  </Button>
                  <Button
                    variant="outlined"
                    startIcon={converting ? <CircularProgress size={18} /> : <UnfoldMoreIcon />}
                    disabled={!formData.ruleSource || converting}
                    onClick={() => handleConvertTemplate(true)}
                  >
                    规则生成/转换（远程规则展开模式）
                  </Button>
                  <Button
                    variant="outlined"
                    color="error"
                    disabled={!formData.text || converting}
                    onClick={() => {
                      openConfirm('清空内容', '确定要清空编辑器中的所有内容吗？', () => {
                        setFormData({ ...formData, text: '' });
                        showMessage('已清空内容');
                      });
                    }}
                  >
                    清空内容
                  </Button>
                </Stack>
                <Stack direction="row" justifyContent="flex-end" spacing={1}>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={editorFullscreen ? <FullscreenExitIcon /> : <FullscreenIcon />}
                    onClick={() => setEditorFullscreen((prev) => !prev)}
                  >
                    {editorFullscreen ? '退出全屏' : '全屏编辑'}
                  </Button>
                </Stack>
                <Box
                  sx={{
                    ...aiWorkspacePanelSx,
                    p: { xs: 1, md: 1.25 }
                  }}
                >
                  {renderAIControlPanel({ minimal: true })}
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
            <Button onClick={handleCloseDialog}>取消</Button>
            <Button variant="contained" disabled={templateEditorMode === 'diff'} onClick={handleSubmit}>
              确定
            </Button>
          </DialogActions>
        )}
      </Dialog>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>

      {/* 确认对话框 */}
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
          <Button onClick={handleConfirmClose}>取消</Button>
          <Button onClick={handleConfirmAction} variant="contained" color="error" autoFocus>
            确定
          </Button>
        </DialogActions>
      </Dialog>

      {/* 错误提示对话框 */}
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
            知道了
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
                使用中的订阅：
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
          <Button onClick={() => setUsageDialog({ ...usageDialog, open: false, subscriptions: [], action: null })}>取消</Button>
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
            继续删除
          </Button>
        </DialogActions>
      </Dialog>

      {/* 基础模板编辑对话框 */}
      <Dialog open={baseTemplateDialogOpen} onClose={() => setBaseTemplateDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>{baseTemplateCategory === 'clash' ? 'Clash' : 'Surge'} 基础模板配置</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
            基础模板用于规则转换时，当模板内容为空时自动填充的默认配置。修改后将影响所有使用默认模板的规则转换操作。
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
          <Button onClick={() => setBaseTemplateDialogOpen(false)}>取消</Button>
          <Button
            variant="contained"
            onClick={handleSaveBaseTemplate}
            disabled={baseTemplateLoading || baseTemplateSaving}
            startIcon={baseTemplateSaving ? <CircularProgress size={18} /> : null}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
