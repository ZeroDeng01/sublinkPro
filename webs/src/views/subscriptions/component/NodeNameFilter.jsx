import { useState, useCallback, useEffect } from 'react';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Alert from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';
import Fade from '@mui/material/Fade';
import Switch from '@mui/material/Switch';
import Collapse from '@mui/material/Collapse';
import Divider from '@mui/material/Divider';
import Chip from '@mui/material/Chip';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import BlockIcon from '@mui/icons-material/Block';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';

// 示例节点名称用于预览
const PREVIEW_NODE_NAMES = ['香港节点-01-Premium', '美国-测试节点-02', '日本-Tokyo-03', '新加坡节点-04', '台湾-Premium-05'];

/**
 * 节点名称过滤规则编辑器
 * 白名单/黑名单过滤，支持文本和正则匹配
 */
export default function NodeNameFilter({ whitelistValue, blacklistValue, onWhitelistChange, onBlacklistChange }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [whitelistRules, setWhitelistRules] = useState([]);
  const [blacklistRules, setBlacklistRules] = useState([]);
  const [expanded, setExpanded] = useState(false);
  const [whitelistIdCounter, setWhitelistIdCounter] = useState(0);
  const [blacklistIdCounter, setBlacklistIdCounter] = useState(0);

  // 解析传入的JSON规则
  useEffect(() => {
    if (whitelistValue) {
      try {
        const parsed = JSON.parse(whitelistValue);
        if (Array.isArray(parsed)) {
          const rulesWithId = parsed.map((rule, idx) => ({
            ...rule,
            id: `whitelist-${idx}`
          }));
          setWhitelistRules(rulesWithId);
          setWhitelistIdCounter(parsed.length);
        }
      } catch {
        setWhitelistRules([]);
      }
    }
    if (blacklistValue) {
      try {
        const parsed = JSON.parse(blacklistValue);
        if (Array.isArray(parsed)) {
          const rulesWithId = parsed.map((rule, idx) => ({
            ...rule,
            id: `blacklist-${idx}`
          }));
          setBlacklistRules(rulesWithId);
          setBlacklistIdCounter(parsed.length);
        }
      } catch {
        setBlacklistRules([]);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 同步规则到父组件
  const syncWhitelistRules = useCallback(
    (newRules) => {
      const rulesForSave = newRules.map(({ id, ...rest }) => rest);
      onWhitelistChange(JSON.stringify(rulesForSave));
    },
    [onWhitelistChange]
  );

  const syncBlacklistRules = useCallback(
    (newRules) => {
      const rulesForSave = newRules.map(({ id, ...rest }) => rest);
      onBlacklistChange(JSON.stringify(rulesForSave));
    },
    [onBlacklistChange]
  );

  // 添加新规则
  const handleAddWhitelistRule = () => {
    const newRule = {
      id: `whitelist-${whitelistIdCounter}`,
      matchMode: 'text',
      pattern: '',
      enabled: true
    };
    const newRules = [...whitelistRules, newRule];
    setWhitelistRules(newRules);
    setWhitelistIdCounter(whitelistIdCounter + 1);
    syncWhitelistRules(newRules);
    setExpanded(true);
  };

  const handleAddBlacklistRule = () => {
    const newRule = {
      id: `blacklist-${blacklistIdCounter}`,
      matchMode: 'text',
      pattern: '',
      enabled: true
    };
    const newRules = [...blacklistRules, newRule];
    setBlacklistRules(newRules);
    setBlacklistIdCounter(blacklistIdCounter + 1);
    syncBlacklistRules(newRules);
    setExpanded(true);
  };

  // 更新规则
  const handleUpdateRule = (listType, id, field, val) => {
    if (listType === 'whitelist') {
      const newRules = whitelistRules.map((rule) => (rule.id === id ? { ...rule, [field]: val } : rule));
      setWhitelistRules(newRules);
      syncWhitelistRules(newRules);
    } else {
      const newRules = blacklistRules.map((rule) => (rule.id === id ? { ...rule, [field]: val } : rule));
      setBlacklistRules(newRules);
      syncBlacklistRules(newRules);
    }
  };

  // 删除规则
  const handleDeleteRule = (listType, id) => {
    if (listType === 'whitelist') {
      const newRules = whitelistRules.filter((rule) => rule.id !== id);
      setWhitelistRules(newRules);
      syncWhitelistRules(newRules);
    } else {
      const newRules = blacklistRules.filter((rule) => rule.id !== id);
      setBlacklistRules(newRules);
      syncBlacklistRules(newRules);
    }
  };

  // 拖拽结束
  const onDragEnd = (listType) => (result) => {
    if (!result.destination) return;
    const rules = listType === 'whitelist' ? whitelistRules : blacklistRules;
    const setRules = listType === 'whitelist' ? setWhitelistRules : setBlacklistRules;
    const syncRules = listType === 'whitelist' ? syncWhitelistRules : syncBlacklistRules;

    const items = Array.from(rules);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setRules(items);
    syncRules(items);
  };

  // 检查节点名称是否匹配规则
  const matchesRules = (nodeName, rules) => {
    for (const rule of rules) {
      if (!rule.enabled || !rule.pattern) continue;
      try {
        if (rule.matchMode === 'regex') {
          const regex = new RegExp(rule.pattern);
          if (regex.test(nodeName)) return true;
        } else {
          if (nodeName.includes(rule.pattern)) return true;
        }
      } catch {
        // 忽略无效正则
      }
    }
    return false;
  };

  // 计算预览结果
  const getFilteredNodes = () => {
    const hasWhitelist = whitelistRules.some((r) => r.enabled && r.pattern);
    const hasBlacklist = blacklistRules.some((r) => r.enabled && r.pattern);

    return PREVIEW_NODE_NAMES.map((name) => {
      const inBlacklist = matchesRules(name, blacklistRules);
      const inWhitelist = matchesRules(name, whitelistRules);

      // 黑名单优先
      if (hasBlacklist && inBlacklist) {
        return { name, status: 'excluded', reason: '黑名单' };
      }
      // 白名单检查
      if (hasWhitelist && !inWhitelist) {
        return { name, status: 'excluded', reason: '不在白名单' };
      }
      return { name, status: 'included', reason: '' };
    });
  };

  // 只统计启用且有pattern的规则
  const activeWhitelistRules = whitelistRules.filter((r) => r.enabled && r.pattern);
  const activeBlacklistRules = blacklistRules.filter((r) => r.enabled && r.pattern);
  const hasActiveWhitelist = activeWhitelistRules.length > 0;
  const hasActiveBlacklist = activeBlacklistRules.length > 0;
  const hasWhitelistRules = whitelistRules.length > 0;
  const hasBlacklistRules = blacklistRules.length > 0;
  const hasAnyRules = hasWhitelistRules || hasBlacklistRules;
  const hasAnyActiveRules = hasActiveWhitelist || hasActiveBlacklist;
  const filteredPreview = getFilteredNodes();
  const includedCount = filteredPreview.filter((n) => n.status === 'included').length;

  // 渲染规则列表
  const renderRuleList = (listType, rules, title, icon, color) => (
    <Box sx={{ mb: 2 }}>
      <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1.5 }}>
        {icon}
        <Typography variant="subtitle2" fontWeight={600} color={`${color}.main`}>
          {title}
        </Typography>
        {rules.length > 0 && (
          <Chip label={`${rules.filter((r) => r.enabled).length}/${rules.length}`} size="small" color={color} variant="outlined" />
        )}
        <Box sx={{ flex: 1 }} />
        <Button
          size="small"
          startIcon={<AddIcon />}
          onClick={listType === 'whitelist' ? handleAddWhitelistRule : handleAddBlacklistRule}
          color={color}
        >
          添加
        </Button>
      </Stack>

      {rules.length > 0 ? (
        <DragDropContext onDragEnd={onDragEnd(listType)}>
          <Droppable droppableId={`${listType}Rules`}>
            {(provided) => (
              <Box ref={provided.innerRef} {...provided.droppableProps}>
                {rules.map((rule, index) => (
                  <Draggable key={rule.id} draggableId={rule.id} index={index}>
                    {(provided, snapshot) => (
                      <Fade in>
                        <Paper
                          ref={provided.innerRef}
                          {...provided.draggableProps}
                          elevation={snapshot.isDragging ? 4 : 0}
                          sx={{
                            p: isMobile ? 1.5 : 1,
                            mb: 1,
                            border: '1px solid',
                            borderColor: rule.enabled ? `${color}.light` : 'divider',
                            borderRadius: 1.5,
                            bgcolor: snapshot.isDragging ? 'action.selected' : rule.enabled ? 'transparent' : 'action.disabledBackground',
                            opacity: rule.enabled ? 1 : 0.6,
                            transition: 'all 0.2s ease'
                          }}
                        >
                          <Stack direction={isMobile ? 'column' : 'row'} spacing={1} alignItems={isMobile ? 'stretch' : 'center'}>
                            {/* 拖拽手柄 */}
                            <Box
                              {...provided.dragHandleProps}
                              sx={{
                                display: 'flex',
                                alignItems: 'center',
                                cursor: 'grab',
                                color: 'text.secondary'
                              }}
                            >
                              <DragIndicatorIcon fontSize="small" />
                            </Box>

                            {/* 启用开关 */}
                            <Switch
                              size="small"
                              checked={rule.enabled}
                              onChange={(e) => handleUpdateRule(listType, rule.id, 'enabled', e.target.checked)}
                              color={color}
                            />

                            {/* 匹配模式 */}
                            <FormControl size="small" sx={{ minWidth: isMobile ? '100%' : 80 }}>
                              <Select
                                value={rule.matchMode}
                                onChange={(e) => handleUpdateRule(listType, rule.id, 'matchMode', e.target.value)}
                              >
                                <MenuItem value="text">文本</MenuItem>
                                <MenuItem value="regex">正则</MenuItem>
                              </Select>
                            </FormControl>

                            {/* 匹配内容 */}
                            <TextField
                              size="small"
                              placeholder={rule.matchMode === 'regex' ? '正则表达式' : '关键字'}
                              value={rule.pattern}
                              onChange={(e) => handleUpdateRule(listType, rule.id, 'pattern', e.target.value)}
                              sx={{ flex: 1, minWidth: isMobile ? '100%' : 150 }}
                              error={
                                rule.matchMode === 'regex' &&
                                rule.pattern &&
                                (() => {
                                  try {
                                    new RegExp(rule.pattern);
                                    return false;
                                  } catch {
                                    return true;
                                  }
                                })()
                              }
                              helperText={
                                rule.matchMode === 'regex' &&
                                rule.pattern &&
                                (() => {
                                  try {
                                    new RegExp(rule.pattern);
                                    return null;
                                  } catch {
                                    return '无效正则';
                                  }
                                })()
                              }
                            />

                            {/* 删除按钮 */}
                            <Tooltip title="删除规则">
                              <IconButton size="small" color="error" onClick={() => handleDeleteRule(listType, rule.id)}>
                                <DeleteOutlineIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                          </Stack>
                        </Paper>
                      </Fade>
                    )}
                  </Draggable>
                ))}
                {provided.placeholder}
              </Box>
            )}
          </Droppable>
        </DragDropContext>
      ) : (
        <Typography variant="body2" color="textSecondary" sx={{ py: 1, textAlign: 'center' }}>
          暂无{title}规则
        </Typography>
      )}
    </Box>
  );

  return (
    <Paper
      elevation={0}
      sx={{
        mb: 2,
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 2,
        overflow: 'hidden'
      }}
    >
      {/* 标题栏 */}
      <Box
        sx={{
          p: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a2027' : '#f5f5f5'} 0%, ${theme.palette.mode === 'dark' ? '#121417' : '#fafafa'} 100%)`,
          cursor: 'pointer',
          '&:hover': {
            bgcolor: 'action.hover'
          }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <FilterAltIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600}>
            节点名称过滤
          </Typography>
          {hasAnyRules && (
            <Typography variant="caption" color="textSecondary">
              (白名单 {activeWhitelistRules.length} / 黑名单 {activeBlacklistRules.length} 条有效规则)
            </Typography>
          )}
        </Stack>
        <Stack direction="row" alignItems="center" spacing={0.5}>
          {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
        </Stack>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ p: 2, pt: 1 }}>
          {/* 说明 */}
          <Alert variant={'standard'} severity="info" sx={{ mb: 2 }}>
            <Typography variant="body2">
              基于节点<strong>原始名称</strong>进行过滤。<strong>黑名单优先级高于白名单</strong>。
            </Typography>
          </Alert>

          {/* 白名单 */}
          {renderRuleList('whitelist', whitelistRules, '白名单', <CheckCircleOutlineIcon color="success" fontSize="small" />, 'success')}

          <Divider sx={{ my: 2 }} />

          {/* 黑名单 */}
          {renderRuleList('blacklist', blacklistRules, '黑名单', <BlockIcon color="error" fontSize="small" />, 'error')}

          {/* 实时预览 */}
          {hasAnyActiveRules && (
            <Fade in>
              <Alert variant={'standard'} severity={includedCount < PREVIEW_NODE_NAMES.length ? 'warning' : 'success'} sx={{ mt: 2 }}>
                <Typography variant="body2" sx={{ mb: 1 }}>
                  <strong>预览效果</strong>（{includedCount}/{PREVIEW_NODE_NAMES.length} 个示例节点通过过滤）
                </Typography>
                <Stack spacing={0.5}>
                  {filteredPreview.map((node, idx) => (
                    <Stack key={idx} direction="row" alignItems="center" spacing={1}>
                      {node.status === 'included' ? (
                        <CheckCircleOutlineIcon fontSize="small" color="success" />
                      ) : (
                        <BlockIcon fontSize="small" color="error" />
                      )}
                      <Typography
                        variant="body2"
                        sx={{
                          textDecoration: node.status === 'excluded' ? 'line-through' : 'none',
                          color: node.status === 'excluded' ? 'text.disabled' : 'text.primary'
                        }}
                      >
                        {node.name}
                        {node.reason && (
                          <Typography component="span" variant="caption" color="error" sx={{ ml: 1 }}>
                            ({node.reason})
                          </Typography>
                        )}
                      </Typography>
                    </Stack>
                  ))}
                </Stack>
              </Alert>
            </Fade>
          )}
        </Box>
      </Collapse>
    </Paper>
  );
}
