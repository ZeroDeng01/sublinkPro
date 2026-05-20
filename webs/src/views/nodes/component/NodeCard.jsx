import PropTypes from 'prop-types';

// material-ui
import { alpha, useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import NodeProtocolChip from './NodeProtocolChip';

// utils
import {
  getDelayDisplay,
  getFraudScoreDisplay,
  getIpTypeDisplay,
  getNodeUnlockSummaryDisplay,
  getQualityStatusDisplay,
  getResidentialDisplay,
  getSpeedDisplay,
  formatCountry
} from '../utils';
import { getNodePanelSx, getNodeTagChipSx, getNodeThemeTokens } from '../nodeTheme';

/**
 * 移动端节点卡片组件（精简版）
 * 只显示核心信息，点击卡片打开详情面板
 */
export default function NodeCard({ node, isSelected, tagColorMap, protocolMeta, onSelect, onViewDetails }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const unlockDisplay = getNodeUnlockSummaryDisplay(node, { limit: 2 });
  const effectiveName = node.EffectiveName || node.Name || node.LinkName;
  const secondaryName = node.NameMode === 'remark' ? node.LinkName : node.Name;
  const showSecondaryName = secondaryName && secondaryName !== effectiveName;

  return (
    <MainCard
      content={false}
      border
      shadow={theme.shadows[1]}
      sx={{
        ...getNodePanelSx(theme, tokens, tokens.palette.primary.main, {
          interactive: true,
          selected: isSelected
        }),
        backgroundImage: 'none',
        bgcolor: isSelected ? alpha(theme.palette.primary.main, isDark ? 0.14 : 0.06) : tokens.cardSurface,
        cursor: 'pointer'
      }}
      onClick={(e) => {
        // 点击复选框时不触发详情
        if (e.target.closest('input[type="checkbox"]')) return;
        onViewDetails(node);
      }}
    >
      <Box p={2}>
        <Box sx={{ position: 'relative', mb: 1.5, pr: 10 }}>
          <Box sx={{ position: 'absolute', top: 0, right: 0 }}>
            <Chip
              label={node.LinkCountry ? formatCountry(node.LinkCountry) : '🏳️ 未知'}
              color={node.LinkCountry ? 'secondary' : 'default'}
              variant="outlined"
              size="small"
            />
          </Box>
          <Stack direction="row" alignItems="flex-start" sx={{ minWidth: 0 }}>
            <Checkbox
              checked={isSelected}
              onChange={(e) => {
                e.stopPropagation();
                onSelect(node);
              }}
              sx={{ p: 0.5, flexShrink: 0 }}
            />
            <Tooltip title={effectiveName} placement="top">
              <Box sx={{ flex: 1, minWidth: 0, pr: 1 }}>
                <Typography
                  variant="subtitle1"
                  fontWeight="bold"
                  sx={{
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap'
                  }}
                >
                  {effectiveName}
                </Typography>
                {showSecondaryName && (
                  <Typography
                    variant="caption"
                    sx={{ color: tokens.secondaryText, display: 'block', overflow: 'hidden', textOverflow: 'ellipsis' }}
                  >
                    {node.NameMode === 'remark' ? `原始：${secondaryName}` : `备注：${secondaryName}`}
                  </Typography>
                )}
              </Box>
            </Tooltip>
          </Stack>
        </Box>

        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mb: 1 }}>
          <NodeProtocolChip link={node.Link} protocolMeta={protocolMeta} maxWidth={104} />
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
        </Stack>

        <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap sx={{ mb: 1 }}>
          {(() => {
            const d = getDelayDisplay(node.DelayTime, node.DelayStatus);
            return (
              <Chip
                icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>⏱️</span>}
                label={d.label}
                color={d.color}
                variant={d.variant}
                size="small"
              />
            );
          })()}
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
        </Stack>

        <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap sx={{ mb: 1 }}>
          {(() => {
            const ipTypeDisplay = getIpTypeDisplay(node.IsBroadcast, node.QualityStatus, node.QualityFamily);
            const residentialDisplay = getResidentialDisplay(node.IsResidential, node.QualityStatus, node.QualityFamily);
            const fraudScoreDisplay = getFraudScoreDisplay(node.FraudScore, node.QualityStatus, node.QualityFamily);
            const qualityStatusDisplay = getQualityStatusDisplay(node.QualityStatus, node.QualityFamily);
            const isUntested =
              ipTypeDisplay.label === '未检测' && residentialDisplay.label === '未检测' && fraudScoreDisplay.label === '未检测';
            const shouldMergeQualityTags =
              node.QualityStatus !== 'success' &&
              ipTypeDisplay.label === residentialDisplay.label &&
              residentialDisplay.label === fraudScoreDisplay.label;

            if (isUntested) {
              return <Chip label="未检测" color="default" variant="outlined" size="small" />;
            }

            if (shouldMergeQualityTags) {
              const mergedChip = (
                <Chip
                  label={qualityStatusDisplay.label}
                  color={qualityStatusDisplay.color}
                  variant={qualityStatusDisplay.variant}
                  size="small"
                />
              );
              return qualityStatusDisplay.tooltip ? <Tooltip title={qualityStatusDisplay.tooltip}>{mergedChip}</Tooltip> : mergedChip;
            }

            const ipTypeChip = (
              <Chip label={ipTypeDisplay.label} color={ipTypeDisplay.color} variant={ipTypeDisplay.variant} size="small" />
            );
            const residentialChip = (
              <Chip label={residentialDisplay.label} color={residentialDisplay.color} variant={residentialDisplay.variant} size="small" />
            );
            const fraudChip = (
              <Chip
                label={node.QualityStatus === 'success' ? `评分 ${fraudScoreDisplay.label}` : fraudScoreDisplay.label}
                color={fraudScoreDisplay.color}
                variant={fraudScoreDisplay.variant}
                size="small"
                sx={fraudScoreDisplay.sx}
              />
            );

            return (
              <>
                {ipTypeDisplay.tooltip ? <Tooltip title={ipTypeDisplay.tooltip}>{ipTypeChip}</Tooltip> : ipTypeChip}
                {residentialDisplay.tooltip ? <Tooltip title={residentialDisplay.tooltip}>{residentialChip}</Tooltip> : residentialChip}
                {fraudScoreDisplay.tooltip ? <Tooltip title={fraudScoreDisplay.tooltip}>{fraudChip}</Tooltip> : fraudChip}
              </>
            );
          })()}
        </Stack>

        {unlockDisplay?.compactItems?.length > 0 && (
          <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap sx={{ mb: 1 }}>
            {unlockDisplay.compactItems.map((item) => {
              const chip = (
                <Chip
                  key={`unlock-${item.provider}`}
                  icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>🔓</span>}
                  label={item.compactLabel}
                  color={item.color}
                  variant={item.variant}
                  size="small"
                />
              );

              return item.tooltip ? (
                <Tooltip key={`unlock-tip-${item.provider}`} title={item.tooltip}>
                  {chip}
                </Tooltip>
              ) : (
                chip
              );
            })}
            {unlockDisplay.extraCount > 0 && (
              <Chip label={`+${unlockDisplay.extraCount}`} color="default" variant="outlined" size="small" />
            )}
          </Stack>
        )}

        {/* 标签区 */}
        {node.Tags && (
          <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
            {node.Tags.split(',')
              .filter((t) => t.trim())
              .map((tag, idx) => {
                const tagName = tag.trim();
                const tagColor = tagColorMap?.[tagName] || tokens.palette.primary.main;
                return (
                  <Chip
                    key={`tag-${idx}`}
                    label={tagName}
                    size="small"
                    sx={{ fontSize: '10px', height: 20, ...getNodeTagChipSx(theme, tokens, tagColor) }}
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
    IsBroadcast: PropTypes.bool,
    IsResidential: PropTypes.bool,
    FraudScore: PropTypes.number,
    LandingIP: PropTypes.string,
    CreatedAt: PropTypes.string,
    UpdatedAt: PropTypes.string,
    LatencyCheckAt: PropTypes.string,
    SpeedCheckAt: PropTypes.string,
    Tags: PropTypes.string
  }).isRequired,
  isSelected: PropTypes.bool.isRequired,
  tagColorMap: PropTypes.object,
  protocolMeta: PropTypes.array,
  onSelect: PropTypes.func.isRequired,
  onViewDetails: PropTypes.func.isRequired
};
