import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Collapse from '@mui/material/Collapse';
import CircularProgress from '@mui/material/CircularProgress';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme, keyframes } from '@mui/material/styles';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from '../../../utils/colorUtils';

import CloseIcon from '@mui/icons-material/Close';
import PersonIcon from '@mui/icons-material/Person';
import PublicIcon from '@mui/icons-material/Public';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import RouteIcon from '@mui/icons-material/Route';
import LayersIcon from '@mui/icons-material/Layers';
import AdjustIcon from '@mui/icons-material/Adjust';
import FlagIcon from '@mui/icons-material/Flag';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';

import ChainCanvasView from './ChainCanvasView';
import { getChainProxyIconButtonSx, getChainProxyThemeTokens } from './chainProxyTheme';
import { COUNTRY_FALLBACK_EMOJI, isoToFlag } from '../../../utils/countryDisplay';

const pulseAnimation = keyframes`
  0%, 100% {
    opacity: 0.4;
    transform: translateX(0);
  }
  50% {
    opacity: 1;
    transform: translateX(4px);
  }
`;

const getCountryFlag = (code) => {
  return isoToFlag(code, COUNTRY_FALLBACK_EMOJI);
};

const getTypeColor = (type, palette) => {
  const colors = {
    template_group: palette.primary.main,
    custom_group: palette.secondary.main,
    dynamic_node: palette.warning.main,
    specified_node: palette.success.main
  };
  return colors[type] || palette.grey[500];
};

const getTypeVisualTokens = (type, tokens) => {
  const palette = tokens.palette;
  const visualMap = {
    template_group: {
      color: palette.primary.main,
      surface: tokens.primarySurface,
      border: tokens.primarySoftBorder,
      strongBorder: tokens.primaryStrongBorder
    },
    custom_group: {
      color: palette.secondary.main,
      surface: tokens.secondarySurface,
      border: tokens.secondarySoftBorder,
      strongBorder: withAlpha(palette.secondary.main, tokens.isDark ? 0.46 : 0.28)
    },
    dynamic_node: {
      color: palette.warning.main,
      surface: tokens.warningSurface,
      border: tokens.warningSoftBorder,
      strongBorder: withAlpha(palette.warning.main, tokens.isDark ? 0.44 : 0.28)
    },
    specified_node: {
      color: palette.success.main,
      surface: tokens.successSurface,
      border: tokens.successSoftBorder,
      strongBorder: withAlpha(palette.success.main, tokens.isDark ? 0.44 : 0.28)
    }
  };

  return (
    visualMap[type] || {
      color: getTypeColor(type, palette),
      surface: tokens.elevatedSurface,
      border: tokens.softBorder,
      strongBorder: tokens.softBorder
    }
  );
};

const getConnectorChevronColor = (color, isDark) => withAlpha(color, isDark ? 0.58 : 0.42);

const getTypeLabel = (type, t) => {
  const key = `subscriptions.chain.proxyTypeShort.${type}`;
  return t(key, type);
};

