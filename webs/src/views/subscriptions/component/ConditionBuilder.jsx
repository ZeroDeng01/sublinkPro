import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@mui/material/styles';
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
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import {
  getChainProxyFieldControlSx,
  getChainProxyIconButtonSx,
  getChainProxyThemeTokens,
  getChainProxyToggleButtonGroupSx
} from './chainProxyTheme';
import {
  getNodeConditionFieldMeta,
  getNodeConditionValueOptions,
  isNodeConditionNumericField,
  isNodeConditionSelectField
} from '../../../utils/nodeConditionOptions';

export default function ConditionBuilder({ value, onChange, fields = [], operators = [], title }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const fieldControlSx = getChainProxyFieldControlSx(tokens);
  const logicToggleSx = getChainProxyToggleButtonGroupSx(tokens);
  const removeButtonSx = getChainProxyIconButtonSx(tokens, theme.palette.error.main);
  const palette = tokens.palette;
  const [logic, setLogic] = useState(value?.logic || 'and');
  const [conditions, setConditions] = useState(value?.conditions || []);

  useEffect(() => {
    if (value) {
      setLogic(value.logic || 'and');
      setConditions(value.conditions || []);
    }
  }, [value]);

  const notifyChange = (newLogic, newConditions) => {
    onChange?.({
      logic: newLogic,
      conditions: newConditions
    });
  };

  const handleLogicChange = (_event, newLogic) => {
    if (newLogic !== null) {
      setLogic(newLogic);
      notifyChange(newLogic, conditions);
    }
  };

  const handleAddCondition = () => {
    const newConditions = [...conditions, { field: fields[0]?.value || '', operator: 'contains', value: '' }];
    setConditions(newConditions);
    notifyChange(logic, newConditions);
  };

  const handleRemoveCondition = (index) => {
    const newConditions = conditions.filter((_, i) => i !== index);
    setConditions(newConditions);
    notifyChange(logic, newConditions);
  };

  const handleConditionChange = (index, field, newValue) => {
    const newConditions = [...conditions];
    newConditions[index] = { ...newConditions[index], [field]: newValue };

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
    return operators.filter((op) => ['equals', 'not_equals', 'contains', 'not_contains', 'regex'].includes(op.value));
  };

  return (
    <Box
      sx={{
        p: 2,
        border: `1px solid ${tokens.panelBorder}`,
        borderRadius: 2,
        backgroundColor: tokens.containerSurface,
        backgroundImage: tokens.dialogSurfaceGradient,
        backdropFilter: 'blur(8px)',
        boxShadow: tokens.panelShadow
      }}
    >
      <Stack spacing={2}>
        <Stack direction="row" alignItems="center" justifyContent="space-between" flexWrap="wrap" gap={1}>
          <Typography variant="subtitle2" color={tokens.secondaryText}>
            {title || t('subscriptions.conditions.title')}
          </Typography>
          <ToggleButtonGroup value={logic} exclusive onChange={handleLogicChange} size="small" sx={logicToggleSx}>
            <ToggleButton value="and">{t('subscriptions.conditions.logic.and')}</ToggleButton>
            <ToggleButton value="or">{t('subscriptions.conditions.logic.or')}</ToggleButton>
          </ToggleButtonGroup>
        </Stack>

        {conditions.map((condition, index) => (
          <Stack
            key={index}
            direction="row"
            spacing={1}
            alignItems="center"
            flexWrap="wrap"
            sx={{
              p: 1.25,
              borderRadius: 1.5,
              backgroundColor: tokens.mutedPanelSurface,
              border: `1px solid ${tokens.softBorder}`,
              boxShadow: tokens.insetHighlight
            }}
          >
            <FormControl size="small" sx={{ minWidth: 100, ...fieldControlSx }}>
              <InputLabel color="primary">{t('subscriptions.conditions.field')}</InputLabel>
              <Select
                value={condition.field}
                label={t('subscriptions.conditions.field')}
                onChange={(e) => handleConditionChange(index, 'field', e.target.value)}
              >
                {fields.map((field) => (
                  <MenuItem key={field.value} value={field.value}>
                    {field.labelKey ? t(field.labelKey, field.label) : field.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl size="small" sx={{ minWidth: 90, ...fieldControlSx }}>
              <InputLabel color="primary">{t('subscriptions.conditions.operator')}</InputLabel>
              <Select
                value={condition.operator}
                label={t('subscriptions.conditions.operator')}
                onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}
              >
                {getOperatorsForField(condition.field).map((op) => (
                  <MenuItem key={op.value} value={op.value}>
                    {op.labelKey ? t(op.labelKey, op.label) : op.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {getNodeConditionValueOptions(condition.field) ? (
              <FormControl size="small" sx={{ flex: 1, minWidth: 100, ...fieldControlSx }}>
                <InputLabel color="primary">{t('subscriptions.conditions.value')}</InputLabel>
                <Select
                  value={condition.value}
                  label={t('subscriptions.conditions.value')}
                  onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                >
                  {getNodeConditionValueOptions(condition.field).map((opt) => (
                    <MenuItem key={opt.value} value={opt.value}>
                      {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            ) : (
              <TextField
                size="small"
                label={t('subscriptions.conditions.value')}
                value={condition.value}
                onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                type={isNodeConditionNumericField(condition.field) ? 'number' : 'text'}
                sx={{ flex: 1, minWidth: 100, ...fieldControlSx }}
              />
            )}

            <IconButton
              size="small"
              onClick={() => handleRemoveCondition(index)}
              sx={{ ...removeButtonSx, color: theme.palette.error.main }}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Stack>
        ))}

        <Button
          startIcon={<AddIcon />}
          size="small"
          onClick={handleAddCondition}
          sx={{
            alignSelf: 'flex-start',
            color: palette.primary.main,
            borderColor: tokens.primarySoftBorder,
            backgroundColor: tokens.fieldSurface,
            boxShadow: tokens.insetHighlight,
            '&:hover': {
              bgcolor: tokens.hoverSurface,
              borderColor: tokens.primaryStrongBorder
            }
          }}
          variant="outlined"
        >
          {t('subscriptions.conditions.add')}
        </Button>

        {conditions.length === 0 && (
          <Typography variant="body2" color={tokens.secondaryText} sx={{ fontStyle: 'italic' }}>
            {t('subscriptions.conditions.empty')}
          </Typography>
        )}
      </Stack>
    </Box>
  );
}
