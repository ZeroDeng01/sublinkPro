import { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import InputAdornment from '@mui/material/InputAdornment';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Tooltip from '@mui/material/Tooltip';
import Alert from '@mui/material/Alert';

// icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import SaveIcon from '@mui/icons-material/Save';
import RestoreIcon from '@mui/icons-material/Restore';
import EditIcon from '@mui/icons-material/Edit';
import VisibilityIcon from '@mui/icons-material/Visibility';
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff';
import SettingsIcon from '@mui/icons-material/Settings';
import SecurityIcon from '@mui/icons-material/Security';
import NetworkCheckIcon from '@mui/icons-material/NetworkCheck';
import VpnKeyIcon from '@mui/icons-material/VpnKey';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// api
import { parseNodeLink, updateNodeRawInfo } from '../../../api/nodes';
import { buildFieldMetaMap, getFieldGroupKey } from '../../../utils/protocolPresentation';
import { getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';
import { withAlpha } from '../../../utils/colorUtils';

/**
 * 字段分组配置
 * 用于将字段分组展示，提高可读性
 */
const FIELD_GROUPS = {
  basic: {
    labelKey: 'nodes.rawInfo.groups.basic',
    icon: <SettingsIcon fontSize="small" />,
    // 匹配规则：字段名包含这些关键词
    keywords: ['Name', 'Ps', 'Server', 'Host', 'Add', 'Port', 'Hostname']
  },
  auth: {
    labelKey: 'nodes.rawInfo.groups.auth',
    icon: <VpnKeyIcon fontSize="small" />,
    keywords: ['Password', 'Uuid', 'Id', 'Auth', 'Username']
  },
  transport: {
    labelKey: 'nodes.rawInfo.groups.transport',
    icon: <NetworkCheckIcon fontSize="small" />,
    keywords: ['Net', 'Type', 'Path', 'Encryption', 'Cipher', 'Method', 'Obfs', 'Protocol', 'Flow', 'Mode', 'ServiceName', 'HeaderType']
  },
  tls: {
    labelKey: 'nodes.rawInfo.groups.tls',
    icon: <SecurityIcon fontSize="small" />,
    keywords: [
      'Tls',
      'Security',
      'Sni',
      'Alpn',
      'Fp',
      'Pbk',
      'Sid',
      'Peer',
      'Insecure',
      'SkipCertVerify',
      'ClientFingerprint',
      'AllowInsecure'
    ]
  },
  advanced: {
    labelKey: 'nodes.rawInfo.groups.advanced',
    icon: <SettingsIcon fontSize="small" />,
    keywords: []
  }
};

/**
 * 获取字段显示标签
 */
const getFieldLabel = (fieldName, fieldMeta) => {
  // 优先使用元数据中的 label
  if (fieldMeta?.label) {
    return fieldMeta.label + '(' + fieldName.split('.').pop() + ')';
  }
  // 否则使用字段名的最后一部分
  return fieldName.split('.').pop();
};

/**
 * 渲染字段输入控件
 */
const FieldInput = ({ fieldName, fieldMeta, value, onChange, disabled, fieldSx, tokens }) => {
  const { t } = useTranslation();
  const [showSecret, setShowSecret] = useState(false);
  const fieldType = fieldMeta?.type || 'string';
  const label = getFieldLabel(fieldName, fieldMeta);
  const placeholder = fieldMeta?.placeholder || '';
  const helperText = fieldMeta?.description || '';
  const multiline = fieldMeta?.multiline || String(value ?? '').length > 50;

  // 布尔类型使用开关
  if (fieldType === 'bool') {
    return (
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          p: 1.5,
          px: 2,
          minHeight: 52,
          borderRadius: 1.5,
          bgcolor: tokens?.fieldSurface || 'background.paper',
          border: '1px solid',
          borderColor: value === true || value === 'true' ? tokens?.selectedBorder : tokens?.softBorder,
          transition: 'all 0.2s ease',
          ...(value === true || value === 'true'
            ? {
                bgcolor: tokens?.selectedSurface || 'action.selected'
              }
            : {}),
          ...fieldSx,
          '& .MuiOutlinedInput-root': undefined
        }}
      >
        <Box sx={{ pr: 2 }}>
          <Typography variant="body2" fontWeight={600} color={tokens?.primaryText || 'text.primary'}>
            {label}
          </Typography>
          {helperText && (
            <Typography
              variant="caption"
              color={tokens?.tertiaryText || 'text.secondary'}
              display="block"
              sx={{ mt: 0.5, lineHeight: 1.2 }}
            >
              {helperText}
            </Typography>
          )}
        </Box>
        <Switch
          checked={value === true || value === 'true'}
          onChange={(e) => onChange(fieldName, e.target.checked)}
          disabled={disabled}
          color="primary"
          sx={{ ml: 'auto', mr: -1 }}
        />
      </Box>
    );
  }

  // 数字类型
  if (fieldType === 'int') {
    return (
      <TextField
        label={label}
        type="number"
        value={value ?? ''}
        onChange={(e) => onChange(fieldName, e.target.value ? parseInt(e.target.value, 10) : '')}
        disabled={disabled}
        size="small"
        fullWidth
        variant="outlined"
        placeholder={placeholder}
        helperText={helperText}
        slotProps={{ inputLabel: { shrink: true } }}
        sx={fieldSx}
      />
    );
  }

  if (Array.isArray(fieldMeta?.options) && fieldMeta.options.length > 0) {
    return (
      <TextField
        label={label}
        value={value ?? ''}
        onChange={(e) => onChange(fieldName, e.target.value)}
        disabled={disabled}
        size="small"
        fullWidth
        variant="outlined"
        select
        SelectProps={{ native: true }}
        InputLabelProps={{ shrink: true }}
        helperText={helperText}
        sx={fieldSx}
      >
        <option value="">{t('nodes.rawInfo.selectPlaceholder')}</option>
        {fieldMeta.options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </TextField>
    );
  }

  return (
    <TextField
      label={label}
      type={fieldMeta?.secret && !showSecret ? 'password' : 'text'}
      value={value ?? ''}
      onChange={(e) => onChange(fieldName, e.target.value)}
      disabled={disabled}
      size="small"
      fullWidth
      variant="outlined"
      placeholder={placeholder}
      helperText={helperText}
      multiline={multiline}
      maxRows={3}
      slotProps={{ inputLabel: { shrink: true } }}
      sx={fieldSx}
      InputProps={
        fieldMeta?.secret
          ? {
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={t('auth.login.togglePassword')} placement="top">
                    <Box component="span" sx={{ display: 'inline-flex' }}>
                      <IconButton
                        edge="end"
                        size="small"
                        onClick={() => setShowSecret((prev) => !prev)}
                        disabled={disabled}
                        sx={{
                          color: showSecret ? 'primary.main' : tokens?.secondaryText || 'text.secondary',
                          bgcolor: showSecret ? tokens?.selectedSurface || 'action.selected' : 'transparent',
                          '&:hover': {
                            bgcolor: showSecret ? tokens?.selectedHoverSurface : tokens?.hoverSurface
                          }
                        }}
                      >
                        {showSecret ? <VisibilityOffIcon fontSize="small" /> : <VisibilityIcon fontSize="small" />}
                      </IconButton>
                    </Box>
                  </Tooltip>
                </InputAdornment>
              )
            }
          : undefined
      }
    />
  );
};

