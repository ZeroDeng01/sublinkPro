import { useState, useEffect, useRef } from 'react';
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

// 预设标签组
const presetGroups = [
  { value: '速度评级', description: '根据测速结果分类：优秀、良好、一般、差' },
  { value: '延迟评级', description: '根据延迟分类：低延迟、中等延迟、高延迟' },
  { value: '地区分类', description: '按地理区域分类：亚洲、欧洲、美洲等' },
  { value: '用途分类', description: '按使用场景分类：流媒体、游戏、下载等' },
  { value: '稳定性', description: '按节点稳定性分类：稳定、不稳定' }
];

export default function TagDialog({ open, onClose, onSave, editingTag, existingGroups = [] }) {
  const [name, setName] = useState('');
  const [color, setColor] = useState('#1976d2');
  const [description, setDescription] = useState('');
  const [groupName, setGroupName] = useState('');
  const colorPickerRef = useRef(null);

  // 处理颜色输入，支持带或不带#的hex值
  const handleColorInput = (value) => {
    let newColor = value.trim();
    // 如果输入不以#开头且看起来像hex值，自动添加#
    if (newColor && !newColor.startsWith('#') && /^[0-9A-Fa-f]{3,6}$/.test(newColor)) {
      newColor = '#' + newColor;
    }
    setColor(newColor);
  };

  // 验证颜色格式
  const isValidColor = (c) => {
    return /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(c);
  };

  // 合并预设组和已有组
  const allGroupOptions = [...new Set([...presetGroups.map((g) => g.value), ...existingGroups])];

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
    const preset = presetGroups.find((g) => g.value === group);
    return preset ? preset.description : null;
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{editingTag ? '编辑标签' : '添加标签'}</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 1, display: 'flex', flexDirection: 'column', gap: 2 }}>
          {/* 帮助说明 */}
          <Alert severity="info" sx={{ '& .MuiAlert-message': { width: '100%' } }}>
            <Typography variant="body2" sx={{ fontWeight: 500, mb: 0.5 }}>
              💡 标签使用说明
            </Typography>
            <Typography variant="caption" component="div">
              • <strong>标签</strong>：用于对节点进行分类标记，可用于筛选和自动规则
              <br />• <strong>标签组</strong>：同一组内的标签互斥，添加新标签时会自动移除同组的旧标签
              <br />• 例如：创建"优秀"和"差"两个标签并设为同组，测速时节点只会保留最新的评级
            </Typography>
          </Alert>

          <TextField
            label="标签名称"
            value={name}
            onChange={(e) => setName(e.target.value)}
            fullWidth
            required
            autoFocus
            disabled={!!editingTag}
          />

          {/* 标签组选择 */}
          <Box>
            <Autocomplete
              freeSolo
              value={groupName}
              onChange={(e, newValue) => setGroupName(newValue || '')}
              onInputChange={(e, newValue) => setGroupName(newValue || '')}
              options={allGroupOptions}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="标签组 (可选)"
                  placeholder="选择或输入标签组名称"
                  helperText="同一组内的标签互斥，为空则不参与互斥"
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
                  推荐组：
                </Typography>
                {presetGroups.slice(0, 3).map((g) => (
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

          {/* 颜色选择 */}
          <Box>
            <Typography variant="body2" sx={{ mb: 1.5, fontWeight: 500 }}>
              标签颜色
            </Typography>

            {/* 预设颜色 + 颜色选择器按钮 */}
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

              {/* 颜色选择器按钮 */}
              <Tooltip title="打开颜色选择器">
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

              {/* 隐藏的原生颜色选择器 */}
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

            {/* 自定义颜色输入 */}
            <Box sx={{ mt: 2, display: 'flex', gap: 1, alignItems: 'center', flexWrap: 'wrap' }}>
              <TextField
                label="自定义颜色 (HEX)"
                value={color}
                onChange={(e) => handleColorInput(e.target.value)}
                size="small"
                error={color && !isValidColor(color)}
                helperText={color && !isValidColor(color) ? '请输入有效的HEX颜色值 (如 #FF5733)' : ''}
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
  editingTag: PropTypes.object,
  existingGroups: PropTypes.array
};
