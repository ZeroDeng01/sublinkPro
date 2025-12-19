import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

// material-ui
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import Divider from '@mui/material/Divider';
import Chip from '@mui/material/Chip';
import Card from '@mui/material/Card';

// icons
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';

// 节点字段选项
const nodeFields = [
  { value: 'name', label: '备注' },
  { value: 'link_name', label: '原始名称' },
  { value: 'link_country', label: '国家代码' },
  { value: 'protocol', label: '协议类型' },
  { value: 'source', label: '来源' },
  { value: 'group', label: '分组' },
  { value: 'speed', label: '速度 (MB/s)' },
  { value: 'delay_time', label: '延迟 (ms)' },
  { value: 'speed_status', label: '速度状态' },
  { value: 'delay_status', label: '延迟状态' },
  { value: 'link_address', label: '地址' },
  { value: 'link_host', label: 'Host' },
  { value: 'link_port', label: '端口' },
  { value: 'dialer_proxy_name', label: '前置代理' },
  { value: 'link', label: '节点链接' }
];

// 状态选项（与后端 constants/status.go 保持同步）
const statusOptions = [
  { value: 'untested', label: '未测速' },
  { value: 'success', label: '成功' },
  { value: 'timeout', label: '超时' },
  { value: 'error', label: '失败' }
];

// 操作符选项
const operators = [
  { value: 'equals', label: '等于', type: 'string' },
  { value: 'not_equals', label: '不等于', type: 'string' },
  { value: 'contains', label: '包含', type: 'string' },
  { value: 'not_contains', label: '不包含', type: 'string' },
  { value: 'regex', label: '正则匹配', type: 'string' },
  { value: 'greater_than', label: '大于', type: 'number' },
  { value: 'less_than', label: '小于', type: 'number' },
  { value: 'greater_or_equal', label: '大于等于', type: 'number' },
  { value: 'less_or_equal', label: '小于等于', type: 'number' }
];

// 数值字段
const numericFields = ['speed', 'delay_time'];

// 状态字段（使用下拉框选择值）
const statusFields = ['speed_status', 'delay_status'];

