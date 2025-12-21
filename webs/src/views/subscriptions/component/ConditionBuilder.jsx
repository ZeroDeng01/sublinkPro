import { useState, useEffect } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import IconButton from '@mui/material/IconButton';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Typography from '@mui/material/Typography';
// Paper 组件已改为 Box
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';

/**
 * 通用条件构建器组件
 * 用于构建 AND/OR 组合的条件表达式
 */
export default function ConditionBuilder({ value, onChange, fields = [], operators = [], title = '条件配置' }) {
  // 定义特殊字段类型
  const numericFields = ['speed', 'delay_time'];
  const statusFields = ['speed_status', 'delay_status'];

  // 状态选项（与 RuleDialog.jsx 保持一致）
  const STATUS_OPTIONS = [
    { value: 'untested', label: '未测速' },
    { value: 'success', label: '成功' },
    { value: 'timeout', label: '超时' },
    { value: 'error', label: '失败' }
  ];

  // 初始化条件数据
  const [logic, setLogic] = useState(value?.logic || 'and');
  const [conditions, setConditions] = useState(value?.conditions || []);

  // 当外部 value 变化时更新内部状态
  useEffect(() => {
    if (value) {
      setLogic(value.logic || 'and');
      setConditions(value.conditions || []);
    }
  }, [value]);

  // 通知父组件数据变化
  const notifyChange = (newLogic, newConditions) => {
    onChange?.({
      logic: newLogic,
      conditions: newConditions
    });
  };

  // 切换逻辑运算符
  const handleLogicChange = (event, newLogic) => {
    if (newLogic !== null) {
      setLogic(newLogic);
      notifyChange(newLogic, conditions);
    }
  };

  // 添加条件
  const handleAddCondition = () => {
    const newConditions = [...conditions, { field: fields[0]?.value || '', operator: 'contains', value: '' }];
    setConditions(newConditions);
    notifyChange(logic, newConditions);
  };

  // 删除条件
  const handleRemoveCondition = (index) => {
    const newConditions = conditions.filter((_, i) => i !== index);
    setConditions(newConditions);
    notifyChange(logic, newConditions);
  };

  // 更新条件字段
  const handleConditionChange = (index, field, newValue) => {
    const newConditions = [...conditions];
    newConditions[index] = { ...newConditions[index], [field]: newValue };

    // 如果改变了字段，需要重置操作符和值以避免不兼容
    if (field === 'field') {
      const isStatus = statusFields.includes(newValue);
      const isNumeric = numericFields.includes(newValue);
      const currentOp = newConditions[index].operator;

      if (isStatus) {
        // 状态字段只能用 equals 或 not_equals
        if (!['equals', 'not_equals'].includes(currentOp)) {
          newConditions[index].operator = 'equals';
        }
        // 清空值，强制用户从下拉框选择
        newConditions[index].value = '';
      } else if (isNumeric) {
        // 数值字段默认使用 greater_than 如果当前操作符不兼容
        // 注意：这里我们假设 operators 包含了所有类型的操作符，具体逻辑可能需要根据 operators 的额外信息判断
        // 但为了简化，如果当前是 regex 或 contains 等字符串专用的，切换到 greater_than
        if (['contains', 'not_contains', 'regex'].includes(currentOp)) {
          newConditions[index].operator = 'greater_than';
        } else if (!currentOp) {
          newConditions[index].operator = 'greater_than';
        }
      } else {
        // 其他字段（字符串），如果是数值专用操作符，切换回 contains
        if (['greater_than', 'less_than', 'greater_or_equal', 'less_or_equal'].includes(currentOp)) {
          newConditions[index].operator = 'contains';
        }
      }
    }

    setConditions(newConditions);
    notifyChange(logic, newConditions);
  };

  // 获取字段对应的操作符列表
  const getOperatorsForField = (fieldValue) => {
    if (statusFields.includes(fieldValue)) {
      return operators.filter((op) => ['equals', 'not_equals'].includes(op.value));
    }
    if (numericFields.includes(fieldValue)) {
      return operators;
    }
    // 文本字段只支持字符串操作符
    return operators.filter((op) => ['equals', 'not_equals', 'contains', 'not_contains', 'regex'].includes(op.value));
  };

  return (
    <Box
      sx={{
        p: 2,
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 1,
        backgroundColor: 'background.default'
      }}
    >
      <Stack spacing={2}>
        {/* 标题和逻辑切换 */}
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="subtitle2" color="text.secondary">
            {title}
          </Typography>
          <ToggleButtonGroup value={logic} exclusive onChange={handleLogicChange} size="small" color="primary">
            <ToggleButton value="and">全部满足 (AND)</ToggleButton>
            <ToggleButton value="or">满足任一 (OR)</ToggleButton>
          </ToggleButtonGroup>
        </Stack>

        {/* 条件列表 */}
        {conditions.map((condition, index) => (
          <Stack key={index} direction="row" spacing={1} alignItems="center">
            <FormControl size="small" sx={{ minWidth: 120 }}>
              <InputLabel>字段</InputLabel>
              <Select value={condition.field} label="字段" onChange={(e) => handleConditionChange(index, 'field', e.target.value)}>
                {fields.map((field) => (
                  <MenuItem key={field.value} value={field.value}>
                    {field.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl size="small" sx={{ minWidth: 120 }}>
              <InputLabel>操作</InputLabel>
              <Select value={condition.operator} label="操作" onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}>
                {getOperatorsForField(condition.field).map((op) => (
                  <MenuItem key={op.value} value={op.value}>
                    {op.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {statusFields.includes(condition.field) ? (
              <FormControl size="small" sx={{ flex: 1 }}>
                <InputLabel>状态值</InputLabel>
                <Select value={condition.value} label="状态值" onChange={(e) => handleConditionChange(index, 'value', e.target.value)}>
                  {STATUS_OPTIONS.map((opt) => (
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
                value={condition.value}
                onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                sx={{ flex: 1 }}
                type={numericFields.includes(condition.field) ? 'number' : 'text'}
              />
            )}

            <IconButton size="small" color="error" onClick={() => handleRemoveCondition(index)}>
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Stack>
        ))}

        {/* 添加条件按钮 */}
        <Button startIcon={<AddIcon />} size="small" onClick={handleAddCondition} sx={{ alignSelf: 'flex-start' }}>
          添加条件
        </Button>

        {/* 空状态提示 */}
        {conditions.length === 0 && (
          <Typography variant="body2" color="text.secondary" sx={{ fontStyle: 'italic' }}>
            尚未添加任何条件
          </Typography>
        )}
      </Stack>
    </Box>
  );
}
