import { useState, useEffect } from 'react';
import { alpha, useTheme } from '@mui/material/styles';
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
import { withAlpha } from '../../../utils/colorUtils';
import {
  getNodeConditionFieldMeta,
  getNodeConditionValueOptions,
  isNodeConditionNumericField,
  isNodeConditionSelectField
} from '../../../utils/nodeConditionOptions';

/**
 * 通用条件构建器组件
 * 用于构建 AND/OR 组合的条件表达式
 */
export default function ConditionBuilder({ value, onChange, fields = [], operators = [], title = '条件配置' }) {
  const theme = useTheme();
  const palette = theme.vars?.palette || theme.palette;
  const isDark = theme.palette.mode === 'dark';
  const panelBorder = isDark ? withAlpha(palette.divider, 0.78) : withAlpha(palette.divider, 0.9);
  const containerSurface = isDark ? withAlpha(palette.background.paper, 0.42) : palette.background.paper;
  const fieldSurface = isDark ? withAlpha(palette.background.default, 0.72) : palette.background.default;

  const fieldControlSx = {
    '& .MuiInputLabel-root': { color: 'text.secondary' },
    '& .MuiInputLabel-root.Mui-focused': { color: 'primary.main' },
    '& .MuiOutlinedInput-root': {
      backgroundColor: fieldSurface,
      boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}` : 'none',
      '& .MuiOutlinedInput-notchedOutline': { borderColor: panelBorder },
      '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: alpha(theme.palette.primary.main, 0.4) },
      '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: theme.palette.primary.main }
    },
    '& .MuiSelect-select': { color: 'text.primary' },
    '& .MuiInputBase-input': { color: 'text.primary' },
    '& .MuiSelect-icon': { color: 'text.secondary' }
  };
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
      const nextFieldMeta = getNodeConditionFieldMeta(newValue);
      const isSelectField = isNodeConditionSelectField(newValue);
      const isNumeric = isNodeConditionNumericField(newValue);
      const currentOp = newConditions[index].operator;
      const allowedOperators = nextFieldMeta?.operators || [];

      if (isSelectField) {
        if (!allowedOperators.includes(currentOp)) {
          newConditions[index].operator = allowedOperators[0] || 'equals';
        }
        newConditions[index].value = '';
      } else if (isNumeric) {
        if (!allowedOperators.includes(currentOp)) {
          newConditions[index].operator = allowedOperators[0] || 'greater_than';
        }
      } else {
        if (!allowedOperators.includes(currentOp)) {
          newConditions[index].operator = allowedOperators[0] || 'contains';
        }
      }
    }

    setConditions(newConditions);
    notifyChange(logic, newConditions);
  };

  // 获取字段对应的操作符列表
  const getOperatorsForField = (fieldValue) => {
    const fieldMeta = getNodeConditionFieldMeta(fieldValue);
    if (Array.isArray(fieldMeta?.operators) && fieldMeta.operators.length > 0) {
      return operators.filter((op) => fieldMeta.operators.includes(op.value));
    }
    if (isNodeConditionSelectField(fieldValue)) {
      return operators.filter((op) => ['equals', 'not_equals'].includes(op.value));
    }
    if (isNodeConditionNumericField(fieldValue)) {
      return operators;
    }
    // 文本字段只支持字符串操作符
    return operators.filter((op) => ['equals', 'not_equals', 'contains', 'not_contains', 'regex'].includes(op.value));
  };

  return (
    <Box
      sx={{
        p: 2,
        border: `1px solid ${panelBorder}`,
        borderRadius: 2,
        backgroundColor: containerSurface,
        backdropFilter: 'blur(8px)',
        boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}` : 'none'
      }}
    >
      <Stack spacing={2}>
        {/* 标题和逻辑切换 */}
        <Stack direction="row" alignItems="center" justifyContent="space-between" flexWrap="wrap" gap={1}>
          <Typography variant="subtitle2" color="text.secondary">
            {title}
          </Typography>
          <ToggleButtonGroup
            value={logic}
            exclusive
            onChange={handleLogicChange}
            size="small"
            sx={{
              '& .MuiToggleButton-root': {
                color: 'text.secondary',
                backgroundColor: fieldSurface,
                borderColor: panelBorder,
                fontSize: 12,
                py: 0.5,
                '&.Mui-selected': {
                  color: 'primary.main',
                  bgcolor: alpha(theme.palette.primary.main, 0.14),
                  borderColor: alpha(theme.palette.primary.main, 0.42)
                },
                '&:hover': {
                  bgcolor: alpha(theme.palette.primary.main, 0.08)
                }
              }
            }}
          >
            <ToggleButton value="and">全部满足 (AND)</ToggleButton>
            <ToggleButton value="or">满足任一 (OR)</ToggleButton>
          </ToggleButtonGroup>
        </Stack>

        {/* 条件列表 */}
        {conditions.map((condition, index) => (
          <Stack key={index} direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            <FormControl size="small" sx={{ minWidth: 100, ...fieldControlSx }}>
              <InputLabel color="primary">字段</InputLabel>
              <Select value={condition.field} label="字段" onChange={(e) => handleConditionChange(index, 'field', e.target.value)}>
                {fields.map((field) => (
                  <MenuItem key={field.value} value={field.value}>
                    {field.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl size="small" sx={{ minWidth: 90, ...fieldControlSx }}>
              <InputLabel color="primary">操作</InputLabel>
              <Select value={condition.operator} label="操作" onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}>
                {getOperatorsForField(condition.field).map((op) => (
                  <MenuItem key={op.value} value={op.value}>
                    {op.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {getNodeConditionValueOptions(condition.field) ? (
              <FormControl size="small" sx={{ flex: 1, minWidth: 100, ...fieldControlSx }}>
                <InputLabel color="primary">值</InputLabel>
                <Select value={condition.value} label="值" onChange={(e) => handleConditionChange(index, 'value', e.target.value)}>
                  {getNodeConditionValueOptions(condition.field).map((opt) => (
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
                type={isNodeConditionNumericField(condition.field) ? 'number' : 'text'}
                sx={{
                  flex: 1,
                  minWidth: 100,
                  '& .MuiInputLabel-root': { color: theme.palette.text.secondary },
                  '& .MuiOutlinedInput-root': {
                    backgroundColor: fieldSurface,
                    boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}` : 'none',
                    '& .MuiOutlinedInput-notchedOutline': { borderColor: panelBorder },
                    '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: alpha(theme.palette.primary.main, 0.4) },
                    '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: theme.palette.primary.main }
                  },
                  '& .MuiInputBase-input': { color: theme.palette.text.primary }
                }}
              />
            )}

            <IconButton
              size="small"
              onClick={() => handleRemoveCondition(index)}
              sx={{
                color: 'error.light',
                '&:hover': { bgcolor: alpha(theme.palette.error.main, 0.08) }
              }}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Stack>
        ))}

        {/* 添加条件按钮 */}
        <Button
          startIcon={<AddIcon />}
          size="small"
          onClick={handleAddCondition}
          sx={{
            alignSelf: 'flex-start',
            color: 'primary.main',
            borderColor: alpha(theme.palette.primary.main, 0.24),
            '&:hover': { bgcolor: alpha(theme.palette.primary.main, 0.08) }
          }}
        >
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
