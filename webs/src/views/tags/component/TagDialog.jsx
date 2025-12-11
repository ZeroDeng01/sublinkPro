import { useState, useEffect } from 'react';
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

export default function TagDialog({ open, onClose, onSave, editingTag }) {
  const [name, setName] = useState('');
  const [color, setColor] = useState('#1976d2');
  const [description, setDescription] = useState('');

  useEffect(() => {
    if (editingTag) {
      setName(editingTag.name || '');
      setColor(editingTag.color || '#1976d2');
      setDescription(editingTag.description || '');
    } else {
      setName('');
      setColor('#1976d2');
      setDescription('');
    }
  }, [editingTag, open]);

  const handleSave = () => {
    if (!name.trim()) return;
    onSave({ name: name.trim(), color, description });
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{editingTag ? '编辑标签' : '添加标签'}</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 1, display: 'flex', flexDirection: 'column', gap: 2 }}>
          <TextField label="标签名称" value={name} onChange={(e) => setName(e.target.value)} fullWidth required autoFocus />
          <Box>
            <Typography variant="body2" sx={{ mb: 1 }}>
              标签颜色
            </Typography>
            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
              {colorPresets.map((c) => (
                <Box
                  key={c}
                  onClick={() => setColor(c)}
                  sx={{
                    width: 32,
                    height: 32,
                    borderRadius: '50%',
                    backgroundColor: c,
                    cursor: 'pointer',
                    border: color === c ? '3px solid #000' : '2px solid transparent',
                    transition: 'all 0.2s',
                    '&:hover': {
                      transform: 'scale(1.1)'
                    }
                  }}
                />
              ))}
            </Box>
            <TextField
              label="自定义颜色 (HEX)"
              value={color}
              onChange={(e) => setColor(e.target.value)}
              size="small"
              sx={{ mt: 1.5, width: 150 }}
              InputProps={{
                startAdornment: (
                  <Box
                    sx={{
                      width: 20,
                      height: 20,
                      borderRadius: '4px',
                      backgroundColor: color,
                      mr: 1
                    }}
                  />
                )
              }}
            />
          </Box>
          <TextField
            label="描述 (可选)"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            fullWidth
            multiline
            rows={2}
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={handleSave} disabled={!name.trim()}>
          保存
        </Button>
      </DialogActions>
    </Dialog>
  );
}

TagDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  editingTag: PropTypes.object
};