function ChainNodeCard({ node, index, isLast, isMobile, theme }) {
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState(false);
  const hasNodes = node.nodes && node.nodes.length > 0;
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const typeVisuals = getTypeVisualTokens(node.type, tokens);
  const typeColor = typeVisuals.color;
  const { elevatedSurface, nestedPanelSurface, secondaryText, tertiaryText, cardShadow, insetHighlight, primaryText } = tokens;

  return (
    <Box sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', alignItems: 'center' }}>
      <Card
        sx={{
          minWidth: isMobile ? '100%' : 140,
          maxWidth: isMobile ? '100%' : 180,
          backgroundColor: elevatedSurface,
          border: `1px solid ${typeVisuals.border}`,
          borderRadius: 2,
          transition: 'all 0.2s ease',
          boxShadow: cardShadow,
          '&:hover': {
            backgroundColor: typeVisuals.surface,
            borderColor: typeVisuals.strongBorder,
            boxShadow: isDark ? insetHighlight : theme.shadows[2]
          }
        }}
      >
        <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
          <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
            <Chip
              size="small"
              label={getTypeLabel(node.type, t)}
              sx={{
                bgcolor: typeVisuals.surface,
                color: typeColor,
                fontWeight: 600,
                fontSize: 10,
                height: 18,
                border: `1px solid ${typeVisuals.border}`,
                '& .MuiChip-label': { px: 0.75 }
              }}
            />
            <Typography variant="caption" sx={{ color: secondaryText }}>
              #{index + 1}
            </Typography>
          </Stack>

          <Typography
            variant="body2"
            fontWeight={600}
            sx={{
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              mb: hasNodes ? 0.5 : 0
            }}
            title={node.name}
          >
            {node.name || t('subscriptions.chain.unconfigured')}
          </Typography>

          {hasNodes && (
            <Box
              onClick={() => setExpanded(!expanded)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                cursor: 'pointer',
                color: typeColor,
                '&:hover': { textDecoration: 'underline' }
              }}
            >
              <Typography variant="caption" fontWeight={600}>
                {t('subscriptions.chain.preview.nodeCount', { count: node.nodes.length })}
              </Typography>
              {expanded ? <ExpandLessIcon sx={{ fontSize: 16 }} /> : <ExpandMoreIcon sx={{ fontSize: 16 }} />}
            </Box>
          )}
          <Collapse in={expanded}>
            <Box
              sx={{
                mt: 0.5,
                maxHeight: 100,
                overflow: 'auto',
                bgcolor: nestedPanelSurface,
                border: '1px solid',
                borderColor: typeVisuals.border,
                borderRadius: 1,
                p: 0.5,
                boxShadow: insetHighlight
              }}
            >
              {node.nodes?.map((n, i) => (
                <Typography key={i} variant="caption" display="block" sx={{ py: 0.25, color: n.name ? primaryText : tertiaryText }}>
                  {getCountryFlag(n.linkCountry)} {n.name}
                </Typography>
              ))}
            </Box>
          </Collapse>
        </CardContent>
      </Card>

      {!isLast && (
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            px: isMobile ? 0 : 1,
            py: isMobile ? 0.5 : 0,
            transform: isMobile ? 'rotate(90deg)' : 'none'
          }}
        >
          {[0, 1].map((i) => (
            <Box
              key={i}
              sx={{
                width: 0,
                height: 0,
                borderTop: '5px solid transparent',
                borderBottom: '5px solid transparent',
                borderLeft: `6px solid ${getConnectorChevronColor(tokens.palette.primary.main, isDark)}`,
                ml: i > 0 ? -0.3 : 0,
                animation: `${pulseAnimation} 1.5s ease-in-out infinite`,
                animationDelay: `${i * 0.15}s`
              }}
            />
          ))}
        </Box>
      )}
    </Box>
  );
}

ChainNodeCard.propTypes = {
  node: PropTypes.object.isRequired,
  index: PropTypes.number.isRequired,
  isLast: PropTypes.bool.isRequired,
  isMobile: PropTypes.bool.isRequired,
  theme: PropTypes.object.isRequired
};

