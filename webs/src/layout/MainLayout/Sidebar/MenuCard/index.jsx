import { memo, useState, useEffect, useMemo } from 'react';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import Card from '@mui/material/Card';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import Button from '@mui/material/Button';

// assets
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import NewReleasesIcon from '@mui/icons-material/NewReleases';

// project imports
import useConfig from 'hooks/useConfig';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

// GitHub 仓库配置
const GITHUB_REPO = 'ZeroDeng01/sublinkPro';
const GITHUB_URL = `https://github.com/${GITHUB_REPO}`;
const GITHUB_API_RELEASES = `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`;

// ==============================|| SIDEBAR - VERSION CARD ||============================== //

const getMenuCardTokens = (theme, isDark, hasUpdate, statusColor, versionStatus) => {
  const { palette, dialogSurface, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  return {
    palette,
    primaryText,
    secondaryText,
    tertiaryText,
    emphasizedText: isDark ? withAlpha(primaryText, 0.98) : primaryText,
    cardSurface: isDark
      ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.96)} 0%, ${dialogSurface} 100%)`
      : `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.98)} 0%, ${withAlpha(palette.background.default, 0.9)} 100%)`,
    cardBorder: hasUpdate ? withAlpha(statusColor, isDark ? 0.28 : 0.2) : panelBorder,
    panelBorder,
    statusBorder: withAlpha(statusColor, isDark ? 0.28 : 0.2),
    currentVersionSurface: isDark
      ? `linear-gradient(180deg, ${withAlpha(nestedPanelSurface, 0.66)} 0%, ${withAlpha(mutedPanelSurface, 0.96)} 100%)`
      : withAlpha(palette.background.default, 0.78),
    currentVersionBorder: isDark ? withAlpha(palette.divider, 0.76) : withAlpha(palette.divider, 0.76),
    statusPanelSurface: isDark
      ? `linear-gradient(180deg, ${withAlpha(statusColor, 0.14)} 0%, ${withAlpha(mutedPanelSurface, 0.96)} 100%)`
      : withAlpha(statusColor, versionStatus.key === 'update' ? 0.08 : 0.06),
    iconButtonBackground: isDark ? withAlpha(nestedPanelSurface, 0.42) : withAlpha(palette.background.paper, 0.98),
    iconButtonHoverBackground: isDark
      ? withAlpha(nestedPanelSurface, 0.58)
      : withAlpha(hasUpdate ? statusColor : theme.palette.primary.main, 0.08),
    titleColor: isDark ? withAlpha(primaryText, 0.96) : primaryText,
    mutedTextColor: isDark ? withAlpha(primaryText, 0.72) : secondaryText,
    statusTextColor: isDark ? withAlpha(primaryText, 0.88) : primaryText,
    statusHintColor:
      versionStatus.key === 'update'
        ? isDark
          ? withAlpha(theme.palette.warning.light, 0.9)
          : secondaryText
        : isDark
          ? withAlpha(primaryText, 0.72)
          : secondaryText
  };
};

function MenuCard() {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const palette = theme.vars?.palette || theme.palette;
  const { version } = useConfig();
  const [latestVersion, setLatestVersion] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchLatestVersion = async () => {
      setLoading(true);
      try {
        const res = await fetch(GITHUB_API_RELEASES);
        if (res.ok) {
          const data = await res.json();
          setLatestVersion(data.tag_name || '');
        }
      } catch (error) {
        console.error('获取版本信息失败:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchLatestVersion();
  }, []);

  const currentVersion = version || 'dev';
  const normalizedCurrentVersion = currentVersion.replace(/^v/, '');
  const normalizedLatestVersion = latestVersion.replace(/^v/, '');
  const hasLatestVersion = Boolean(normalizedLatestVersion);
  const hasUpdate = hasLatestVersion && Boolean(normalizedCurrentVersion) && normalizedLatestVersion !== normalizedCurrentVersion;
  const releasesPageHref = `${GITHUB_URL}/releases`;
  const releaseHref = latestVersion ? `${GITHUB_URL}/releases/tag/${latestVersion}` : releasesPageHref;

  const versionStatus = useMemo(() => {
    if (loading) {
      return {
        key: 'loading',
        label: '检查更新中',
        tone: 'info',
        hint: '正在获取最新版本',
        actionLabel: '查看 Releases',
        actionHref: releasesPageHref,
        actionVariant: 'text'
      };
    }

    if (hasUpdate) {
      return {
        key: 'update',
        label: '可更新',
        tone: 'warning',
        hint: latestVersion ? `最新版本 ${latestVersion}` : '发现新版本',
        actionLabel: '查看更新',
        actionHref: releaseHref,
        actionVariant: 'contained'
      };
    }

    if (hasLatestVersion) {
      return {
        key: 'current',
        label: '已是最新',
        tone: 'success',
        hint: latestVersion ? `已同步到 ${latestVersion}` : '无需更新',
        actionLabel: '查看 Releases',
        actionHref: releasesPageHref,
        actionVariant: 'text'
      };
    }

    return {
      key: 'unknown',
      label: '暂未确认',
      tone: 'default',
      hint: '手动查看',
      actionLabel: '查看 Releases',
      actionHref: releasesPageHref,
      actionVariant: 'text'
    };
  }, [hasLatestVersion, hasUpdate, latestVersion, loading, releaseHref, releasesPageHref]);

  const statusToneMap = {
    warning: isDark ? theme.palette.warning.main : theme.palette.warning.dark,
    success: theme.palette.success.main,
    info: theme.palette.info.main,
    default: isDark ? withAlpha(palette.text.primary, 0.72) : palette.text.secondary
  };

  const statusColor = statusToneMap[versionStatus.tone] || statusToneMap.default;
  const {
    cardSurface,
    cardBorder,
    panelBorder,
    statusBorder,
    mutedTextColor,
    statusTextColor,
    emphasizedText,
    statusHintColor,
    statusPanelSurface,
    currentVersionSurface,
    currentVersionBorder,
    titleColor,
    iconButtonBackground,
    iconButtonHoverBackground
  } = getMenuCardTokens(theme, isDark, hasUpdate, statusColor, versionStatus);
  const titleAccentColor = hasUpdate ? statusColor : theme.palette.primary.main;
  const iconButtonColor = hasUpdate ? statusColor : mutedTextColor;
  const ctaBackground = isDark ? alpha(theme.palette.warning.main, 0.18) : theme.palette.warning.dark;
  const ctaHoverBackground = isDark ? alpha(theme.palette.warning.main, 0.26) : theme.palette.warning.main;
  const ctaColor = isDark ? withAlpha(palette.warning.light, 0.98) : theme.palette.common.white;
  const ctaBorderColor = isDark ? alpha(theme.palette.warning.main, 0.28) : alpha(theme.palette.warning.dark, 0.22);
  const statusActionIsContained = versionStatus.actionVariant === 'contained';

  return (
    <Card
      sx={{
        background: cardSurface,
        mb: 2.75,
        overflow: 'hidden',
        position: 'relative',
        boxShadow: isDark
          ? `0 10px 24px ${alpha(theme.palette.common.black, 0.18)}, inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}`
          : `0 1px 3px ${alpha(theme.palette.common.black, 0.05)}`,
        border: `1px solid ${cardBorder}`,
        '&:after': {
          content: '""',
          position: 'absolute',
          width: 120,
          height: 120,
          bgcolor: hasUpdate
            ? alpha(theme.palette.warning.main, isDark ? 0.12 : 0.08)
            : isDark
              ? alpha(theme.palette.primary.main, 0.04)
              : alpha(theme.palette.primary.main, 0.06),
          borderRadius: '50%',
          top: -72,
          right: -68
        }
      }}
    >
      <Box sx={{ p: 1.35, position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1 }}>
          <Avatar
            variant="rounded"
            sx={{
              ...theme.typography.mediumAvatar,
              width: 32,
              height: 32,
              borderRadius: 2,
              color: titleAccentColor,
              border: '1px solid',
              borderColor: hasUpdate ? withAlpha(statusColor, isDark ? 0.3 : 0.2) : alpha(theme.palette.primary.main, isDark ? 0.28 : 0.16),
              bgcolor: hasUpdate
                ? withAlpha(statusColor, isDark ? 0.16 : 0.1)
                : isDark
                  ? alpha(theme.palette.background.paper, 0.16)
                  : alpha(theme.palette.primary.main, 0.06),
              boxShadow: 'none',
              flexShrink: 0
            }}
          >
            {hasUpdate ? <NewReleasesIcon fontSize="small" /> : <InfoOutlinedIcon fontSize="small" />}
          </Avatar>

          <Box sx={{ minWidth: 0, flex: 1 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 0.75 }}>
              <Box sx={{ minWidth: 0 }}>
                <Typography
                  variant="subtitle2"
                  sx={{
                    color: titleColor,
                    fontWeight: 700,
                    lineHeight: 1.2
                  }}
                >
                  SublinkPro
                </Typography>
                <Typography variant="caption" sx={{ display: 'block', color: mutedTextColor, mt: 0.15, lineHeight: 1.2 }}>
                  当前系统版本
                </Typography>
              </Box>

              <Tooltip title="查看 GitHub 仓库">
                <IconButton
                  size="small"
                  component="a"
                  href={GITHUB_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label="查看 GitHub 仓库"
                  sx={{
                    p: 0.5,
                    color: iconButtonColor,
                    background: iconButtonBackground,
                    border: '1px solid',
                    borderColor: panelBorder,
                    '&:hover': {
                      background: iconButtonHoverBackground,
                      color: statusTextColor
                    }
                  }}
                >
                  <OpenInNewIcon sx={{ fontSize: 14 }} />
                </IconButton>
              </Tooltip>
            </Box>

            <Box
              sx={{
                mt: 0.7,
                px: 0.8,
                py: 0.65,
                borderRadius: 1.25,
                background: currentVersionSurface,
                border: '1px solid',
                borderColor: currentVersionBorder,
                boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.02)}` : 'none'
              }}
            >
              <Typography
                variant="body2"
                sx={{
                  color: emphasizedText,
                  fontWeight: 700,
                  lineHeight: 1.2,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}
              >
                {currentVersion}
              </Typography>
            </Box>
          </Box>
        </Box>

        <Box
          sx={{
            mt: 0.8,
            p: 0.9,
            borderRadius: 1.5,
            background: statusPanelSurface,
            border: '1px solid',
            borderColor: statusBorder,
            boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.02)}` : 'none'
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 0.75, minWidth: 0 }}>
            <Box sx={{ minWidth: 0, flex: 1 }}>
              <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 0.65, minWidth: 0 }}>
                <Box
                  sx={{
                    width: 6,
                    height: 6,
                    borderRadius: '50%',
                    bgcolor: statusColor,
                    boxShadow: `0 0 0 3px ${withAlpha(statusColor, isDark ? 0.14 : 0.1)}`,
                    flexShrink: 0,
                    mt: 0.6
                  }}
                />
                <Box sx={{ minWidth: 0, flex: 1 }}>
                  <Typography
                    variant="body2"
                    sx={{
                      color: statusTextColor,
                      fontWeight: 700,
                      lineHeight: 1.2,
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {versionStatus.label}
                  </Typography>
                  <Typography
                    variant="caption"
                    sx={{
                      display: 'block',
                      mt: 0.2,
                      color: statusHintColor,
                      minWidth: 0,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {versionStatus.hint}
                  </Typography>
                </Box>
              </Box>
            </Box>

            {statusActionIsContained ? (
              <Button
                size="small"
                variant="contained"
                color="warning"
                href={versionStatus.actionHref}
                target="_blank"
                rel="noopener noreferrer"
                sx={{
                  minWidth: 'auto',
                  px: 1.05,
                  py: 0.35,
                  fontSize: '0.72rem',
                  lineHeight: 1,
                  fontWeight: 700,
                  color: ctaColor,
                  bgcolor: ctaBackground,
                  border: '1px solid',
                  borderColor: ctaBorderColor,
                  boxShadow: 'none',
                  flexShrink: 0,
                  '&:hover': {
                    bgcolor: ctaHoverBackground,
                    boxShadow: 'none'
                  }
                }}
              >
                {versionStatus.actionLabel}
              </Button>
            ) : (
              <Button
                size="small"
                variant="text"
                href={versionStatus.actionHref}
                target="_blank"
                rel="noopener noreferrer"
                sx={{
                  minWidth: 'auto',
                  px: 0.6,
                  py: 0.25,
                  fontSize: '0.72rem',
                  fontWeight: 600,
                  color: mutedTextColor,
                  '&:hover': {
                    bgcolor: withAlpha(theme.palette.primary.main, isDark ? 0.12 : 0.06),
                    color: statusTextColor
                  },
                  flexShrink: 0
                }}
              >
                {versionStatus.actionLabel}
              </Button>
            )}
          </Box>
        </Box>
      </Box>
    </Card>
  );
}

export default memo(MenuCard);