export default function RuleDialog({ open, onClose, onSave, editingRule, tags }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const [name, setName] = useState('');
  const [tagName, setTagName] = useState('');
  const [enabled, setEnabled] = useState(true);
  const [triggerType, setTriggerType] = useState('subscription_update');
  const [logic, setLogic] = useState('and');
  const [conditions, setConditions] = useState([{ field: 'link_country', operator: 'equals', value: '' }]);

  useEffect(() => {
    if (editingRule) {
      setName(editingRule.name || '');
      setTagName(editingRule.tagName || '');
      setEnabled(editingRule.enabled !== false);
      setTriggerType(editingRule.triggerType || 'subscription_update');
      try {
        const parsed = JSON.parse(editingRule.conditions || '{}');
        setLogic(parsed.logic || 'and');
        setConditions(parsed.conditions?.length > 0 ? parsed.conditions : [{ field: 'link_country', operator: 'equals', value: '' }]);
      } catch {
        setLogic('and');
        setConditions([{ field: 'link_country', operator: 'equals', value: '' }]);
      }
    } else {
      setName('');
      setTagName('');
      setEnabled(true);
      setTriggerType('subscription_update');
      setLogic('and');
      setConditions([{ field: 'link_country', operator: 'equals', value: '' }]);
    }
  }, [editingRule, open]);

  const handleAddCondition = () => {
    setConditions([...conditions, { field: 'link_country', operator: 'equals', value: '' }]);
  };

  const handleRemoveCondition = (index) => {
    if (conditions.length > 1) {
      setConditions(conditions.filter((_, i) => i !== index));
    }
  };

  const handleConditionChange = (index, key, value) => {
    const newConditions = [...conditions];
    newConditions[index][key] = value;

    // 如果字段变化，检查操作符是否兼容
    if (key === 'field') {
      const isNumeric = numericFields.includes(value);
      const isStatus = statusFields.includes(value);
      const currentOp = newConditions[index].operator;
      const opInfo = operators.find((o) => o.value === currentOp);

      if (isStatus) {
        // 状态字段只能使用 equals 或 not_equals
        if (!['equals', 'not_equals'].includes(currentOp)) {
          newConditions[index].operator = 'equals';
        }
        // 清空值，让用户从下拉框选择
        newConditions[index].value = '';
      } else if (isNumeric && opInfo?.type === 'string' && !['equals', 'not_equals'].includes(currentOp)) {
        newConditions[index].operator = 'greater_than';
      } else if (!isNumeric && opInfo?.type === 'number') {
        newConditions[index].operator = 'equals';
      }
    }

    setConditions(newConditions);
  };

  const handleSave = () => {
    if (!name.trim() || !tagName) return;
    const conditionsJson = JSON.stringify({ logic, conditions });
    onSave({
      name: name.trim(),
      tagName: tagName,
      enabled,
      triggerType,
      conditions: conditionsJson
    });
  };

  const getAvailableOperators = (field) => {
    const isNumeric = numericFields.includes(field);
    const isStatus = statusFields.includes(field);

    if (isStatus) {
      // 状态字段只支持等于和不等于
      return operators.filter((o) => ['equals', 'not_equals'].includes(o.value));
    }
    if (isNumeric) {
      return operators;
    }
    return operators.filter((o) => o.type === 'string');
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth fullScreen={isMobile}>
      <DialogTitle>{editingRule ? '编辑规则' : '添加规则'}</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 1, display: 'flex', flexDirection: 'column', gap: 2 }}>
          {/* 基本信息 */}
          <Box sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', gap: 2 }}>
            <TextField label="规则名称" value={name} onChange={(e) => setName(e.target.value)} fullWidth required />
            <FormControl sx={{ minWidth: isMobile ? '100%' : 150 }} fullWidth={isMobile}>
              <InputLabel>关联标签</InputLabel>
              <Select value={tagName} label="关联标签" onChange={(e) => setTagName(e.target.value)}>
                {tags.map((tag) => (
                  <MenuItem key={tag.name} value={tag.name}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Box sx={{ width: 12, height: 12, borderRadius: '50%', backgroundColor: tag.color }} />
                      {tag.name}
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>

          <Box
            sx={{
              display: 'flex',
              flexDirection: isMobile ? 'column' : 'row',
              gap: 2,
              alignItems: isMobile ? 'stretch' : 'center'
            }}
          >
            <FormControl sx={{ minWidth: isMobile ? '100%' : 150 }} fullWidth={isMobile}>
              <InputLabel>触发时机</InputLabel>
              <Select value={triggerType} label="触发时机" onChange={(e) => setTriggerType(e.target.value)}>
                <MenuItem value="subscription_update">订阅更新后</MenuItem>
                <MenuItem value="speed_test">测速完成后</MenuItem>
              </Select>
            </FormControl>
            <FormControlLabel control={<Switch checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />} label="启用规则" />
          </Box>

          <Divider sx={{ my: 1 }} />

          {/* 条件配置 */}
          <Box>
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                mb: 2,
                flexWrap: 'wrap',
                gap: 1
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
                <Typography variant="subtitle1">匹配条件</Typography>
                <Chip
                  label={logic === 'and' ? '全部满足 (AND)' : '任一满足 (OR)'}
                  size="small"
                  onClick={() => setLogic(logic === 'and' ? 'or' : 'and')}
                  sx={{ cursor: 'pointer' }}
                />
              </Box>
              <Button size="small" startIcon={<AddIcon />} onClick={handleAddCondition}>
                添加条件
              </Button>
            </Box>

            {conditions.map((cond, index) =>
              isMobile ? (
                // 移动端：卡片式布局
                <Card key={index} variant="outlined" sx={{ mb: 1.5, p: 1.5, position: 'relative' }}>
                  <IconButton
                    size="small"
                    color="error"
                    onClick={() => handleRemoveCondition(index)}
                    disabled={conditions.length === 1}
                    sx={{ position: 'absolute', top: 8, right: 8 }}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5, pr: 4 }}>
                    <FormControl size="small" fullWidth>
                      <InputLabel>字段</InputLabel>
                      <Select value={cond.field} label="字段" onChange={(e) => handleConditionChange(index, 'field', e.target.value)}>
                        {nodeFields.map((f) => (
                          <MenuItem key={f.value} value={f.value}>
                            {f.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <FormControl size="small" fullWidth>
                      <InputLabel>操作符</InputLabel>
                      <Select
                        value={cond.operator}
                        label="操作符"
                        onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}
                      >
                        {getAvailableOperators(cond.field).map((op) => (
                          <MenuItem key={op.value} value={op.value}>
                            {op.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    {statusFields.includes(cond.field) ? (
                      <FormControl size="small" fullWidth>
                        <InputLabel>值</InputLabel>
                        <Select value={cond.value} label="值" onChange={(e) => handleConditionChange(index, 'value', e.target.value)}>
                          {statusOptions.map((opt) => (
                            <MenuItem key={opt.value} value={opt.value}>
                              {opt.label}
                            </MenuItem>
                          ))}
                        </Select>
                      </FormControl>
                    ) : (
                      <TextField
                        size="small"
                        label="值"
                        value={cond.value}
                        onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                        fullWidth
                        type={numericFields.includes(cond.field) ? 'number' : 'text'}
                        placeholder={cond.operator === 'regex' ? '正则表达式' : ''}
                      />
                    )}
                  </Box>
                </Card>
              ) : (
                // 桌面端：水平布局
                <Box key={index} sx={{ display: 'flex', gap: 1, mb: 1.5, alignItems: 'center' }}>
                  <FormControl size="small" sx={{ minWidth: 140 }}>
                    <InputLabel>字段</InputLabel>
                    <Select value={cond.field} label="字段" onChange={(e) => handleConditionChange(index, 'field', e.target.value)}>
                      {nodeFields.map((f) => (
                        <MenuItem key={f.value} value={f.value}>
                          {f.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <FormControl size="small" sx={{ minWidth: 120 }}>
                    <InputLabel>操作符</InputLabel>
                    <Select value={cond.operator} label="操作符" onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}>
                      {getAvailableOperators(cond.field).map((op) => (
                        <MenuItem key={op.value} value={op.value}>
                          {op.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  {statusFields.includes(cond.field) ? (
                    <FormControl size="small" sx={{ minWidth: 140 }}>
                      <InputLabel>值</InputLabel>
                      <Select value={cond.value} label="值" onChange={(e) => handleConditionChange(index, 'value', e.target.value)}>
                        {statusOptions.map((opt) => (
                          <MenuItem key={opt.value} value={opt.value}>
                            {opt.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  ) : (
                    <TextField
                      size="small"
                      label="值"
                      value={cond.value}
                      onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                      sx={{ flex: 1 }}
                      type={numericFields.includes(cond.field) ? 'number' : 'text'}
                      placeholder={cond.operator === 'regex' ? '正则表达式' : ''}
                    />
                  )}
                  <IconButton size="small" color="error" onClick={() => handleRemoveCondition(index)} disabled={conditions.length === 1}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Box>
              )
            )}

            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 1 }}>
              示例：国家代码 等于 CN — 匹配所有中国节点
            </Typography>
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={handleSave} disabled={!name.trim() || !tagName}>
          保存
        </Button>
      </DialogActions>
    </Dialog>
  );
}

RuleDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  editingRule: PropTypes.object,
  tags: PropTypes.array.isRequired
};
