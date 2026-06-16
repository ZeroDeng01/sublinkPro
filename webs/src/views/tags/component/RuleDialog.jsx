import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
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

// project imports
import UnlockConditionInput from './UnlockConditionInput';
import {
  getNodeConditionFieldMeta,
  getNodeConditionFields,
  getNodeConditionValueOptions,
  isNodeConditionNumericField,
  isNodeConditionSelectField
} from '../../../utils/nodeConditionOptions';
import { UNLOCK_CONDITION_FIELD, convertBackendToUI, convertUIToBackend } from '../../../utils/unlockConditionConverter';

const operators = [
  { value: 'equals', labelKey: 'tags.dialog.rule.operators.equals', type: 'string' },
  { value: 'not_equals', labelKey: 'tags.dialog.rule.operators.notEquals', type: 'string' },
  { value: 'contains', labelKey: 'tags.dialog.rule.operators.contains', type: 'string' },
  { value: 'not_contains', labelKey: 'tags.dialog.rule.operators.notContains', type: 'string' },
  { value: 'regex', labelKey: 'tags.dialog.rule.operators.regex', type: 'string' },
  { value: 'greater_than', labelKey: 'tags.dialog.rule.operators.greaterThan', type: 'number' },
  { value: 'less_than', labelKey: 'tags.dialog.rule.operators.lessThan', type: 'number' },
  { value: 'greater_or_equal', labelKey: 'tags.dialog.rule.operators.greaterOrEqual', type: 'number' },
  { value: 'less_or_equal', labelKey: 'tags.dialog.rule.operators.lessOrEqual', type: 'number' }
];

