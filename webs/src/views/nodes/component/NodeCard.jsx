import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// project imports
import MainCard from 'ui-component/cards/MainCard';

// utils
import { getDelayDisplay, getSpeedDisplay, formatCountry } from '../utils';

/**
 * 移动端节点卡片组件（精简版）
 * 只显示核心信息，点击卡片打开详情面板
 */
export default function NodeCard({ node, isSelected, tagColorMap, onSelect, onViewDetails }) {
  const theme = useTheme();

  return (
    <MainCard
      content={false}
      border
      shadow={theme.shadows[1]}
      sx={{
        cursor: 'pointer',
        transition: 'all 0.2s ease',
        '&:hover': {
          boxShadow: theme.shadows[4],
          transform: 'translateY(-2px)'
        }
      }}
      onClick={(e) => {
        // 点击复选框时不触发详情
        if (e.target.closest('input[type="checkbox"]')) return;
        onViewDetails(node);
      }}
    >
      <Box p={2}>
        {/* 头部: 勾选框 + 名称 + 延迟 */}
        <Stack direction="row" justifyContent="space-between" alignItems="flex-start" mb={1.5}>
          <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
            <Checkbox
              checked={isSelected}
              onChange={(e) => {
                e.stopPropagation();
                onSelect(node);
              }}
              sx={{ p: 0.5, flexShrink: 0 }}
            />
            <Tooltip title={node.Name} placement="top">
              <Typography
                variant="subtitle1"
                fontWeight="bold"
                sx={{
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  maxWidth: '180px'
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

        {/* 信息区: 分组 + 来源 + 速度 + 国家 */}
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mb: 1 }}>
          {node.Group && (
            <Tooltip title={`分组: ${node.Group}`}>
              <Chip
                icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>📁</span>}
                label={node.Group}
                color="warning"
                variant="outlined"
                size="small"
                sx={{ maxWidth: '100px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
              />
            </Tooltip>
          )}
          {node.Source && node.Source !== 'manual' && (
            <Tooltip title={`来源: ${node.Source}`}>
              <Chip
                icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>📥</span>}
                label={node.Source}
                color="info"
                variant="outlined"
                size="small"
                sx={{ maxWidth: '100px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
              />
            </Tooltip>
          )}
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
          {node.LinkCountry && <Chip label={formatCountry(node.LinkCountry)} color="secondary" variant="outlined" size="small" />}
        </Stack>

        {/* 标签区 */}
        {node.Tags && (
          <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
            {node.Tags.split(',')
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
        )}

        {/* 点击提示 */}
        <Typography
          variant="caption"
          color="text.secondary"
          sx={{
            display: 'block',
            mt: 1.5,
            textAlign: 'center',
            opacity: 0.6
          }}
        >
          点击查看详情
        </Typography>
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
    DelayStatus: PropTypes.number,
    Speed: PropTypes.number,
    SpeedStatus: PropTypes.number,
    DialerProxyName: PropTypes.string,
    LinkCountry: PropTypes.string,
    LandingIP: PropTypes.string,
    CreatedAt: PropTypes.string,
    UpdatedAt: PropTypes.string,
    LatencyCheckAt: PropTypes.string,
    SpeedCheckAt: PropTypes.string,
    Tags: PropTypes.string
  }).isRequired,
  isSelected: PropTypes.bool.isRequired,
  tagColorMap: PropTypes.object,
  onSelect: PropTypes.func.isRequired,
  onViewDetails: PropTypes.func.isRequired
};
