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
export default function ConditionBuilder({
    value,
    onChange,
    fields = [],
    operators = [],
    title = '条件配置'
}) {
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
        const newConditions = [
            ...conditions,
            { field: fields[0]?.value || '', operator: 'contains', value: '' }
        ];
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
        const newConditions = conditions.map((cond, i) => {
            if (i === index) {
                return { ...cond, [field]: newValue };
            }
            return cond;
        });
        setConditions(newConditions);
        notifyChange(logic, newConditions);
    };

    // 获取字段对应的操作符列表
    const getOperatorsForField = (fieldValue) => {
        const field = fields.find((f) => f.value === fieldValue);
        // 数值字段支持比较操作符
        const numericFields = ['speed', 'delay_time'];
        if (numericFields.includes(fieldValue)) {
            return operators;
        }
        // 文本字段只支持字符串操作符
        return operators.filter((op) =>
            ['equals', 'not_equals', 'contains', 'not_contains', 'regex'].includes(op.value)
        );
    };

    return (
        <Box sx={{
            p: 2,
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 1,
            backgroundColor: 'background.default'
        }}>
            <Stack spacing={2}>
                {/* 标题和逻辑切换 */}
                <Stack direction="row" alignItems="center" justifyContent="space-between">
                    <Typography variant="subtitle2" color="text.secondary">
                        {title}
                    </Typography>
                    <ToggleButtonGroup
                        value={logic}
                        exclusive
                        onChange={handleLogicChange}
                        size="small"
                        color="primary"
                    >
                        <ToggleButton value="and">全部满足 (AND)</ToggleButton>
                        <ToggleButton value="or">满足任一 (OR)</ToggleButton>
                    </ToggleButtonGroup>
                </Stack>

                {/* 条件列表 */}
                {conditions.map((condition, index) => (
                    <Stack key={index} direction="row" spacing={1} alignItems="center">
                        <FormControl size="small" sx={{ minWidth: 120 }}>
                            <InputLabel>字段</InputLabel>
                            <Select
                                value={condition.field}
                                label="字段"
                                onChange={(e) => handleConditionChange(index, 'field', e.target.value)}
                            >
                                {fields.map((field) => (
                                    <MenuItem key={field.value} value={field.value}>
                                        {field.label}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <FormControl size="small" sx={{ minWidth: 120 }}>
                            <InputLabel>操作</InputLabel>
                            <Select
                                value={condition.operator}
                                label="操作"
                                onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}
                            >
                                {getOperatorsForField(condition.field).map((op) => (
                                    <MenuItem key={op.value} value={op.value}>
                                        {op.label}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <TextField
                            size="small"
                            label="值"
                            value={condition.value}
                            onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                            sx={{ flex: 1 }}
                        />

                        <IconButton
                            size="small"
                            color="error"
                            onClick={() => handleRemoveCondition(index)}
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
                    sx={{ alignSelf: 'flex-start' }}
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