export default function RuleDialog({ open, onClose, onSave, editingRule, tags }) {
  const { t } = useTranslation();
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
        // 转换后端格式到UI格式（合并解锁字段）
        const backendConditions =
          parsed.conditions?.length > 0 ? parsed.conditions : [{ field: 'link_country', operator: 'equals', value: '' }];
        const uiConditions = convertBackendToUI(backendConditions);
        setConditions(uiConditions);
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
    const previousField = newConditions[index].field;
    newConditions[index][key] = value;

    if (key === 'field') {
      const isNumeric = isNodeConditionNumericField(value);
      const isSelectField = isNodeConditionSelectField(value);
      const fieldMeta = getNodeConditionFieldMeta(value);
      const allowedOperators = fieldMeta?.operators || [];
      const currentOp = newConditions[index].operator;
      const opInfo = operators.find((o) => o.value === currentOp);

      // 从解锁情况切换到其他字段，或从其他字段切换到解锁情况时，需要重置 value
      const wasPreviousUnlock = previousField === UNLOCK_CONDITION_FIELD;
      const isCurrentUnlock = value === UNLOCK_CONDITION_FIELD;

      // 处理解锁情况虚拟字段
      if (isCurrentUnlock) {
        newConditions[index].operator = 'equals';
        newConditions[index].value = {
          provider: '',
          status: '',
          keyword: '',
          providerOperator: 'equals',
          statusOperator: 'equals',
          keywordOperator: 'contains'
        };
      } else if (wasPreviousUnlock || isSelectField) {
        // 从解锁情况切换出来，或切换到下拉字段，重置为空字符串
        if (!allowedOperators.includes(currentOp)) {
          newConditions[index].operator = allowedOperators[0] || 'equals';
        }
        newConditions[index].value = '';
      } else if (isNumeric && opInfo?.type === 'string' && !['equals', 'not_equals'].includes(currentOp)) {
        newConditions[index].operator = allowedOperators[0] || 'greater_than';
      } else if (!isNumeric && opInfo?.type === 'number') {
        newConditions[index].operator = allowedOperators[0] || 'equals';
      }
    }

    setConditions(newConditions);
  };

  // 处理解锁情况复合值的变化
  const handleUnlockConditionChange = (index, subField, subValue) => {
    const newConditions = [...conditions];
    if (!newConditions[index].value || typeof newConditions[index].value !== 'object') {
      newConditions[index].value = {
        provider: '',
        status: '',
        keyword: '',
        providerOperator: 'equals',
        statusOperator: 'equals',
        keywordOperator: 'contains'
      };
    }
    // subField 为 'value' 时，subValue 是整个对象
    if (subField === 'value') {
      newConditions[index].value = subValue;
    } else {
      newConditions[index].value = {
        ...newConditions[index].value,
        [subField]: subValue
      };
    }
    setConditions(newConditions);
  };

  const handleSave = () => {
    if (!name.trim() || !tagName) return;

    // 转换UI格式到后端格式（展开解锁条件）
    const backendConditions = convertUIToBackend(conditions);
    const conditionsJson = JSON.stringify({ logic, conditions: backendConditions });

    onSave({
      name: name.trim(),
      tagName: tagName,
      enabled,
      triggerType,
      conditions: conditionsJson
    });
  };

  const getAvailableOperators = (field) => {
    const fieldMeta = getNodeConditionFieldMeta(field);
    if (Array.isArray(fieldMeta?.operators) && fieldMeta.operators.length > 0) {
      return operators.filter((o) => fieldMeta.operators.includes(o.value));
    }
    const isNumeric = isNodeConditionNumericField(field);
    const isSelectField = isNodeConditionSelectField(field);
    if (isSelectField) {
      return operators.filter((o) => ['equals', 'not_equals'].includes(o.value));
    }
    if (isNumeric) {
      return operators;
    }
    return operators.filter((o) => o.type === 'string');
  };

  const getVisibleFields = () => {
    const hiddenFields = ['unlock_provider', 'unlock_status', 'unlock_keyword'];
    return getNodeConditionFields().filter((f) => !f.hidden && !hiddenFields.includes(f.value));
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth fullScreen={isMobile}>
      <DialogTitle>{editingRule ? t('tags.dialog.rule.editTitle') : t('tags.dialog.rule.addTitle')}</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 1, display: 'flex', flexDirection: 'column', gap: 2 }}>
          <Box sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', gap: 2 }}>
            <TextField label={t('tags.fields.ruleName')} value={name} onChange={(e) => setName(e.target.value)} fullWidth required />
            <FormControl sx={{ minWidth: isMobile ? '100%' : 150 }} fullWidth={isMobile}>
              <InputLabel>{t('tags.fields.tag')}</InputLabel>
              <Select value={tagName} label={t('tags.fields.tag')} onChange={(e) => setTagName(e.target.value)}>
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
              <InputLabel>{t('tags.fields.trigger')}</InputLabel>
              <Select value={triggerType} label={t('tags.fields.trigger')} onChange={(e) => setTriggerType(e.target.value)}>
                <MenuItem value="subscription_update">{t('tags.trigger.subscriptionUpdate')}</MenuItem>
                <MenuItem value="speed_test">{t('tags.trigger.speedTest')}</MenuItem>
              </Select>
            </FormControl>
            <FormControlLabel
              control={<Switch checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />}
              label={t('tags.dialog.rule.enableRule')}
            />
          </Box>

          <Divider sx={{ my: 1 }} />

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
                <Typography variant="subtitle1">{t('tags.dialog.rule.conditions')}</Typography>
                <Chip
                  label={logic === 'and' ? t('tags.dialog.rule.logic.and') : t('tags.dialog.rule.logic.or')}
                  size="small"
                  onClick={() => setLogic(logic === 'and' ? 'or' : 'and')}
                  sx={{ cursor: 'pointer' }}
                />
              </Box>
              <Button size="small" startIcon={<AddIcon />} onClick={handleAddCondition}>
                {t('tags.dialog.rule.addCondition')}
              </Button>
            </Box>

            {conditions.map((cond, index) =>
              isMobile ? (
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
                      <InputLabel>{t('tags.dialog.rule.field')}</InputLabel>
                      <Select
                        value={cond.field}
                        label={t('tags.dialog.rule.field')}
                        onChange={(e) => handleConditionChange(index, 'field', e.target.value)}
                      >
                        {getVisibleFields().map((f) => (
                          <MenuItem key={f.value} value={f.value}>
                            {f.labelKey ? t(f.labelKey, f.label) : f.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <FormControl size="small" fullWidth>
                      <InputLabel>{t('tags.dialog.rule.operator')}</InputLabel>
                      <Select
                        value={cond.operator}
                        label={t('tags.dialog.rule.operator')}
                        onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}
                        disabled={cond.field === UNLOCK_CONDITION_FIELD}
                      >
                        {getAvailableOperators(cond.field).map((op) => (
                          <MenuItem key={op.value} value={op.value}>
                            {t(op.labelKey)}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    {cond.field === UNLOCK_CONDITION_FIELD ? (
                      <UnlockConditionInput
                        value={cond.value}
                        onChange={(newValue) => handleUnlockConditionChange(index, 'value', newValue)}
                        isMobile={true}
                      />
                    ) : getNodeConditionValueOptions(cond.field) ? (
                      <FormControl size="small" fullWidth>
                        <InputLabel>{t('tags.dialog.rule.value')}</InputLabel>
                        <Select
                          value={cond.value}
                          label={t('tags.dialog.rule.value')}
                          onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                        >
                          {getNodeConditionValueOptions(cond.field).map((opt) => (
                            <MenuItem key={opt.value} value={opt.value}>
                              {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
                            </MenuItem>
                          ))}
                        </Select>
                      </FormControl>
                    ) : (
                      <TextField
                        size="small"
                        label={t('tags.dialog.rule.value')}
                        value={cond.value}
                        onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                        fullWidth
                        type={isNodeConditionNumericField(cond.field) ? 'number' : 'text'}
                        placeholder={cond.operator === 'regex' ? t('tags.dialog.rule.regexPlaceholder') : ''}
                      />
                    )}
                  </Box>
                </Card>
              ) : (
                <Box key={index} sx={{ display: 'flex', gap: 1, mb: 1.5, alignItems: 'center' }}>
                  <FormControl size="small" sx={{ minWidth: 140 }}>
                    <InputLabel>{t('tags.dialog.rule.field')}</InputLabel>
                    <Select
                      value={cond.field}
                      label={t('tags.dialog.rule.field')}
                      onChange={(e) => handleConditionChange(index, 'field', e.target.value)}
                    >
                      {getVisibleFields().map((f) => (
                        <MenuItem key={f.value} value={f.value}>
                          {f.labelKey ? t(f.labelKey, f.label) : f.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <FormControl size="small" sx={{ minWidth: 120 }}>
                    <InputLabel>{t('tags.dialog.rule.operator')}</InputLabel>
                    <Select
                      value={cond.operator}
                      label={t('tags.dialog.rule.operator')}
                      onChange={(e) => handleConditionChange(index, 'operator', e.target.value)}
                      disabled={cond.field === UNLOCK_CONDITION_FIELD}
                    >
                      {getAvailableOperators(cond.field).map((op) => (
                        <MenuItem key={op.value} value={op.value}>
                          {t(op.labelKey)}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  {cond.field === UNLOCK_CONDITION_FIELD ? (
                    <UnlockConditionInput
                      value={cond.value}
                      onChange={(newValue) => handleUnlockConditionChange(index, 'value', newValue)}
                      isMobile={false}
                    />
                  ) : getNodeConditionValueOptions(cond.field) ? (
                    <FormControl size="small" sx={{ minWidth: 140 }}>
                      <InputLabel>{t('tags.dialog.rule.value')}</InputLabel>
                      <Select
                        value={cond.value}
                        label={t('tags.dialog.rule.value')}
                        onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                      >
                        {getNodeConditionValueOptions(cond.field).map((opt) => (
                          <MenuItem key={opt.value} value={opt.value}>
                            {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  ) : (
                    <TextField
                      size="small"
                      label={t('tags.dialog.rule.value')}
                      value={cond.value}
                      onChange={(e) => handleConditionChange(index, 'value', e.target.value)}
                      sx={{ flex: 1 }}
                      type={isNodeConditionNumericField(cond.field) ? 'number' : 'text'}
                      placeholder={cond.operator === 'regex' ? t('tags.dialog.rule.regexPlaceholder') : ''}
                    />
                  )}
                  <IconButton size="small" color="error" onClick={() => handleRemoveCondition(index)} disabled={conditions.length === 1}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Box>
              )
            )}

            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 1 }}>
              {t('tags.dialog.rule.example')}
            </Typography>
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={handleSave} disabled={!name.trim() || !tagName}>
          {t('common.save')}
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
