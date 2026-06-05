import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

// material-ui
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Autocomplete from '@mui/material/Autocomplete';
import Chip from '@mui/material/Chip';
import Alert from '@mui/material/Alert';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import InputAdornment from '@mui/material/InputAdornment';

// icons
import ColorLensIcon from '@mui/icons-material/ColorLens';

// Color presets
const colorPresets = [
  '#1976d2', // Blue
  '#388e3c', // Green
  '#d32f2f', // Red
  '#f57c00', // Orange
  '#7b1fa2', // Purple
  '#0097a7', // Cyan
  '#c2185b', // Pink
  '#455a64', // Blue Grey
  '#5d4037', // Brown
  '#616161' // Grey
];

const presetGroups = [
  {
    valueKey: 'speedRating',
    labelKey: 'tags.dialog.tag.presetGroups.speedRating',
    descKey: 'tags.dialog.tag.presetDescriptions.speedRating'
  },
  {
    valueKey: 'latencyRating',
    labelKey: 'tags.dialog.tag.presetGroups.latencyRating',
    descKey: 'tags.dialog.tag.presetDescriptions.latencyRating'
  },
  {
    valueKey: 'regionCategory',
    labelKey: 'tags.dialog.tag.presetGroups.regionCategory',
    descKey: 'tags.dialog.tag.presetDescriptions.regionCategory'
  },
  {
    valueKey: 'usageCategory',
    labelKey: 'tags.dialog.tag.presetGroups.usageCategory',
    descKey: 'tags.dialog.tag.presetDescriptions.usageCategory'
  },
  { valueKey: 'stability', labelKey: 'tags.dialog.tag.presetGroups.stability', descKey: 'tags.dialog.tag.presetDescriptions.stability' }
];