function RuleChainFlow({ rule, isMobile, theme }) {
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState(!rule.fullyCovered);
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const { elevatedSurface, softBorder, secondaryText, tertiaryText, cardShadow, coveredSurface, coveredBorder } = tokens;
  const expandButtonSx = getChainProxyIconButtonSx(tokens);

  const isFullyCovered = rule.enabled && rule.fullyCovered;

  return (
    <Card
      variant="outlined"
      sx={{
        mb: 2,
        borderRadius: 2,
        opacity: rule.enabled ? (isFullyCovered ? 0.6 : 1) : 0.4,
        transition: 'all 0.2s ease',
        boxShadow: cardShadow,
        borderColor: isFullyCovered ? coveredBorder : softBorder,
        bgcolor: isFullyCovered ? coveredSurface : elevatedSurface
      }}
    >
      <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between" mb={1.5}>
          <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap">
            <Typography variant="subtitle1" fontWeight={700} sx={{ textDecoration: isFullyCovered ? 'line-through' : 'none' }}>
              {rule.ruleName || t('subscriptions.chain.unnamedRule')}
            </Typography>
            {!rule.enabled && <Chip label={t('common.disabled')} size="small" color="default" />}
            {rule.enabled && isFullyCovered && (
              <Chip label={t('subscriptions.chain.preview.covered')} size="small" color="warning" variant="outlined" />
            )}
            {rule.enabled && !isFullyCovered && rule.effectiveNodes > 0 && (
              <Chip
                label={t('subscriptions.chain.preview.effectiveNodes', { count: rule.effectiveNodes })}
                size="small"
                color="success"
                variant="outlined"
              />
            )}
            {rule.enabled && rule.coveredNodes > 0 && !isFullyCovered && (
              <Chip
                label={t('subscriptions.chain.preview.coveredNodes', { count: rule.coveredNodes })}
                size="small"
                color="warning"
                variant="outlined"
              />
            )}
          </Stack>
          <IconButton size="small" onClick={() => setExpanded(!expanded)} sx={expandButtonSx}>
            {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
          </IconButton>
        </Stack>

        <Collapse in={expanded}>
          <Box
            sx={{
              display: 'flex',
              flexDirection: isMobile ? 'column' : 'row',
              alignItems: 'center',
              gap: 0,
              py: 1,
              px: isMobile ? 0 : 1,
              overflowX: 'auto'
            }}
          >
            <Box sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', alignItems: 'center' }}>
              <Box
                sx={{
                  px: 1.5,
                  py: 0.75,
                  borderRadius: 2,
                  bgcolor: tokens.infoSurface,
                  border: `1px dashed ${tokens.infoSoftBorder}`,
                  textAlign: 'center',
                  minWidth: 60
                }}
              >
                <PersonIcon sx={{ color: theme.palette.info.main, fontSize: 20 }} />
                <Typography variant="caption" display="block" fontWeight={600}>
                  {t('subscriptions.chain.preview.user')}
                </Typography>
              </Box>
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  px: isMobile ? 0 : 1,
                  py: isMobile ? 0.5 : 0,
                  transform: isMobile ? 'rotate(90deg)' : 'none'
                }}
              >
                {[0, 1].map((i) => (
                  <Box
                    key={i}
                    sx={{
                      width: 0,
                      height: 0,
                      borderTop: '5px solid transparent',
                      borderBottom: '5px solid transparent',
                      borderLeft: `6px solid ${getConnectorChevronColor(tokens.palette.info.main, isDark)}`,
                      ml: i > 0 ? -0.3 : 0
                    }}
                  />
                ))}
              </Box>
            </Box>

            {rule.links?.map((node, index) => (
              <ChainNodeCard
                key={index}
                node={node}
                index={index}
                isLast={index === rule.links.length - 1}
                isMobile={isMobile}
                theme={theme}
              />
            ))}

            {rule.links?.length > 0 && (
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  px: isMobile ? 0 : 1,
                  py: isMobile ? 0.5 : 0,
                  transform: isMobile ? 'rotate(90deg)' : 'none'
                }}
              >
                {[0, 1].map((i) => (
                  <Box
                    key={i}
                    sx={{
                      width: 0,
                      height: 0,
                      borderTop: '5px solid transparent',
                      borderBottom: '5px solid transparent',
                      borderLeft: `6px solid ${getConnectorChevronColor(tokens.palette.warning.main, isDark)}`,
                      ml: i > 0 ? -0.3 : 0
                    }}
                  />
                ))}
              </Box>
            )}

            <Box
              sx={{
                px: 1.5,
                py: 0.75,
                borderRadius: 2,
                bgcolor: tokens.warningSurface,
                border: `1px solid ${tokens.warningSoftBorder}`,
                textAlign: 'center',
                minWidth: isMobile ? '100%' : 100
              }}
            >
              {rule.targetType === 'all' ? (
                <LayersIcon sx={{ color: theme.palette.warning.main, fontSize: 20 }} />
              ) : rule.targetType === 'conditions' ? (
                <AdjustIcon sx={{ color: theme.palette.warning.main, fontSize: 20 }} />
              ) : (
                <FlagIcon sx={{ color: theme.palette.warning.main, fontSize: 20 }} />
              )}
              <Typography variant="caption" display="block" fontWeight={600} color="warning.main">
                {t('subscriptions.chain.preview.landingNode')}
              </Typography>
              <Typography variant="caption" display="block" sx={{ color: tertiaryText }}>
                {rule.targetInfo}
              </Typography>
              {rule.targetNodes?.length > 0 && (
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  {t('subscriptions.chain.preview.nodeCountParen', { count: rule.targetNodes.length })}
                </Typography>
              )}
            </Box>

            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                px: isMobile ? 0 : 1,
                py: isMobile ? 0.5 : 0,
                transform: isMobile ? 'rotate(90deg)' : 'none'
              }}
            >
              {[0, 1].map((i) => (
                <Box
                  key={i}
                  sx={{
                    width: 0,
                    height: 0,
                    borderTop: '5px solid transparent',
                    borderBottom: '5px solid transparent',
                    borderLeft: `6px solid ${getConnectorChevronColor(tokens.palette.success.main, isDark)}`,
                    ml: i > 0 ? -0.3 : 0
                  }}
                />
              ))}
            </Box>

            <Box
              sx={{
                px: 1.5,
                py: 0.75,
                borderRadius: 2,
                bgcolor: tokens.successSurface,
                border: `1px dashed ${tokens.successSoftBorder}`,
                textAlign: 'center',
                minWidth: 60
              }}
            >
              <PublicIcon sx={{ color: theme.palette.success.main, fontSize: 20 }} />
              <Typography variant="caption" display="block" fontWeight={600}>
                {t('subscriptions.chain.preview.internet')}
              </Typography>
            </Box>
          </Box>
        </Collapse>
      </CardContent>
    </Card>
  );
}