FieldInput.propTypes = {
  fieldName: PropTypes.string.isRequired,
  fieldMeta: PropTypes.object,
  value: PropTypes.any,
  onChange: PropTypes.func.isRequired,
  disabled: PropTypes.bool,
  fieldSx: PropTypes.object,
  tokens: PropTypes.object
};

/**
 * 节点原始信息编辑器组件
 */
export default function NodeRawInfoEditor({ node, protocolMeta, onUpdate, showMessage }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens);

  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [parsedInfo, setParsedInfo] = useState(null);
  const [editedFields, setEditedFields] = useState({});
  const [error, setError] = useState(null);
  const [expandedGroups, setExpandedGroups] = useState(['basic']);

  // 获取当前协议的元数据
  const currentProtocolMeta = useMemo(() => {
    if (!parsedInfo?.protocol || !protocolMeta) return null;
    return protocolMeta.find((p) => p.name === parsedInfo.protocol);
  }, [parsedInfo, protocolMeta]);

  // 创建字段元数据映射
  const fieldMetaMap = useMemo(() => {
    if (!currentProtocolMeta?.fields) return {};
    return buildFieldMetaMap(currentProtocolMeta.fields);
  }, [currentProtocolMeta]);

  // 解析节点链接
  useEffect(() => {
    if (!node?.Link) {
      setParsedInfo(null);
      return;
    }

    setLoading(true);
    setError(null);
    parseNodeLink(node.Link)
      .then((res) => {
        if (res.data) {
          setParsedInfo(res.data);
          setEditedFields(res.data.fields || {});
          setExpandedGroups(['basic']);
        }
      })
      .catch((err) => {
        console.error('解析节点失败:', err);
        setError(t('nodes.rawInfo.messages.parseFailed'));
      })
      .finally(() => setLoading(false));
  }, [node?.Link, t]);

  useEffect(() => {
    if (!editedFields || Object.keys(editedFields).length === 0) {
      return;
    }

    const autoExpanded = ['basic'];
    Object.keys(editedFields).forEach((fieldName) => {
      const groupKey = getFieldGroupKey(fieldName, fieldMetaMap[fieldName]);
      if (groupKey !== 'advanced' && !autoExpanded.includes(groupKey)) {
        autoExpanded.push(groupKey);
      }
    });
    setExpandedGroups(autoExpanded);
  }, [editedFields, fieldMetaMap]);

  // 按分组组织字段
  const groupedFields = useMemo(() => {
    if (!editedFields) return {};

    const groups = {
      basic: [],
      auth: [],
      transport: [],
      tls: [],
      advanced: [],
      other: []
    };

    Object.keys(editedFields).forEach((fieldName) => {
      const group = getFieldGroupKey(fieldName, fieldMetaMap[fieldName]);
      const targetGroup = groups[group] ? group : 'other';
      groups[targetGroup].push(fieldName);
    });

    // 移除空分组
    Object.keys(groups).forEach((key) => {
      if (groups[key].length === 0) {
        delete groups[key];
      }
    });

    return groups;
  }, [editedFields, fieldMetaMap]);

  // 处理字段值变更
  const handleFieldChange = (fieldName, value) => {
    setEditedFields((prev) => ({
      ...prev,
      [fieldName]: value
    }));
  };

  // 重置编辑
  const handleReset = () => {
    if (parsedInfo?.fields) {
      setEditedFields({ ...parsedInfo.fields });
    }
    setEditMode(false);
  };

  // 保存更改
  const handleSave = async () => {
    if (!node?.ID) return;

    setSaving(true);
    try {
      const res = await updateNodeRawInfo(node.ID, editedFields);
      if (res.data) {
        showMessage?.(t('common.saveSuccess'), 'success');
        setEditMode(false);
        onUpdate?.();
      }
    } catch (err) {
      console.error('保存失败:', err);
      showMessage?.(err.response?.data?.msg || t('common.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  // 切换分组展开状态
  const handleGroupToggle = (groupKey) => {
    setExpandedGroups((prev) => (prev.includes(groupKey) ? prev.filter((k) => k !== groupKey) : [...prev, groupKey]));
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
        <CircularProgress size={32} />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  if (!parsedInfo) {
    return (
      <Typography variant="body2" color="text.secondary" sx={{ py: 2, textAlign: 'center' }}>
        {t('nodes.rawInfo.unavailable')}
      </Typography>
    );
  }

  return (
    <Box>
      {/* 头部：协议类型和编辑按钮 */}
      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 3 }}>
        <Typography variant="caption" sx={{ color: tokens.secondaryText, fontWeight: 600 }}>
          {t('nodes.rawInfo.fieldCount', { count: Object.keys(editedFields).length })}
        </Typography>

        <Tooltip title={editMode ? t('nodes.rawInfo.viewMode') : t('nodes.rawInfo.editMode')}>
          <IconButton
            size="small"
            onClick={() => setEditMode(!editMode)}
            sx={{
              bgcolor: editMode ? tokens.hoverSurface : 'transparent',
              color: editMode ? 'primary.main' : 'text.secondary'
            }}
          >
            {editMode ? <VisibilityIcon fontSize="small" /> : <EditIcon fontSize="small" />}
          </IconButton>
        </Tooltip>
      </Stack>

      {/* 字段分组展示 */}
      {Object.entries(groupedFields).map(([groupKey, fields]) => {
        const groupConfig = FIELD_GROUPS[groupKey] || { labelKey: 'nodes.rawInfo.groups.other', icon: <SettingsIcon fontSize="small" /> };

        return (
          <Accordion
            key={groupKey}
            expanded={expandedGroups.includes(groupKey)}
            onChange={() => handleGroupToggle(groupKey)}
            disableGutters
            elevation={0}
            sx={{
              bgcolor: tokens.nestedPanelSurface,
              '&:before': { display: 'none' },
              border: '1px solid',
              borderColor: expandedGroups.includes(groupKey) ? tokens.selectedBorder : tokens.softBorder,
              borderRadius: 2,
              mb: 1.5,
              overflow: 'hidden',
              transition: 'all 0.2s ease',
              ...(expandedGroups.includes(groupKey) && {
                boxShadow: tokens.isDark
                  ? `0 4px 16px ${withAlpha(theme.palette.common.black, 0.2)}`
                  : `0 4px 12px ${withAlpha(theme.palette.common.black, 0.04)}`
              })
            }}
          >
            <AccordionSummary
              expandIcon={
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    bgcolor: tokens.fieldSurface,
                    borderRadius: '50%',
                    width: 28,
                    height: 28,
                    border: '1px solid',
                    borderColor: tokens.subtleBorder,
                    transition: 'all 0.2s ease',
                    '&:hover': {
                      bgcolor: tokens.hoverSurface,
                      color: 'primary.main',
                      borderColor: tokens.selectedBorder
                    }
                  }}
                >
                  <ExpandMoreIcon fontSize="small" />
                </Box>
              }
              sx={{
                minHeight: 56,
                bgcolor: expandedGroups.includes(groupKey) ? tokens.hoverSurface : tokens.toolbarSurface,
                borderBottom: expandedGroups.includes(groupKey) ? '1px solid' : 'none',
                borderColor: tokens.softBorder,
                transition: 'background-color 0.2s ease',
                '&:hover': {
                  bgcolor: expandedGroups.includes(groupKey) ? tokens.hoverSurface : tokens.fieldSurfaceActive
                },
                '&.Mui-expanded': {
                  minHeight: 56
                },
                '& .MuiAccordionSummary-content, & .MuiAccordionSummary-content.Mui-expanded': {
                  my: 1.5,
                  alignItems: 'center'
                }
              }}
            >
              <Stack direction="row" alignItems="center" spacing={1.5} sx={{ width: '100%' }}>
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: expandedGroups.includes(groupKey) ? 'primary.main' : tokens.secondaryText,
                    bgcolor: expandedGroups.includes(groupKey) ? tokens.selectedSurface : tokens.fieldSurface,
                    borderRadius: 1.5,
                    width: 32,
                    height: 32,
                    border: '1px solid',
                    borderColor: expandedGroups.includes(groupKey) ? tokens.selectedBorder : tokens.softBorder,
                    transition: 'all 0.2s ease'
                  }}
                >
                  {groupConfig.icon}
                </Box>
                <Typography
                  variant="subtitle1"
                  fontWeight={600}
                  sx={{
                    flexGrow: 1,
                    color: expandedGroups.includes(groupKey) ? tokens.primaryText : tokens.secondaryText,
                    transition: 'color 0.2s ease'
                  }}
                >
                  {t(groupConfig.labelKey)}
                </Typography>
                <Chip
                  label={fields.length}
                  size="small"
                  sx={{
                    height: 22,
                    fontSize: 12,
                    fontWeight: 600,
                    bgcolor: expandedGroups.includes(groupKey) ? 'primary.main' : tokens.fieldSurface,
                    color: expandedGroups.includes(groupKey) ? theme.palette.primary.contrastText : tokens.primaryText,
                    border: '1px solid',
                    borderColor: expandedGroups.includes(groupKey) ? 'primary.main' : tokens.subtleBorder,
                    transition: 'all 0.2s ease'
                  }}
                />
              </Stack>
            </AccordionSummary>
            <AccordionDetails sx={{ px: { xs: 2, sm: 2.5 }, pt: { xs: 2, sm: 2.5 }, pb: { xs: 2.25, sm: 3 }, bgcolor: tokens.cardSurface }}>
              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: { xs: '1fr', sm: fields.length > 2 ? 'repeat(2, 1fr)' : '1fr' },
                  gap: { xs: 2, sm: 2.5 },
                  alignItems: 'start'
                }}
              >
                {fields.map((fieldName) => (
                  <Box key={fieldName}>
                    <FieldInput
                      fieldName={fieldName}
                      fieldMeta={fieldMetaMap[fieldName]}
                      value={editedFields[fieldName]}
                      onChange={handleFieldChange}
                      disabled={!editMode}
                      fieldSx={fieldControlSx}
                      tokens={tokens}
                    />
                  </Box>
                ))}
              </Box>
            </AccordionDetails>
          </Accordion>
        );
      })}

      {/* 编辑模式下的操作按钮 */}
      {editMode && (
        <>
          <Divider sx={{ my: 2 }} />
          <Stack direction="row" spacing={1} justifyContent="flex-end">
            <Button
              variant="outlined"
              startIcon={<RestoreIcon />}
              onClick={handleReset}
              disabled={saving}
              size={isMobile ? 'medium' : 'small'}
            >
              {t('common.reset')}
            </Button>
            <Button
              variant="contained"
              startIcon={saving ? <CircularProgress size={16} color="inherit" /> : <SaveIcon />}
              onClick={handleSave}
              disabled={saving}
              size={isMobile ? 'medium' : 'small'}
            >
              {t('common.save')}
            </Button>
          </Stack>
        </>
      )}
    </Box>
  );
}

NodeRawInfoEditor.propTypes = {
  node: PropTypes.object, // 节点对象
  protocolMeta: PropTypes.array, // 协议元数据列表
  onUpdate: PropTypes.func, // 更新成功回调
  showMessage: PropTypes.func // 消息提示函数
};
