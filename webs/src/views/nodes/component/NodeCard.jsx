import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import SpeedIcon from '@mui/icons-material/Speed';

// project imports
import MainCard from 'ui-component/cards/MainCard';

// utils
import { formatDateTime, formatCountry, getDelayDisplay, getSpeedDisplay } from '../utils';

/**
 * 移动端节点卡片组件
 */
export default function NodeCard({ node, isSelected, tagColorMap, onSelect, onSpeedTest, onCopy, onEdit, onDelete }) {
  const theme = useTheme();

  return (
    <MainCard content={false} border shadow={theme.shadows[1]}>
      <Box p={2}>
        {/* Header: Checkbox, Name, Delay */}
        <Stack direction="row" justifyContent="space-between" alignItems="flex-start" mb={1.5}>
          <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
            <Checkbox checked={isSelected} onChange={() => onSelect(node)} sx={{ p: 0.5, flexShrink: 0 }} />
            <Tooltip title={node.Name} placement="top">
              <Typography
                variant="subtitle1"
                fontWeight="bold"
                sx={{
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  maxWidth: '200px'
                }}
              >
                {node.Name}
              </Typography>
            </Tooltip>
          </Stack>
          <Box sx={{ flexShrink: 0, ml: 1 }}>
            {(() => {
              const d = getDelayDisplay(node.DelayTime, node.DelayStatus);
              return <Chip label={d.label} color={d.color} variant={d.variant} size="small" />;
            })()}
          </Box>
        </Stack>

        {/* Info Section: Chips for Group, Source, Speed */}
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mb: 1.5 }}>
          <Tooltip title={`分组: ${node.Group || '未分组'}`}>
            <Chip
              icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>📁</span>}
              label={node.Group || '未分组'}
              color="warning"
              variant="outlined"
              size="small"
              sx={{ maxWidth: '120px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
            />
          </Tooltip>
          <Chip
            icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>📡</span>}
            label={node.Source === 'manual' ? '手动添加' : node.Source || '未知'}
            color={node.Source === 'manual' ? 'success' : 'info'}
            variant="outlined"
            size="small"
            sx={{ maxWidth: '100px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
          />
          {(() => {
            const s = getSpeedDisplay(node.Speed, node.SpeedStatus);
            return (
              <Chip
                icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>⚡</span>}
                label={s.label}
                color={s.color}
                variant={s.variant}
                size="small"
              />
            );
          })()}
          {node.DialerProxyName && (
            <Tooltip title={`前置代理: ${node.DialerProxyName}`}>
              <Chip
                icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>🔗</span>}
                label={node.DialerProxyName}
                variant="outlined"
                size="small"
                sx={{ maxWidth: '100px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
              />
            </Tooltip>
          )}
          {node.LinkCountry && (
            <Tooltip title={`国家: ${node.LinkCountry}`}>
              <Chip label={formatCountry(node.LinkCountry)} color="secondary" variant="outlined" size="small" />
            </Tooltip>
          )}
          {node.Tags &&
            node.Tags.split(',')
              .filter((t) => t.trim())
              .map((tag, idx) => {
                const tagName = tag.trim();
                const tagColor = tagColorMap?.[tagName] || '#1976d2';
                return (
                  <Chip
                    key={`tag-${idx}`}
                    label={tagName}
                    size="small"
                    sx={{
                      fontSize: '10px',
                      height: 20,
                      backgroundColor: tagColor,
                      color: '#fff'
                    }}
                  />
                );
              })}
        </Stack>

        {/* Time Info Section */}
        <Box sx={{ bgcolor: 'action.hover', borderRadius: 1, p: 1, mb: 1.5 }}>
          <Stack spacing={0.5}>
            <Box>
              <Typography variant="caption" color="textSecondary" display="block">
                创建时间
              </Typography>
              <Typography variant="caption" fontWeight="medium">
                {node.CreatedAt ? formatDateTime(node.CreatedAt) : '-'}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="textSecondary" display="block">
                更新时间
              </Typography>
              <Typography variant="caption" fontWeight="medium">
                {node.UpdatedAt ? formatDateTime(node.UpdatedAt) : '-'}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="textSecondary" display="block">
                延迟测试
              </Typography>
              <Typography variant="caption" fontWeight="medium" color="primary">
                {node.LatencyCheckAt ? formatDateTime(node.LatencyCheckAt) : '-'}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="textSecondary" display="block">
                速度测试
              </Typography>
              <Typography variant="caption" fontWeight="medium" color="primary">
                {node.SpeedCheckAt ? formatDateTime(node.SpeedCheckAt) : '-'}
              </Typography>
            </Box>
          </Stack>
        </Box>

        {/* Action Buttons */}
        <Stack direction="row" spacing={0.5} justifyContent="flex-end">
          <Tooltip title="测速">
            <IconButton size="small" onClick={() => onSpeedTest(node)}>
              <SpeedIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title="复制链接">
            <IconButton size="small" onClick={() => onCopy(node.Link)}>
              <ContentCopyIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title="编辑">
            <IconButton size="small" onClick={() => onEdit(node)}>
              <EditIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title="删除">
            <IconButton size="small" color="error" onClick={() => onDelete(node)}>
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Stack>
      </Box>
    </MainCard>
  );
}

NodeCard.propTypes = {
  node: PropTypes.shape({
    ID: PropTypes.number,
    Name: PropTypes.string,
    Link: PropTypes.string,
    Group: PropTypes.string,
    Source: PropTypes.string,
    DelayTime: PropTypes.number,
    Speed: PropTypes.number,
    DialerProxyName: PropTypes.string,
    LinkCountry: PropTypes.string,
    CreatedAt: PropTypes.string,
    UpdatedAt: PropTypes.string,
    LatencyCheckAt: PropTypes.string,
    SpeedCheckAt: PropTypes.string,
    Tags: PropTypes.string
  }).isRequired,
  isSelected: PropTypes.bool.isRequired,
  tagColorMap: PropTypes.object,
  onSelect: PropTypes.func.isRequired,
  onSpeedTest: PropTypes.func.isRequired,
  onCopy: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired
};
