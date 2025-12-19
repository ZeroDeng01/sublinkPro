import { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
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
import SettingsIcon from '@mui/icons-material/Settings';
import SecurityIcon from '@mui/icons-material/Security';
import NetworkCheckIcon from '@mui/icons-material/NetworkCheck';
import VpnKeyIcon from '@mui/icons-material/VpnKey';

// api
import { parseNodeLink, updateNodeRawInfo } from '../../../api/nodes';

/**
 * 字段分组配置
 * 用于将字段分组展示，提高可读性
 */
const FIELD_GROUPS = {
  basic: {
    label: '基础信息',
    icon: <SettingsIcon fontSize="small" />,
    // 匹配规则：字段名包含这些关键词
    keywords: ['Name', 'Ps', 'Server', 'Host', 'Add', 'Port', 'Hostname']
  },
  auth: {
    label: '认证信息',
    icon: <VpnKeyIcon fontSize="small" />,
    keywords: ['Password', 'Uuid', 'Id', 'Auth', 'Username']
  },
  transport: {
    label: '传输配置',
    icon: <NetworkCheckIcon fontSize="small" />,
    keywords: ['Net', 'Type', 'Path', 'Encryption', 'Cipher', 'Method', 'Obfs', 'Protocol', 'Flow', 'Mode', 'ServiceName', 'HeaderType']
  },
  tls: {
    label: 'TLS/安全',
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
  }
};

/**
 * 获取字段所属分组
 */
const getFieldGroup = (fieldName) => {
  const baseName = fieldName.split('.').pop();
  for (const [groupKey, group] of Object.entries(FIELD_GROUPS)) {
    if (group.keywords.some((keyword) => baseName.toLowerCase().includes(keyword.toLowerCase()))) {
      return groupKey;
    }
  }
  return 'other';
};

/**
 * 获取字段显示标签
 */
const getFieldLabel = (fieldName, fieldMeta) => {
  // 优先使用元数据中的 label
  if (fieldMeta?.label) {
    return fieldMeta.label;
  }
  // 否则使用字段名的最后一部分
  return fieldName.split('.').pop();
};

/**
 * 渲染字段输入控件
 */
const FieldInput = ({ fieldName, fieldMeta, value, onChange, disabled }) => {
  const fieldType = fieldMeta?.type || 'string';
  const label = getFieldLabel(fieldName, fieldMeta);

  // 布尔类型使用开关
  if (fieldType === 'bool') {
    return (
      <FormControlLabel
        control={
          <Switch
            checked={value === true || value === 'true'}
            onChange={(e) => onChange(fieldName, e.target.checked)}
            disabled={disabled}
          />
        }
        label={label}
        sx={{ ml: 0 }}
      />
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
      />
    );
  }

  return (
    <TextField
      label={label}
      type={'text'}
      value={value ?? ''}
      onChange={(e) => onChange(fieldName, e.target.value)}
      disabled={disabled}
      size="small"
      fullWidth
      variant="outlined"
      multiline={String(value).length > 50}
      maxRows={3}
    />
  );
};

FieldInput.propTypes = {
  fieldName: PropTypes.string.isRequired,
  fieldMeta: PropTypes.object,
  value: PropTypes.any,
  onChange: PropTypes.func.isRequired,
  disabled: PropTypes.bool
};

/**
 * 节点原始信息编辑器组件
 */
export default function NodeRawInfoEditor({ node, protocolMeta, onUpdate, showMessage }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

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
    const map = {};
    currentProtocolMeta.fields.forEach((f) => {
      map[f.name] = f;
    });
    return map;
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
        }
      })
      .catch((err) => {
        console.error('解析节点失败:', err);
        setError('解析节点信息失败');
      })
      .finally(() => setLoading(false));
  }, [node?.Link]);

  // 按分组组织字段
  const groupedFields = useMemo(() => {
    if (!editedFields) return {};

    const groups = {
      basic: [],
      auth: [],
      transport: [],
      tls: [],
      other: []
    };

    Object.keys(editedFields).forEach((fieldName) => {
      const group = getFieldGroup(fieldName);
      groups[group].push(fieldName);
    });

    // 移除空分组
    Object.keys(groups).forEach((key) => {
      if (groups[key].length === 0) {
        delete groups[key];
      }
    });

    return groups;
  }, [editedFields]);

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
        showMessage?.('保存成功', 'success');
        setEditMode(false);
        onUpdate?.();
      }
    } catch (err) {
      console.error('保存失败:', err);
      showMessage?.(err.response?.data?.msg || '保存失败', 'error');
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
        无法解析节点信息
      </Typography>
    );
  }

  return (
    <Box>
      {/* 头部：协议类型和编辑按钮 */}
      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Chip
            label={currentProtocolMeta?.label || parsedInfo.protocol}
            size="small"
            sx={{
              bgcolor: currentProtocolMeta?.color || theme.palette.primary.main,
              color: '#fff',
              fontWeight: 600
            }}
          />
          <Typography variant="caption" color="text.secondary">
            {Object.keys(editedFields).length} 个字段
          </Typography>
        </Stack>

        <Tooltip title={editMode ? '查看模式' : '编辑模式'}>
          <IconButton
            size="small"
            onClick={() => setEditMode(!editMode)}
            sx={{
              bgcolor: editMode ? alpha(theme.palette.primary.main, 0.1) : 'transparent',
              color: editMode ? 'primary.main' : 'text.secondary'
            }}
          >
            {editMode ? <VisibilityIcon fontSize="small" /> : <EditIcon fontSize="small" />}
          </IconButton>
        </Tooltip>
      </Stack>

      {/* 字段分组展示 */}
      {Object.entries(groupedFields).map(([groupKey, fields]) => {
        const groupConfig = FIELD_GROUPS[groupKey] || { label: '其他配置', icon: <SettingsIcon fontSize="small" /> };

        return (
          <Accordion
            key={groupKey}
            expanded={expandedGroups.includes(groupKey)}
            onChange={() => handleGroupToggle(groupKey)}
            disableGutters
            elevation={0}
            sx={{
              bgcolor: 'transparent',
              '&:before': { display: 'none' },
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 2,
              mb: 1,
              overflow: 'hidden'
            }}
          >
            <AccordionSummary
              expandIcon={<ExpandMoreIcon />}
              sx={{
                minHeight: 48,
                '& .MuiAccordionSummary-content': { my: 1 }
              }}
            >
              <Stack direction="row" alignItems="center" spacing={1}>
                {groupConfig.icon}
                <Typography variant="subtitle2" fontWeight={600}>
                  {groupConfig.label}
                </Typography>
                <Chip label={fields.length} size="small" sx={{ height: 20, fontSize: 11 }} />
              </Stack>
            </AccordionSummary>
            <AccordionDetails sx={{ pt: 0 }}>
              <Stack spacing={isMobile ? 2 : 1.5}>
                {fields.map((fieldName) => (
                  <Box key={fieldName}>
                    <FieldInput
                      fieldName={fieldName}
                      fieldMeta={fieldMetaMap[fieldName]}
                      value={editedFields[fieldName]}
                      onChange={handleFieldChange}
                      disabled={!editMode}
                    />
                  </Box>
                ))}
              </Stack>
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
              重置
            </Button>
            <Button
              variant="contained"
              startIcon={saving ? <CircularProgress size={16} color="inherit" /> : <SaveIcon />}
              onClick={handleSave}
              disabled={saving}
              size={isMobile ? 'medium' : 'small'}
            >
              保存
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
