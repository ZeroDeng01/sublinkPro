import { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Collapse from '@mui/material/Collapse';
import IconButton from '@mui/material/IconButton';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

// icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import FingerprintIcon from '@mui/icons-material/Fingerprint';

// 示例节点名用于预览
const PREVIEW_NODE_NAME = '香港节点-01';

/**
 * 节点名称唯一化配置组件
 * 通过添加机场标识前缀防止多机场间节点名称重复
 */
export default function NodeNameUniquifyConfig({ enabled, prefix, airportId, onChange }) {
  const [expanded, setExpanded] = useState(false);

  // 根据当前配置计算预览结果
  const previewResult = useMemo(() => {
    if (!enabled) {
      return PREVIEW_NODE_NAME;
    }
    // 使用用户自定义前缀，或默认的 [A{id}] 格式
    const displayPrefix = prefix || `[${airportId || 0}]`;
    return displayPrefix + PREVIEW_NODE_NAME;
  }, [enabled, prefix, airportId]);

  // 自动展开面板（当开启功能时）
  useEffect(() => {
    if (enabled) {
      setExpanded(true);
    }
  }, [enabled]);

  const handleEnabledChange = (event) => {
    onChange({ enabled: event.target.checked, prefix });
  };

  const handlePrefixChange = (event) => {
    onChange({ enabled, prefix: event.target.value });
  };

  return (
    <Box
      sx={{
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 2,
        overflow: 'hidden'
      }}
    >
      {/* 标题栏 */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 2,
          py: 1,
          bgcolor: 'background.default',
          cursor: 'pointer',
          '&:hover': { bgcolor: 'action.hover' }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <FingerprintIcon color="action" fontSize="small" />
          <Typography variant="body2" fontWeight={500}>
            节点名称唯一化
          </Typography>
          {enabled && (
            <Typography
              variant="caption"
              sx={{
                px: 1,
                py: 0.25,
                borderRadius: 1,
                bgcolor: 'primary.main',
                color: 'primary.contrastText'
              }}
            >
              已开启
            </Typography>
          )}
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Switch size="small" checked={enabled} onClick={(e) => e.stopPropagation()} onChange={handleEnabledChange} />
          <IconButton size="small">{expanded ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />}</IconButton>
        </Box>
      </Box>

      {/* 展开内容 */}
      <Collapse in={expanded}>
        <Box sx={{ p: 2, borderTop: '1px solid', borderColor: 'divider' }}>
          <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 2 }}>
            为节点名称添加机场标识前缀，防止多个机场之间节点名称重复。同一机场的节点每次拉取生成的前缀保持一致。
          </Typography>

          {/* 自定义前缀输入 */}
          <TextField
            fullWidth
            size="small"
            label="自定义前缀（可选）"
            placeholder={`留空使用默认前缀 [${airportId || 0}]`}
            value={prefix}
            onChange={handlePrefixChange}
            disabled={!enabled}
            helperText="可自定义前缀使节点名称更具可读性"
            sx={{ mb: 2 }}
          />

          {/* 预览效果 */}
          <Box
            sx={{
              p: 1.5,
              borderRadius: 1,
              bgcolor: enabled ? 'action.selected' : 'action.disabledBackground'
            }}
          >
            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 0.5 }}>
              预览效果
            </Typography>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
              <Typography variant="body2" sx={{ fontFamily: 'monospace', color: 'text.secondary', textDecoration: 'line-through' }}>
                {PREVIEW_NODE_NAME}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                →
              </Typography>
              <Typography
                variant="body2"
                sx={{
                  fontFamily: 'monospace',
                  fontWeight: 500,
                  color: enabled ? 'primary.main' : 'text.disabled'
                }}
              >
                {previewResult}
              </Typography>
            </Box>
          </Box>
        </Box>
      </Collapse>
    </Box>
  );
}

NodeNameUniquifyConfig.propTypes = {
  enabled: PropTypes.bool.isRequired,
  prefix: PropTypes.string,
  airportId: PropTypes.number,
  onChange: PropTypes.func.isRequired
};

NodeNameUniquifyConfig.defaultProps = {
  prefix: '',
  airportId: 0
};
