import { useState, useEffect, useMemo } from 'react';
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
import FormControlLabel from '@mui/material/FormControlLabel';
import Checkbox from '@mui/material/Checkbox';
import Alert from '@mui/material/Alert';
import { useTheme } from '@mui/material/styles';

const EXPIRE_TYPE_NEVER = 0;
const EXPIRE_TYPE_DAYS = 1;
const EXPIRE_TYPE_DATETIME = 2;

export default function ShareBatchCreateDialog({ open, subscription, existingNames, onClose, onSubmit }) {
  const { t } = useTranslation();
  const theme = useTheme();

  const [formData, setFormData] = useState({
    baseName: '',
    count: 10,
    expireType: EXPIRE_TYPE_NEVER,
    expireDays: 30,
    expireAt: '',
    enabled: true
  });

  const [errors, setErrors] = useState({});

  // 初始化基础名称为订阅名称
  useEffect(() => {
    if (open && subscription?.Name) {
      setFormData((prev) => ({
        ...prev,
        baseName: subscription.Name
      }));
    }
  }, [open, subscription]);

  // 生成预览名称
  const previewNames = useMemo(() => {
    if (!formData.baseName || formData.count < 1) return '';

    const names = [];
    const showCount = Math.min(formData.count, 5);

    for (let i = 1; i <= showCount; i++) {
      names.push(`${formData.baseName}-${i}`);
    }

    if (formData.count > 5) {
      return `${names.slice(0, 3).join(', ')}, ..., ${formData.baseName}-${formData.count}`;
    }

    return names.join(', ');
  }, [formData.baseName, formData.count]);

  // 检查名称冲突
  const conflictingNames = useMemo(() => {
    if (!formData.baseName || !existingNames) return [];

    const conflicts = [];
    for (let i = 1; i <= formData.count; i++) {
      const name = `${formData.baseName}-${i}`;
      if (existingNames.includes(name)) {
        conflicts.push(name);
      }
    }
    return conflicts;
  }, [formData.baseName, formData.count, existingNames]);

  const handleChange = (field, value) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => ({ ...prev, [field]: null }));
  };

  const validate = () => {
    const newErrors = {};

    if (!formData.baseName || formData.baseName.trim() === '') {
      newErrors.baseName = t('subscriptions.share.batch.baseNameRequired');
    }

    if (formData.count < 1 || formData.count > 100) {
      newErrors.count = t('subscriptions.share.batch.countRange');
    }

    if (formData.expireType === EXPIRE_TYPE_DAYS && formData.expireDays <= 0) {
      newErrors.expireDays = t('subscriptions.share.batch.expireDaysInvalid');
    }

    if (formData.expireType === EXPIRE_TYPE_DATETIME && !formData.expireAt) {
      newErrors.expireAt = t('subscriptions.share.batch.expireDateRequired');
    }

    if (conflictingNames.length > 0) {
      newErrors.conflicts = true;
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;

    const submitData = {
      baseName: formData.baseName.trim(),
      count: formData.count,
      expireType: formData.expireType,
      expireDays: formData.expireDays,
      expireAt: formData.expireAt,
      enabled: formData.enabled
    };

    onSubmit(submitData);
  };

  const handleClose = () => {
    setFormData({
      baseName: '',
      count: 10,
      expireType: EXPIRE_TYPE_NEVER,
      expireDays: 30,
      expireAt: '',
      enabled: true
    });
    setErrors({});
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('subscriptions.share.batch.createTitle')}</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          {/* 基础设置 */}
          <TextField
            label={t('subscriptions.share.batch.baseName')}
            value={formData.baseName}
            onChange={(e) => handleChange('baseName', e.target.value)}
            error={!!errors.baseName}
            helperText={errors.baseName}
            required
            fullWidth
          />

          <TextField
            label={t('subscriptions.share.batch.count')}
            type="number"
            value={formData.count}
            onChange={(e) => handleChange('count', parseInt(e.target.value) || 1)}
            error={!!errors.count}
            helperText={errors.count || t('subscriptions.share.batch.countHint')}
            inputProps={{ min: 1, max: 100 }}
            required
            fullWidth
          />

          {/* 预览 */}
          {previewNames && (
            <Box
              sx={{
                p: 2,
                bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
                borderRadius: 1
              }}
            >
              <Typography variant="caption" color="textSecondary" display="block" gutterBottom>
                {t('subscriptions.share.batch.preview')}:
              </Typography>
              <Typography variant="body2">{previewNames}</Typography>
              <Typography variant="caption" color="textSecondary" sx={{ mt: 1, display: 'block' }}>
                {t('subscriptions.share.batch.totalCount')}: {formData.count}
              </Typography>
            </Box>
          )}

          {/* 名称冲突警告 */}
          {conflictingNames.length > 0 && (
            <Alert severity="error">
              {t('subscriptions.share.batch.nameConflict')}: {conflictingNames.slice(0, 5).join(', ')}
              {conflictingNames.length > 5 && ` (+${conflictingNames.length - 5} more)`}
            </Alert>
          )}

          {/* 过期设置 */}
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

          <FormControlLabel
            control={<Checkbox checked={formData.enabled} onChange={(e) => handleChange('enabled', e.target.checked)} />}
            label={t('subscriptions.share.form.enabled')}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{t('common.cancel')}</Button>
        <Button onClick={handleSubmit} variant="contained" disabled={conflictingNames.length > 0 || !formData.baseName}>
          {t('subscriptions.share.batch.create')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