export default function TagDialog({ open, onClose, onSave, editingTag, existingGroups = [] }) {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [color, setColor] = useState('#1976d2');
  const [description, setDescription] = useState('');
  const [groupName, setGroupName] = useState('');
  const colorPickerRef = useRef(null);

  const handleColorInput = (value) => {
    let newColor = value.trim();
    if (newColor && !newColor.startsWith('#') && /^[0-9A-Fa-f]{3,6}$/.test(newColor)) {
      newColor = '#' + newColor;
    }
    setColor(newColor);
  };

  const isValidColor = (c) => {
    return /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(c);
  };

  const localizedPresetGroups = presetGroups.map((group) => ({
    value: t(group.labelKey),
    description: t(group.descKey)
  }));
  const allGroupOptions = [...new Set([...localizedPresetGroups.map((g) => g.value), ...existingGroups])];

  useEffect(() => {
    if (editingTag) {
      setName(editingTag.name || '');
      setColor(editingTag.color || '#1976d2');
      setDescription(editingTag.description || '');
      setGroupName(editingTag.groupName || '');
    } else {
      setName('');
      setColor('#1976d2');
      setDescription('');
      setGroupName('');
    }
  }, [editingTag, open]);

  const handleSave = () => {
    if (!name.trim()) return;
    onSave({ name: name.trim(), color, description, groupName: groupName.trim() });
  };

  const getGroupDescription = (group) => {
    const preset = localizedPresetGroups.find((g) => g.value === group);
    return preset ? preset.description : null;
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{editingTag ? t('tags.dialog.tag.editTitle') : t('tags.dialog.tag.addTitle')}</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 1, display: 'flex', flexDirection: 'column', gap: 2 }}>
          <Alert severity="info" sx={{ '& .MuiAlert-message': { width: '100%' } }}>
            <Typography variant="body2" sx={{ fontWeight: 500, mb: 0.5 }}>
              {t('tags.dialog.tag.helpTitle')}
            </Typography>
            <Typography variant="caption" component="div">
              {t('tags.dialog.tag.helpTag')}
              <br />
              {t('tags.dialog.tag.helpGroup')}
              <br />
              {t('tags.dialog.tag.helpExample')}
            </Typography>
          </Alert>

          <TextField
            label={t('tags.dialog.tag.name')}
            value={name}
            onChange={(e) => setName(e.target.value)}
            fullWidth
            required
            autoFocus
            disabled={!!editingTag}
          />

          <Box>
            <Autocomplete
              freeSolo
              value={groupName}
              onChange={(_, newValue) => setGroupName(newValue || '')}
              onInputChange={(_, newValue) => setGroupName(newValue || '')}
              options={allGroupOptions}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('tags.dialog.tag.group')}
                  placeholder={t('tags.dialog.tag.groupPlaceholder')}
                  helperText={t('tags.dialog.tag.groupHelper')}
                />
              )}
              renderOption={(props, option) => {
                const { key, ...otherProps } = props;
                const desc = getGroupDescription(option);
                return (
                  <li key={key} {...otherProps}>
                    <Box>
                      <Typography variant="body2">{option}</Typography>
                      {desc && (
                        <Typography variant="caption" color="text.secondary">
                          {desc}
                        </Typography>
                      )}
                    </Box>
                  </li>
                );
              }}
            />
            {groupName && (
              <Box sx={{ mt: 1, display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                <Typography variant="caption" color="text.secondary" sx={{ mr: 1 }}>
                  {t('tags.dialog.tag.recommendedGroups')}
                </Typography>
                {localizedPresetGroups.slice(0, 3).map((g) => (
                  <Chip
                    key={g.value}
                    label={g.value}
                    size="small"
                    variant={groupName === g.value ? 'filled' : 'outlined'}
                    onClick={() => setGroupName(g.value)}
                    sx={{ cursor: 'pointer' }}
                  />
                ))}
              </Box>
            )}
          </Box>

          <Divider />

          <Box>
            <Typography variant="body2" sx={{ mb: 1.5, fontWeight: 500 }}>
              {t('tags.dialog.tag.color')}
            </Typography>

            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', alignItems: 'center' }}>
              {colorPresets.map((c) => (
                <Box
                  key={c}
                  onClick={() => setColor(c)}
                  sx={{
                    width: { xs: 36, sm: 32 },
                    height: { xs: 36, sm: 32 },
                    borderRadius: '50%',
                    backgroundColor: c,
                    cursor: 'pointer',
                    border: color.toLowerCase() === c.toLowerCase() ? '3px solid #000' : '2px solid rgba(0,0,0,0.1)',
                    transition: 'all 0.2s',
                    boxShadow: color.toLowerCase() === c.toLowerCase() ? '0 0 0 2px rgba(0,0,0,0.1)' : 'none',
                    '&:hover': {
                      transform: 'scale(1.15)',
                      boxShadow: '0 2px 8px rgba(0,0,0,0.2)'
                    },
                    '&:active': {
                      transform: 'scale(0.95)'
                    }
                  }}
                />
              ))}

              <Tooltip title={t('tags.dialog.tag.openColorPicker')}>
                <IconButton
                  onClick={() => colorPickerRef.current?.click()}
                  sx={{
                    width: { xs: 36, sm: 32 },
                    height: { xs: 36, sm: 32 },
                    border: '2px dashed',
                    borderColor: 'divider',
                    backgroundColor: 'background.paper',
                    '&:hover': {
                      backgroundColor: 'action.hover',
                      borderColor: 'primary.main'
                    }
                  }}
                >
                  <ColorLensIcon sx={{ fontSize: 18 }} />
                </IconButton>
              </Tooltip>

              <input
                ref={colorPickerRef}
                type="color"
                value={isValidColor(color) ? color : '#1976d2'}
                onChange={(e) => setColor(e.target.value)}
                style={{
                  position: 'absolute',
                  opacity: 0,
                  width: 0,
                  height: 0,
                  pointerEvents: 'none'
                }}
              />
            </Box>

            <Box sx={{ mt: 2, display: 'flex', gap: 1, alignItems: 'center', flexWrap: 'wrap' }}>
              <TextField
                label={t('tags.dialog.tag.customColor')}
                value={color}
                onChange={(e) => handleColorInput(e.target.value)}
                size="small"
                error={color && !isValidColor(color)}
                helperText={color && !isValidColor(color) ? t('tags.dialog.tag.invalidColor') : ''}
                sx={{ width: { xs: '100%', sm: 180 } }}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <Box
                        onClick={() => colorPickerRef.current?.click()}
                        sx={{
                          width: 24,
                          height: 24,
                          borderRadius: '4px',
                          backgroundColor: isValidColor(color) ? color : '#ccc',
                          cursor: 'pointer',
                          border: '1px solid rgba(0,0,0,0.1)',
                          transition: 'all 0.2s',
                          '&:hover': {
                            transform: 'scale(1.1)',
                            boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                          }
                        }}
                      />
                    </InputAdornment>
                  )
                }}
              />
            </Box>
          </Box>
          <TextField
            label={t('tags.dialog.tag.description')}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            fullWidth
            multiline
            rows={2}
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={handleSave} disabled={!name.trim()}>
          {t('common.save')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

TagDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  editingTag: PropTypes.object,
  existingGroups: PropTypes.array
};
