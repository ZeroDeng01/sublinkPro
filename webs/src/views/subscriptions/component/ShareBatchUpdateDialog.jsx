import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import RadioGroup from '@mui/material/RadioGroup';
import Radio from '@mui/material/Radio';
import FormControlLabel from '@mui/material/FormControlLabel';

const EXPIRE_TYPE_NEVER = 0;
const EXPIRE_TYPE_DAYS = 1;
const EXPIRE_TYPE_DATETIME = 2;

export default function ShareBatchUpdateDialog({ open, mode, shares, onClose, onSubmit }) {
  const { t } = useTranslation();

  const [formData, setFormData] = useState({
    // 过期设置
    expireType: EXPIRE_TYPE_NEVER,
    expireDays: 30,
    expireAt: '',
    // 启用状态
    enableAction: 'enable' // 'enable' | 'disable' | 'toggle'
  });

  const [errors, setErrors] = useState({});

  // 计算启用/禁用状态统计
  const enabledStats = {
    enabled: shares.filter((s) => s.enabled).length,
    disabled: shares.filter((s) => !s.enabled).length
  };

  const hasLegacy = shares.some((s) => s.is_legacy);

  const handleChange = (field, value) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => ({ ...prev, [field]: null }));
  };

  const validate = () => {
    const newErrors = {};

    if (mode === 'expire') {
      if (formData.expireType === EXPIRE_TYPE_DAYS && formData.expireDays <= 0) {
        newErrors.expireDays = t('subscriptions.share.batch.expireDaysInvalid');
      }

      if (formData.expireType === EXPIRE_TYPE_DATETIME && !formData.expireAt) {
        newErrors.expireAt = t('subscriptions.share.batch.expireDateRequired');
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;

    if (mode === 'expire') {
      const submitData = {
        expireType: formData.expireType,
        expireDays: formData.expireDays,
        expireAt: formData.expireAt
      };
      onSubmit(submitData);
    } else if (mode === 'enabled') {
      onSubmit({ action: formData.enableAction });
    }
  };

  const handleClose = () => {
    setFormData({
      expireType: EXPIRE_TYPE_NEVER,
      expireDays: 30,
      expireAt: '',
      enableAction: 'enable'
    });
    setErrors({});
    onClose();
  };

  // 生成受影响的分享名称列表
  const affectedNames = shares.slice(0, 5).map((s) => s.name);
  const moreCount = shares.length - 5;

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        {mode === 'expire' ? t('subscriptions.share.batch.updateExpire') : t('subscriptions.share.batch.toggleEnabled')}
      </DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          {/* 显示受影响的分享数量 */}
          <Alert severity="info">
            {mode === 'expire'
              ? t('subscriptions.share.batch.willUpdateExpire', { count: shares.length })
              : t('subscriptions.share.batch.willUpdateEnabled', { count: shares.length })}
          </Alert>

          {/* 受影响的分享列表 */}
          <Box>
            <Typography variant="caption" color="textSecondary" gutterBottom>
              {t('subscriptions.share.batch.affectedShares')}:
            </Typography>
            <Typography variant="body2">
              {affectedNames.join(', ')}
              {moreCount > 0 && ` (+${moreCount} ${t('common.more')})`}
            </Typography>
          </Box>

          {/* Legacy 警告 */}
          {hasLegacy && <Alert severity="warning">{t('subscriptions.share.batch.legacyWarning')}</Alert>}

          {/* 模式：更新过期设置 */}
          {mode === 'expire' && (
            <>
              <FormControl fullWidth>
                <InputLabel>{t('subscriptions.share.form.expireType')}</InputLabel>
                <Select
                  value={formData.expireType}
                  label={t('subscriptions.share.form.expireType')}
                  onChange={(e) => handleChange('expireType', e.target.value)}
                >
                  <MenuItem value={EXPIRE_TYPE_NEVER}>{t('subscriptions.share.form.expireTypeNever')}</MenuItem>
                  <MenuItem value={EXPIRE_TYPE_DAYS}>{t('subscriptions.share.form.expireTypeDays')}</MenuItem>
                  <MenuItem value={EXPIRE_TYPE_DATETIME}>{t('subscriptions.share.form.expireTypeDateTime')}</MenuItem>
                </Select>
              </FormControl>

              {formData.expireType === EXPIRE_TYPE_DAYS && (
                <TextField
                  label={t('subscriptions.share.form.expireDays')}
                  type="number"
                  value={formData.expireDays}
                  onChange={(e) => handleChange('expireDays', parseInt(e.target.value) || 0)}
                  error={!!errors.expireDays}
                  helperText={errors.expireDays}
                  inputProps={{ min: 1 }}
                  fullWidth
                />
              )}

              {formData.expireType === EXPIRE_TYPE_DATETIME && (
                <TextField
                  label={t('subscriptions.share.form.expireAt')}
                  type="datetime-local"
                  value={formData.expireAt}
                  onChange={(e) => handleChange('expireAt', e.target.value)}
                  error={!!errors.expireAt}
                  helperText={errors.expireAt}
                  fullWidth
                />
              )}
            </>
          )}

          {/* 模式：切换启用状态 */}
          {mode === 'enabled' && (
            <>
              <Box>
                <Typography variant="body2" color="textSecondary" gutterBottom>
                  {t('subscriptions.share.batch.currentState', { enabled: enabledStats.enabled, disabled: enabledStats.disabled })}
                </Typography>
              </Box>

              <FormControl component="fieldset">
                <RadioGroup value={formData.enableAction} onChange={(e) => handleChange('enableAction', e.target.value)}>
                  <FormControlLabel value="enable" control={<Radio />} label={t('subscriptions.share.batch.enableAll')} />
                  <FormControlLabel value="disable" control={<Radio />} label={t('subscriptions.share.batch.disableAll')} />
                  <FormControlLabel value="toggle" control={<Radio />} label={t('subscriptions.share.batch.toggleEach')} />
                </RadioGroup>
              </FormControl>
            </>
          )}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{t('common.cancel')}</Button>
        <Button onClick={handleSubmit} variant="contained" color="primary">
          {t('common.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