RuleChainFlow.propTypes = {
  rule: PropTypes.object.isRequired,
  isMobile: PropTypes.bool.isRequired,
  theme: PropTypes.object.isRequired
};

function NodeMatchTable({ matchSummary, isMobile }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const {
    elevatedSurface,
    tableHeaderSurface,
    softBorder,
    secondaryText,
    tertiaryText,
    successSurface,
    successSoftBorder,
    errorSurface,
    errorSoftBorder
  } = tokens;
  const matchedCount = matchSummary?.filter((n) => !n.unmatched).length || 0;
  const unmatchedCount = matchSummary?.filter((n) => n.unmatched).length || 0;

  return (
    <Box>
      <Stack direction="row" spacing={2} mb={2}>
        <Chip
          icon={<CheckCircleIcon />}
          label={t('subscriptions.chain.preview.matchedCount', { count: matchedCount })}
          size="small"
          sx={{ bgcolor: successSurface, border: '1px solid', borderColor: successSoftBorder, color: 'success.main' }}
        />
        <Chip
          icon={<CancelIcon />}
          label={t('subscriptions.chain.preview.unmatchedCount', { count: unmatchedCount })}
          size="small"
          sx={{ bgcolor: errorSurface, border: '1px solid', borderColor: errorSoftBorder, color: 'error.main' }}
        />
      </Stack>

      <Box
        sx={{
          maxHeight: 300,
          overflow: 'auto',
          borderRadius: 2,
          border: '1px solid',
          borderColor: softBorder,
          bgcolor: elevatedSurface
        }}
      >
        <Table size="small" stickyHeader>
          <TableHead>
            <TableRow>
              <TableCell sx={{ bgcolor: tableHeaderSurface, color: secondaryText }}>
                {t('subscriptions.chain.preview.table.node')}
              </TableCell>
              <TableCell sx={{ bgcolor: tableHeaderSurface, color: secondaryText }}>
                {t('subscriptions.chain.preview.table.matchedRule')}
              </TableCell>
              <TableCell sx={{ bgcolor: tableHeaderSurface, color: secondaryText }}>
                {t('subscriptions.chain.preview.table.entryProxy')}
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {matchSummary?.map((node) => (
              <TableRow key={node.nodeId} sx={{ opacity: node.unmatched ? 0.56 : 1 }}>
                <TableCell>
                  <Stack direction="row" alignItems="center" spacing={0.5}>
                    <Typography variant="caption">{getCountryFlag(node.linkCountry)}</Typography>
                    <Typography
                      variant="body2"
                      sx={{ maxWidth: isMobile ? 100 : 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                    >
                      {node.nodeName}
                    </Typography>
                  </Stack>
                </TableCell>
                <TableCell>
                  {node.unmatched ? (
                    <Typography variant="caption" sx={{ color: tertiaryText }}>
                      {t('common.none', 'None')}
                    </Typography>
                  ) : (
                    <Typography variant="body2">{node.matchedRule}</Typography>
                  )}
                </TableCell>
                <TableCell>
                  {node.entryProxy ? (
                    <Typography variant="body2">{node.entryProxy}</Typography>
                  ) : (
                    <Typography variant="caption" sx={{ color: tertiaryText }}>
                      -
                    </Typography>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Box>
  );
}

NodeMatchTable.propTypes = {
  matchSummary: PropTypes.array,
  isMobile: PropTypes.bool.isRequired
};

export default function ChainPreviewDialog({ open, onClose, loading, data }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const {
    dialogSurface,
    dialogSurfaceGradient,
    mutedPanelSurface,
    elevatedSurface,
    panelBorder,
    softBorder,
    primaryText,
    secondaryText,
    tertiaryText,
    cardShadow
  } = tokens;
  const iconButtonSx = getChainProxyIconButtonSx(tokens);
  const [tab, setTab] = useState(0);

  const rules = useMemo(() => data?.rules || [], [data?.rules]);
  const matchSummary = useMemo(() => data?.matchSummary || [], [data?.matchSummary]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="lg"
      fullWidth
      fullScreen={isMobile}
      slotProps={{
        paper: {
          sx: {
            borderRadius: isMobile ? 0 : 3,
            minHeight: isMobile ? 'auto' : '70vh',
            border: '1px solid',
            borderColor: panelBorder,
            bgcolor: dialogSurface,
            backgroundImage: dialogSurfaceGradient
          }
        }
      }}
    >
      <DialogTitle sx={{ pb: 1, bgcolor: mutedPanelSurface, borderBottom: '1px solid', borderColor: panelBorder }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Stack direction="row" alignItems="center" spacing={1.5}>
            <RouteIcon sx={{ color: 'primary.main' }} />
            <Box>
              <Typography variant={isMobile ? 'subtitle1' : 'h6'} fontWeight={700}>
                {t('subscriptions.chain.preview.title')}
              </Typography>
              {data?.subscriptionName && (
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  {t('subscriptions.chain.preview.subtitle', { name: data.subscriptionName, count: data.totalNodes })}
                </Typography>
              )}
            </Box>
          </Stack>
          <IconButton onClick={onClose} size="small" sx={iconButtonSx}>
            <CloseIcon />
          </IconButton>
        </Stack>
      </DialogTitle>

      <DialogContent sx={{ pt: 0, bgcolor: dialogSurface }}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 200 }}>
            <CircularProgress />
          </Box>
        ) : rules.length === 0 ? (
          <Box sx={{ textAlign: 'center', py: 6, px: 2 }}>
            <AccountTreeIcon sx={{ fontSize: 56, color: secondaryText, mb: 2 }} />
            <Typography variant="h6" sx={{ color: primaryText }}>
              {t('subscriptions.chain.emptyRules')}
            </Typography>
            <Typography variant="body2" sx={{ color: secondaryText }}>
              {t('subscriptions.chain.preview.addRuleHint')}
            </Typography>
          </Box>
        ) : (
          <>
            <Tabs
              value={tab}
              onChange={(_, v) => setTab(v)}
              sx={{
                borderBottom: 1,
                mb: 2,
                mt: 2,
                bgcolor: elevatedSurface,
                borderRadius: 2,
                border: '1px solid',
                borderColor: softBorder,
                boxShadow: cardShadow,
                '& .MuiTab-root': {
                  color: secondaryText,
                  minHeight: 44
                },
                '& .Mui-selected': {
                  color: primaryText
                }
              }}
            >
              <Tab label={t('subscriptions.chain.preview.ruleChainsTab', { count: rules.length })} />
              <Tab label={t('subscriptions.chain.preview.nodeMatchesTab', { count: matchSummary.length })} />
            </Tabs>

            {tab === 0 && (
              <Box>
                <Typography
                  variant="caption"
                  color="text.secondary"
                  display="block"
                  mb={2}
                  sx={{
                    px: 1,
                    py: 0.75,
                    borderRadius: 1.5,
                    bgcolor: elevatedSurface,
                    border: '1px solid',
                    borderColor: softBorder,
                    color: tertiaryText,
                    boxShadow: cardShadow
                  }}
                >
                  {t('subscriptions.chain.preview.canvasHint')}
                </Typography>
                <ChainCanvasView rules={rules} fullscreen={isMobile} />
              </Box>
            )}

            {tab === 1 && <NodeMatchTable matchSummary={matchSummary} isMobile={isMobile} />}
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

ChainPreviewDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  loading: PropTypes.bool,
  data: PropTypes.shape({
    subscriptionName: PropTypes.string,
    totalNodes: PropTypes.number,
    rules: PropTypes.array,
    matchSummary: PropTypes.array
  })
};
