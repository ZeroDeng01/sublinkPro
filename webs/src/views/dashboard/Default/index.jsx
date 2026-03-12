import { useState, useEffect, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';

// material-ui
import { useTheme, alpha, keyframes } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Skeleton from '@mui/material/Skeleton';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import LinearProgress from '@mui/material/LinearProgress';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';

// icons
import SubscriptionsIcon from '@mui/icons-material/Subscriptions';
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import RefreshIcon from '@mui/icons-material/Refresh';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';
import StarIcon from '@mui/icons-material/Star';
import GitHubIcon from '@mui/icons-material/GitHub';
import BugReportIcon from '@mui/icons-material/BugReport';
import FavoriteIcon from '@mui/icons-material/Favorite';
import FlightTakeoffIcon from '@mui/icons-material/FlightTakeoff';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import EventIcon from '@mui/icons-material/Event';

// icons for protocols
import PublicIcon from '@mui/icons-material/Public';
import FolderIcon from '@mui/icons-material/Folder';
import SourceIcon from '@mui/icons-material/Input';
import LabelIcon from '@mui/icons-material/Label';
import SecurityIcon from '@mui/icons-material/Security';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import TaskProgressPanel from 'components/TaskProgressPanel';
import {
  getSubTotal,
  getNodeTotal,
  getFastestSpeedNode,
  getLowestDelayNode,
  getCountryStats,
  getProtocolStats,
  getTagStats,
  getGroupStats,
  getSourceStats
} from 'api/total';
import { getAirports } from 'api/airports';
import { formatBytes, formatExpireTime, getUsageColor } from 'views/airports/utils';

// ==============================|| 动画定义 ||============================== //

const shimmer = keyframes`
  0% {
    background-position: -200% 0;
  }
  100% {
    background-position: 200% 0;
  }
`;

const float = keyframes`
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-8px);
  }
`;

const pulse = keyframes`
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
`;

const glow = keyframes`
  0%, 100% {
    box-shadow: 0 0 20px rgba(99, 102, 241, 0.3);
  }
  50% {
    box-shadow: 0 0 40px rgba(99, 102, 241, 0.5);
  }
`;

// ==============================|| 问候语计算 ||============================== //

const getGreeting = () => {
  const hour = new Date().getHours();
  if (hour >= 5 && hour < 9) {
    return { text: '早上好', emoji: '🌅', subText: '新的一天开始了' };
  } else if (hour >= 9 && hour < 12) {
    return { text: '上午好', emoji: '☀️', subText: '充满活力的上午' };
  } else if (hour >= 12 && hour < 14) {
    return { text: '中午好', emoji: '🌤️', subText: '记得休息一下' };
  } else if (hour >= 14 && hour < 18) {
    return { text: '下午好', emoji: '🌇', subText: '继续加油' };
  } else if (hour >= 18 && hour < 23) {
    return { text: '晚上好', emoji: '🌙', subText: '辛苦了一天' };
  } else {
    return { text: '夜深了', emoji: '✨', subText: '注意休息' };
  }
};

// ==============================|| 高级统计卡片组件 ||============================== //

const PremiumStatCard = ({
  title,
  value,
  subValue,
  loading,
  icon: Icon,
  gradientColors,
  accentColor,
  index,
  isNodeStat,
  copyLink,
  onCopy
}) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  const handleClick = () => {
    if (isNodeStat && copyLink && onCopy) {
      navigator.clipboard
        .writeText(copyLink)
        .then(() => {
          onCopy('节点链接已复制到剪贴板', 'success');
        })
        .catch(() => {
          onCopy('复制失败', 'error');
        });
    }
  };

  return (
    <Card
      onClick={handleClick}
      sx={{
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 4,
        height: '100%',
        background: isDark
          ? `linear-gradient(145deg, ${alpha(gradientColors[0], 0.15)} 0%, ${alpha(gradientColors[1], 0.08)} 100%)`
          : `linear-gradient(145deg, ${alpha(gradientColors[0], 0.08)} 0%, ${alpha('#fff', 0.95)} 100%)`,
        backdropFilter: 'blur(20px)',
        border: `1px solid ${isDark ? alpha(gradientColors[0], 0.2) : alpha(gradientColors[0], 0.15)}`,
        transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
        animation: `${float} 6s ease-in-out infinite`,
        animationDelay: `${index * 0.3}s`,
        cursor: isNodeStat && copyLink ? 'pointer' : 'default',
        '&:hover': {
          transform: 'translateY(-8px) scale(1.02)',
          boxShadow: `0 20px 40px ${alpha(gradientColors[0], 0.25)}`,
          border: `1px solid ${alpha(gradientColors[0], 0.4)}`,
          '& .stat-icon': {
            transform: 'rotate(10deg) scale(1.1)'
          },
          '& .stat-value': {
            transform: 'scale(1.05)'
          }
        },
        // 顶部彩色边框装饰
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: 4,
          background: `linear-gradient(90deg, ${gradientColors[0]} 0%, ${gradientColors[1]} 100%)`
        },
        // 光泽效果
        '&::after': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: '-100%',
          width: '200%',
          height: '100%',
          background: `linear-gradient(90deg, transparent 0%, ${alpha('#fff', isDark ? 0.03 : 0.1)} 50%, transparent 100%)`,
          animation: `${shimmer} 3s linear infinite`
        }
      }}
    >
      <CardContent sx={{ position: 'relative', zIndex: 1, p: 2.5 }}>
        {/* 背景装饰圆 */}
        <Box
          sx={{
            position: 'absolute',
            top: -30,
            right: -30,
            width: 100,
            height: 100,
            borderRadius: '50%',
            background: `radial-gradient(circle, ${alpha(gradientColors[0], 0.15)} 0%, transparent 70%)`
          }}
        />
        <Box
          sx={{
            position: 'absolute',
            bottom: -20,
            left: -20,
            width: 60,
            height: 60,
            borderRadius: '50%',
            background: `radial-gradient(circle, ${alpha(gradientColors[1], 0.1)} 0%, transparent 70%)`
          }}
        />

        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            {/* 标题 */}
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
              <Box
                sx={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  background: `linear-gradient(135deg, ${gradientColors[0]} 0%, ${gradientColors[1]} 100%)`,
                  animation: `${pulse} 2s ease-in-out infinite`
                }}
              />
              <Typography
                variant="body2"
                sx={{
                  fontWeight: 500,
                  color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                  textTransform: 'uppercase',
                  letterSpacing: 1,
                  fontSize: '0.7rem'
                }}
              >
                {title}
              </Typography>
            </Box>

            {/* 数值 */}
            <Typography
              className="stat-value"
              variant="h1"
              sx={{
                fontWeight: 700,
                fontSize: isNodeStat ? '1.75rem' : '2.25rem',
                background: `linear-gradient(135deg, ${gradientColors[0]} 0%, ${gradientColors[1]} 100%)`,
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                transition: 'transform 0.3s ease',
                lineHeight: 1.2
              }}
            >
              {loading ? (
                <Skeleton width={60} sx={{ bgcolor: alpha(gradientColors[0], 0.2) }} />
              ) : typeof value === 'number' ? (
                value.toLocaleString()
              ) : (
                value
              )}
            </Typography>

            {/* 节点名称/趋势指示器 */}
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 1 }}>
              {isNodeStat && subValue ? (
                <Tooltip title={subValue} arrow placement="bottom">
                  <Typography
                    variant="caption"
                    sx={{
                      color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary,
                      fontWeight: 500,
                      fontSize: '0.7rem',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                      maxWidth: '100%',
                      display: 'block'
                    }}
                  >
                    📍 {subValue}
                  </Typography>
                </Tooltip>
              ) : (
                <>
                  <TrendingUpIcon sx={{ fontSize: 14, color: theme.palette.success.main }} />
                  <Typography
                    variant="caption"
                    sx={{
                      color: theme.palette.success.main,
                      fontWeight: 600,
                      fontSize: '0.7rem'
                    }}
                  >
                    运行中
                  </Typography>
                </>
              )}
            </Box>
          </Box>

          {/* 图标 */}
          <Box
            className="stat-icon"
            sx={{
              width: 56,
              height: 56,
              borderRadius: 2.5,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: `linear-gradient(145deg, ${alpha(gradientColors[0], 0.2)} 0%, ${alpha(gradientColors[1], 0.1)} 100%)`,
              border: `1px solid ${alpha(gradientColors[0], 0.2)}`,
              transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
              flexShrink: 0
            }}
          >
            <Icon
              sx={{
                fontSize: 28,
                color: gradientColors[0]
              }}
            />
          </Box>
        </Box>

        {/* 底部进度条装饰 */}
        <Box sx={{ mt: 2 }}>
          <LinearProgress
            variant="determinate"
            value={loading ? 0 : 100}
            sx={{
              height: 3,
              borderRadius: 1.5,
              bgcolor: alpha(gradientColors[0], 0.1),
              '& .MuiLinearProgress-bar': {
                borderRadius: 1.5,
                background: `linear-gradient(90deg, ${gradientColors[0]} 0%, ${gradientColors[1]} 100%)`
              }
            }}
          />
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| Star 提醒卡片组件 ||============================== //

import { donationConfig } from "config/donation";

const StarReminderCard = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [starCount, setStarCount] = useState(null);

  useEffect(() => {
    const fetchStarCount = async () => {
      try {
        const response = await fetch('https://api.github.com/repos/ZeroDeng01/sublinkPro');
        if (response.ok) {
          const data = await response.json();
          setStarCount(data.stargazers_count);
        }
      } catch (error) {
        console.error('获取Star数量失败:', error);
      }
    };
    fetchStarCount();
  }, []);

  const handleStar = () => {
    window.open('https://github.com/ZeroDeng01/sublinkPro', '_blank');
  };

  const handleFeedback = () => {
    window.open('https://github.com/ZeroDeng01/sublinkPro/issues', '_blank');
  };

  return (
    <Card
      sx={{
        mb: 3,
        borderRadius: 3,
        background: isDark
          ? `linear-gradient(135deg, ${alpha('#fbbf24', 0.15)} 0%, ${alpha('#f59e0b', 0.1)} 50%, ${alpha('#d97706', 0.05)} 100%)`
          : `linear-gradient(135deg, ${alpha('#fef3c7', 0.9)} 0%, ${alpha('#fde68a', 0.6)} 50%, ${alpha('#fcd34d', 0.3)} 100%)`,
        border: `1px solid ${isDark ? alpha('#fbbf24', 0.25) : alpha('#f59e0b', 0.3)}`,
        position: 'relative',
        overflow: 'hidden',
        transition: 'all 0.3s ease',
        '&:hover': {
          transform: 'translateY(-2px)',
          boxShadow: `0 8px 24px ${alpha('#fbbf24', 0.25)}`
        }
      }}
    >
      {/* 背景装饰 */}
      <Box
        sx={{
          position: 'absolute',
          top: -20,
          right: -20,
          width: 100,
          height: 100,
          borderRadius: '50%',
          background: `radial-gradient(circle, ${alpha('#fbbf24', 0.2)} 0%, transparent 70%)`
        }}
      />
      <Box
        sx={{
          position: 'absolute',
          bottom: -30,
          left: '30%',
          width: 80,
          height: 80,
          borderRadius: '50%',
          background: `radial-gradient(circle, ${alpha('#f59e0b', 0.15)} 0%, transparent 70%)`
        }}
      />

      <CardContent sx={{ py: 2.5, px: 3, position: 'relative' }}>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: 2
          }}
        >
          {/* 左侧内容 */}
          <Box sx={{ display: "flex", alignItems: "center", gap: 2, flex: 1, minWidth: { xs: "100%", sm: 280 } }}>
            <Box
              sx={{
                width: 48,
                height: 48,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: `linear-gradient(135deg, ${alpha('#fbbf24', 0.3)} 0%, ${alpha('#f59e0b', 0.2)} 100%)`,
                border: `1px solid ${alpha('#fbbf24', 0.4)}`,
                flexShrink: 0
              }}
            >
              <StarIcon sx={{ fontSize: 28, color: '#f59e0b' }} />
            </Box>
            <Box>
              <Typography
                variant="subtitle1"
                sx={{
                  fontWeight: 600,
                  color: isDark ? '#fcd34d' : '#b45309',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5
                }}
              >
                喜欢这个项目吗？
                <FavoriteIcon sx={{ fontSize: 16, color: '#ef4444' }} />
              </Typography>
              <Typography variant="body2" sx={{ color: isDark ? alpha('#fff', 0.7) : '#92400e' }}>
                如果觉得不错，请给我们一个 Star 支持一下！如果你是L站佬友，也可以使用LDC支持本项目！你的支持是我们前进的动力。
              </Typography>
            </Box>
          </Box>

          {/* 右侧按钮 */}
          <Box
            sx={{
              display: "flex",
              flexDirection: "column",
              alignItems: { xs: "flex-start", sm: "flex-end" },
              gap: 1.5,
              flexShrink: 0,
              width: { xs: "100%", sm: "auto" },
              mt: { xs: 1.5, sm: 0 }
            }}
          >
            {/* 打赏按钮组 */}
            <Box sx={{
              display: "flex",
              alignItems: "center",
              gap: 1,
              flexWrap: "wrap",
              justifyContent: { xs: "flex-start", sm: "flex-end" }
            }}>
              {donationConfig.links.map((item, index) => (
                <Chip
                  key={index}
                  icon={item.icon}
                  label={item.title}
                  component="a"
                  href={item.url}
                  target="_blank"
                  clickable
                  sx={{
                    fontWeight: 600,
                    px: 0.5,
                    height: 36,
                    borderRadius: 2,
                    bgcolor: isDark ? alpha(theme.palette[item.color].main, 0.15) : alpha(theme.palette[item.color].light, 0.5),
                    color: isDark ? theme.palette[item.color].light : theme.palette[item.color].dark,
                    border: `1px solid ${isDark ? alpha(theme.palette[item.color].main, 0.3) : alpha(theme.palette[item.color].main, 0.2)}`,
                    transition: "all 0.2s ease",
                    "&:hover": {
                      bgcolor: isDark ? alpha(theme.palette[item.color].main, 0.25) : alpha(theme.palette[item.color].light, 0.8),
                      transform: "scale(1.05)"
                    },
                    "& .MuiChip-icon": {
                      color: "inherit",
                      fontSize: 18,
                      ml: 1
                    }
                  }}
                />
              ))}
            </Box>

            {/* 工具与 Star 组 */}
            <Box sx={{
              display: "flex",
              alignItems: "center",
              gap: 1.5,
              flexWrap: "wrap",
              justifyContent: { xs: "flex-start", sm: "flex-end" }
            }}>
              <Tooltip title="问题反馈" arrow>
                <IconButton
                  onClick={handleFeedback}
                  size="small"
                  sx={{
                    bgcolor: isDark ? alpha("#fff", 0.1) : alpha("#f59e0b", 0.15),
                    color: isDark ? "#fcd34d" : "#b45309",
                    width: 36,
                    height: 36,
                    borderRadius: 2,
                    "&:hover": {
                      bgcolor: isDark ? alpha("#fff", 0.15) : alpha("#f59e0b", 0.25)
                    }
                  }}
                >
                  <BugReportIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Chip
                icon={<GitHubIcon sx={{ fontSize: 18, color: "inherit !important" }} />}
                label={starCount !== null ? `Star ${starCount >= 1000 ? `${(starCount / 1000).toFixed(1)}k` : starCount}` : "Star"}
                onClick={handleStar}
                sx={{
                  fontWeight: 600,
                  px: 1,
                  height: 36,
                  borderRadius: 2,
                  background: "linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%)",
                  color: "#78350f",
                  border: "none",
                  cursor: "pointer",
                  transition: "all 0.2s ease",
                  "&:hover": {
                    background: "linear-gradient(135deg, #f59e0b 0%, #d97706 100%)",
                    transform: "scale(1.05)"
                  },
                  "& .MuiChip-icon": {
                    color: "#78350f"
                  }
                }}
              />
            </Box>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| 机场流量概览卡片组件 ||============================== //

const AirportUsageCard = ({ airports = [], loading }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  // 筛选开启用量获取且有有效数据的机场
  const airportsWithUsage = useMemo(() => {
    return airports.filter((a) => a.fetchUsageInfo && a.usageTotal > 0);
  }, [airports]);

  // 全局流量汇总
  const { totalUsed, totalQuota, globalPercent } = useMemo(() => {
    const used = airportsWithUsage.reduce((sum, a) => sum + (a.usageUpload || 0) + (a.usageDownload || 0), 0);
    const quota = airportsWithUsage.reduce((sum, a) => sum + a.usageTotal, 0);
    const percent = quota > 0 ? Math.min((used / quota) * 100, 100) : 0;
    return { totalUsed: used, totalQuota: quota, globalPercent: percent };
  }, [airportsWithUsage]);

  // 最近到期机场
  const nearestExpireAirport = useMemo(() => {
    const now = Date.now() / 1000;
    return airportsWithUsage.filter((a) => a.usageExpire > now).sort((a, b) => a.usageExpire - b.usageExpire)[0] || null;
  }, [airportsWithUsage]);

  // 低流量机场 (剩余 < 10%)
  const lowUsageAirports = useMemo(() => {
    return airportsWithUsage.filter((a) => {
      const used = (a.usageUpload || 0) + (a.usageDownload || 0);
      const remaining = a.usageTotal - used;
      return remaining / a.usageTotal < 0.1;
    });
  }, [airportsWithUsage]);

  // 如果没有开启用量获取的机场，不显示此卡片
  if (!loading && airportsWithUsage.length === 0) {
    return null;
  }

  // 根据使用率计算进度条渐变色
  const getProgressGradient = (percent) => {
    if (percent < 60) return `linear-gradient(90deg, ${theme.palette.success.light}, ${theme.palette.success.main})`;
    if (percent < 85) return `linear-gradient(90deg, ${theme.palette.warning.light}, ${theme.palette.warning.main})`;
    return `linear-gradient(90deg, ${theme.palette.error.light}, ${theme.palette.error.main})`;
  };

  return (
    <Card
      sx={{
        mb: 4,
        borderRadius: 4,
        background: isDark
          ? `linear-gradient(145deg, ${alpha('#06b6d4', 0.12)} 0%, ${alpha('#0891b2', 0.06)} 100%)`
          : `linear-gradient(145deg, ${alpha('#06b6d4', 0.08)} 0%, ${alpha('#fff', 0.95)} 100%)`,
        backdropFilter: 'blur(20px)',
        border: `1px solid ${isDark ? alpha('#06b6d4', 0.2) : alpha('#06b6d4', 0.15)}`,
        overflow: 'hidden',
        position: 'relative'
      }}
    >
      {/* 背景装饰 */}
      <Box
        sx={{
          position: 'absolute',
          top: -40,
          right: -40,
          width: 120,
          height: 120,
          borderRadius: '50%',
          background: `radial-gradient(circle, ${alpha('#06b6d4', 0.2)} 0%, transparent 70%)`
        }}
      />

      <CardContent sx={{ p: 3, position: 'relative' }}>
        {/* 标题 */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 3 }}>
          <Box
            sx={{
              width: 40,
              height: 40,
              borderRadius: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)'
            }}
          >
            <FlightTakeoffIcon sx={{ color: '#fff', fontSize: 22 }} />
          </Box>
          <Typography variant="h5" sx={{ fontWeight: 600 }}>
            机场流量概览
          </Typography>
          <Chip
            label={`${airportsWithUsage.length} 个机场`}
            size="small"
            sx={{
              ml: 'auto',
              bgcolor: isDark ? alpha('#06b6d4', 0.2) : alpha('#06b6d4', 0.1),
              color: isDark ? '#67e8f9' : '#0891b2',
              fontWeight: 600
            }}
          />
        </Box>

        {loading ? (
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} variant="rounded" width={200} height={80} sx={{ borderRadius: 2 }} />
            ))}
          </Box>
        ) : (
          <Grid container spacing={3}>
            {/* 全局流量汇总 */}
            <Grid size={{ xs: 12, sm: 6, md: 4 }}>
              <Box
                sx={{
                  p: 2.5,
                  borderRadius: 3,
                  height: '100%',
                  bgcolor: isDark ? alpha('#fff', 0.05) : alpha('#fff', 0.7),
                  border: `1px solid ${isDark ? alpha('#fff', 0.1) : alpha('#06b6d4', 0.15)}`
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <TrendingUpIcon sx={{ fontSize: 18, color: '#06b6d4' }} />
                  <Typography variant="subtitle2" sx={{ color: 'text.secondary', fontWeight: 500 }}>
                    全局流量使用
                  </Typography>
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 700, mb: 1 }}>
                  {formatBytes(totalUsed)} / {formatBytes(totalQuota)}
                </Typography>
                {/* 进度条 */}
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Box
                    sx={{
                      flexGrow: 1,
                      height: 8,
                      borderRadius: 4,
                      backgroundColor: isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                      overflow: 'hidden'
                    }}
                  >
                    <Box
                      sx={{
                        width: `${globalPercent}%`,
                        height: '100%',
                        borderRadius: 4,
                        background: getProgressGradient(globalPercent),
                        transition: 'width 0.3s ease'
                      }}
                    />
                  </Box>
                  <Typography variant="caption" sx={{ fontWeight: 700, color: getUsageColor(globalPercent), minWidth: 45 }}>
                    {globalPercent.toFixed(1)}%
                  </Typography>
                </Box>
              </Box>
            </Grid>

            {/* 最近到期 */}
            <Grid size={{ xs: 12, sm: 6, md: 4 }}>
              <Box
                sx={{
                  p: 2.5,
                  borderRadius: 3,
                  height: '100%',
                  bgcolor: isDark ? alpha('#fff', 0.05) : alpha('#fff', 0.7),
                  border: `1px solid ${isDark ? alpha('#fff', 0.1) : alpha('#06b6d4', 0.15)}`
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <EventIcon sx={{ fontSize: 18, color: '#f59e0b' }} />
                  <Typography variant="subtitle2" sx={{ color: 'text.secondary', fontWeight: 500 }}>
                    最近到期
                  </Typography>
                </Box>
                {nearestExpireAirport ? (
                  <>
                    <Typography variant="h6" sx={{ fontWeight: 600, mb: 0.5, color: isDark ? '#fcd34d' : '#b45309' }}>
                      {nearestExpireAirport.name}
                    </Typography>
                    <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                      {formatExpireTime(nearestExpireAirport.usageExpire)}
                    </Typography>
                  </>
                ) : (
                  <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                    暂无到期信息
                  </Typography>
                )}
              </Box>
            </Grid>

            {/* 低流量警告 */}
            <Grid size={{ xs: 12, sm: 12, md: 4 }}>
              <Box
                sx={{
                  p: 2.5,
                  borderRadius: 3,
                  height: '100%',
                  bgcolor:
                    lowUsageAirports.length > 0
                      ? isDark
                        ? alpha('#ef4444', 0.1)
                        : alpha('#fef2f2', 0.9)
                      : isDark
                        ? alpha('#fff', 0.05)
                        : alpha('#fff', 0.7),
                  border: `1px solid ${lowUsageAirports.length > 0 ? alpha("#ef4444", 0.3) : isDark ? alpha("#fff", 0.1) : alpha("#06b6d4", 0.15)
                  }`
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                  <WarningAmberIcon
                    sx={{
                      fontSize: 18,
                      color: lowUsageAirports.length > 0 ? '#ef4444' : 'text.secondary'
                    }}
                  />
                  <Typography variant="subtitle2" sx={{ color: 'text.secondary', fontWeight: 500 }}>
                    流量不足警告
                  </Typography>
                  {lowUsageAirports.length > 0 && (
                    <Chip
                      label={lowUsageAirports.length}
                      size="small"
                      sx={{
                        ml: 'auto',
                        height: 20,
                        minWidth: 20,
                        bgcolor: '#ef4444',
                        color: '#fff',
                        fontWeight: 700,
                        fontSize: '0.7rem'
                      }}
                    />
                  )}
                </Box>
                {lowUsageAirports.length > 0 ? (
                  <Box sx={{ display: 'flex', gap: 0.75, flexWrap: 'wrap' }}>
                    {lowUsageAirports.map((airport) => {
                      const used = (airport.usageUpload || 0) + (airport.usageDownload || 0);
                      const remaining = airport.usageTotal - used;
                      const remainPercent = ((remaining / airport.usageTotal) * 100).toFixed(1);
                      return (
                        <Tooltip key={airport.id} title={`剩余 ${formatBytes(remaining)} (${remainPercent}%)`} arrow>
                          <Chip
                            label={airport.name}
                            size="small"
                            sx={{
                              bgcolor: isDark ? alpha('#ef4444', 0.2) : alpha('#ef4444', 0.1),
                              color: '#ef4444',
                              fontWeight: 600,
                              fontSize: '0.75rem',
                              '&:hover': {
                                bgcolor: isDark ? alpha('#ef4444', 0.3) : alpha('#ef4444', 0.2)
                              }
                            }}
                          />
                        </Tooltip>
                      );
                    })}
                  </Box>
                ) : (
                  <Typography variant="body2" sx={{ color: isDark ? '#86efac' : '#16a34a' }}>
                    ✓ 所有机场流量充足
                  </Typography>
                )}
              </Box>
            </Grid>
          </Grid>
        )}
      </CardContent>
    </Card>
  );
};

// ==============================|| 欢迎横幅组件 ||============================== //

const WelcomeBanner = ({ greeting }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  return (
    <Card
      sx={{
        mb: 4,
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 4,
        background: isDark
          ? `linear-gradient(135deg, ${alpha('#6366f1', 0.2)} 0%, ${alpha('#8b5cf6', 0.15)} 50%, ${alpha('#a855f7', 0.1)} 100%)`
          : `linear-gradient(135deg, ${alpha('#6366f1', 0.12)} 0%, ${alpha('#8b5cf6', 0.08)} 50%, ${alpha('#a855f7', 0.05)} 100%)`,
        backdropFilter: 'blur(20px)',
        border: `1px solid ${isDark ? alpha('#6366f1', 0.2) : alpha('#6366f1', 0.15)}`,
        animation: `${glow} 4s ease-in-out infinite`
      }}
    >
      {/* 背景装饰图案 */}
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          opacity: 0.5,
          background: `
            radial-gradient(circle at 20% 20%, ${alpha('#6366f1', 0.15)} 0%, transparent 50%),
            radial-gradient(circle at 80% 80%, ${alpha('#a855f7', 0.1)} 0%, transparent 50%),
            radial-gradient(circle at 50% 50%, ${alpha('#8b5cf6', 0.08)} 0%, transparent 70%)
          `
        }}
      />

      {/* 网格装饰 */}
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          opacity: isDark ? 0.03 : 0.05,
          backgroundImage: `
            linear-gradient(to right, ${theme.palette.primary.main} 1px, transparent 1px),
            linear-gradient(to bottom, ${theme.palette.primary.main} 1px, transparent 1px)
          `,
          backgroundSize: '40px 40px'
        }}
      />

      <CardContent sx={{ position: 'relative', zIndex: 1, py: 5, px: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 3 }}>
          <Box>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1 }}>
              <Typography
                variant="h1"
                sx={{
                  fontWeight: 800,
                  fontSize: { xs: '2rem', sm: '2.5rem', md: '3rem' },
                  background: isDark
                    ? 'linear-gradient(135deg, #fff 0%, #e0e7ff 100%)'
                    : 'linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%)',
                  backgroundClip: 'text',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  lineHeight: 1.2
                }}
              >
                {greeting.text}
              </Typography>
              <Typography
                sx={{
                  fontSize: { xs: '2rem', sm: '2.5rem', md: '3rem' },
                  animation: `${float} 3s ease-in-out infinite`
                }}
              >
                {greeting.emoji}
              </Typography>
            </Box>
            <Typography
              variant="body1"
              sx={{
                color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                fontSize: '1.1rem'
              }}
            >
              欢迎使用{' '}
              <Box component="span" sx={{ fontWeight: 700, color: isDark ? '#a5b4fc' : '#6366f1' }}>
                SublinkPro
              </Box>{' '}
              订阅管理系统，{greeting.subText}
            </Typography>
          </Box>

          {/* 装饰图标 */}
          <Box
            sx={{
              display: { xs: 'none', md: 'flex' },
              alignItems: 'center',
              justifyContent: 'center',
              width: 80,
              height: 80,
              borderRadius: '50%',
              background: `linear-gradient(145deg, ${alpha('#6366f1', 0.2)} 0%, ${alpha('#a855f7', 0.1)} 100%)`,
              border: `1px solid ${alpha('#6366f1', 0.3)}`,
              animation: `${float} 4s ease-in-out infinite`
            }}
          >
            <AutoAwesomeIcon sx={{ fontSize: 40, color: isDark ? '#a5b4fc' : '#6366f1' }} />
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| 发布日志组件 ||============================== //

const ReleaseCard = ({ release }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  return (
    <Card
      sx={{
        mb: 2.5,
        borderRadius: 3,
        background: isDark ? alpha(theme.palette.background.paper, 0.6) : alpha('#fff', 0.9),
        backdropFilter: 'blur(10px)',
        border: `1px solid ${isDark ? alpha('#fff', 0.08) : alpha('#000', 0.06)}`,
        transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
        '&:hover': {
          transform: 'translateX(8px)',
          boxShadow: theme.shadows[8],
          borderColor: theme.palette.primary.main
        }
      }}
    >
      <CardContent sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Chip
            label={release.tag_name}
            size="small"
            sx={{
              fontWeight: 700,
              background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
              color: 'white',
              borderRadius: 2,
              px: 0.5
            }}
          />
          <Typography variant="subtitle1" sx={{ fontWeight: 600, flex: 1 }}>
            {release.name}
          </Typography>
          <Chip
            label={new Date(release.published_at).toLocaleDateString('zh-CN', {
              month: 'short',
              day: 'numeric'
            })}
            size="small"
            variant="outlined"
            sx={{ borderRadius: 2 }}
          />
          <Tooltip title="在 GitHub 查看" arrow>
            <IconButton
              size="small"
              component="a"
              href={release.html_url}
              target="_blank"
              rel="noopener noreferrer"
              sx={{
                color: theme.palette.primary.main,
                '&:hover': {
                  background: alpha(theme.palette.primary.main, 0.1)
                }
              }}
            >
              <OpenInNewIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Box>
        <Divider sx={{ mb: 2, opacity: 0.5 }} />
        <Box
          sx={{
            '& h1, & h2, & h3': {
              fontSize: '1rem',
              fontWeight: 600,
              mt: 1.5,
              mb: 0.5,
              color: theme.palette.text.primary
            },
            '& p': {
              mb: 1,
              fontSize: '0.875rem',
              lineHeight: 1.7,
              color: theme.palette.text.secondary
            },
            '& ul, & ol': {
              pl: 2.5,
              mb: 1
            },
            '& li': {
              fontSize: '0.875rem',
              mb: 0.5,
              color: theme.palette.text.secondary,
              '&::marker': {
                color: theme.palette.primary.main
              }
            },
            '& code': {
              backgroundColor: isDark ? alpha('#fff', 0.1) : alpha('#6366f1', 0.1),
              color: isDark ? '#a5b4fc' : '#6366f1',
              padding: '2px 8px',
              borderRadius: 6,
              fontSize: '0.8rem',
              fontFamily: '"JetBrains Mono", monospace'
            },
            '& pre': {
              backgroundColor: isDark ? alpha('#000', 0.3) : alpha('#f1f5f9', 0.8),
              padding: 2,
              borderRadius: 2,
              overflow: 'auto',
              border: `1px solid ${isDark ? alpha('#fff', 0.1) : alpha('#000', 0.05)}`,
              '& code': {
                backgroundColor: 'transparent',
                padding: 0
              }
            },
            '& a': {
              color: theme.palette.primary.main,
              textDecoration: 'none',
              fontWeight: 500,
              '&:hover': {
                textDecoration: 'underline'
              }
            }
          }}
        >
          <ReactMarkdown>{release.body || '暂无更新说明'}</ReactMarkdown>
        </Box>
      </CardContent>
    </Card>
  );
};

// ==============================|| 仪表盘默认页面 ||============================== //

export default function DashboardDefault() {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [subTotal, setSubTotal] = useState(0);
  const [nodeTotal, setNodeTotal] = useState(0);
  const [nodeAvailable, setNodeAvailable] = useState(0);
  const [fastestNode, setFastestNode] = useState(null);
  const [lowestDelayNode, setLowestDelayNode] = useState(null);
  const [countryStats, setCountryStats] = useState({});
  const [protocolStats, setProtocolStats] = useState({});
  const [tagStats, setTagStats] = useState([]);
  const [groupStats, setGroupStats] = useState({});
  const [sourceStats, setSourceStats] = useState({});
  const [releases, setReleases] = useState([]);
  const [airports, setAirports] = useState([]);
  const [loadingStats, setLoadingStats] = useState(true);
  const [loadingReleases, setLoadingReleases] = useState(true);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const greeting = useMemo(() => getGreeting(), []);

  // 显示提示消息
  const showSnackbar = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  // 获取统计数据
  const fetchStats = async () => {
    try {
      setLoadingStats(true);
      const [subRes, nodeRes, fastestRes, lowestDelayRes, countryRes, protocolRes, tagRes, groupRes, sourceRes, airportRes] =
        await Promise.all([
          getSubTotal(),
          getNodeTotal(),
          getFastestSpeedNode(),
          getLowestDelayNode(),
          getCountryStats(),
          getProtocolStats(),
          getTagStats(),
          getGroupStats(),
          getSourceStats(),
          getAirports()
        ]);
      setSubTotal(subRes.data || 0);
      // nodeRes.data 现在返回 { total, available }
      if (nodeRes.data && typeof nodeRes.data === 'object') {
        setNodeTotal(nodeRes.data.total || 0);
        setNodeAvailable(nodeRes.data.available || 0);
      } else {
        setNodeTotal(nodeRes.data || 0);
        setNodeAvailable(0);
      }
      setFastestNode(fastestRes.data || null);
      setLowestDelayNode(lowestDelayRes.data || null);
      setCountryStats(countryRes.data || {});
      setProtocolStats(protocolRes.data || {});
      setTagStats(tagRes.data || []);
      setGroupStats(groupRes.data || {});
      setSourceStats(sourceRes.data || {});
      setAirports(airportRes.data?.list || airportRes.data || []);
    } catch (error) {
      console.error('获取统计数据失败:', error);
    } finally {
      setLoadingStats(false);
    }
  };

  // 获取 GitHub 发布日志
  const fetchReleases = async () => {
    try {
      setLoadingReleases(true);
      const response = await fetch('https://api.github.com/repos/ZeroDeng01/sublinkPro/releases?per_page=5');
      if (!response.ok) throw new Error('Failed to fetch releases');
      const data = await response.json();
      setReleases(data);
    } catch (error) {
      console.error('获取发布日志失败:', error);
      setReleases([]);
    } finally {
      setLoadingReleases(false);
    }
  };

  useEffect(() => {
    fetchStats();
    fetchReleases();
  }, []);

  // 统计卡片配置
  const statsConfig = [
    {
      title: '订阅总数',
      value: subTotal,
      icon: SubscriptionsIcon,
      gradientColors: ['#6366f1', '#8b5cf6'],
      accentColor: '#6366f1'
    },
    {
      title: '节点统计',
      value: `${nodeAvailable} / ${nodeTotal}`,
      subValue: '测速通过 / 总节点',
      icon: CloudQueueIcon,
      gradientColors: ['#06b6d4', '#0891b2'],
      accentColor: '#06b6d4',
      isNodeStat: true
    },
    {
      title: '最快速度',
      value: fastestNode?.Speed ? `${fastestNode.Speed.toFixed(2)} MB/s` : '--',
      subValue: fastestNode?.Name || '暂无数据',
      icon: SpeedIcon,
      gradientColors: ['#10b981', '#059669'],
      accentColor: '#10b981',
      isNodeStat: true,
      copyLink: fastestNode?.Link
    },
    {
      title: '最低延迟',
      value: lowestDelayNode?.DelayTime ? `${lowestDelayNode.DelayTime} ms` : '--',
      subValue: lowestDelayNode?.Name || '暂无数据',
      icon: TimerIcon,
      gradientColors: ['#f59e0b', '#d97706'],
      accentColor: '#f59e0b',
      isNodeStat: true,
      copyLink: lowestDelayNode?.Link
    }
  ];

  return (
    <Box sx={{ pb: 3 }}>
      {/* 欢迎横幅 */}
      <WelcomeBanner greeting={greeting} />

      {/* Star 提醒卡片 */}
      <StarReminderCard />

      {/* 任务进度面板 */}
      <TaskProgressPanel />

      {/* 统计卡片 */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        {statsConfig.map((stat, index) => (
          <Grid key={stat.title} size={{ xs: 12, sm: 6, md: 3 }}>
            <PremiumStatCard
              title={stat.title}
              value={stat.value}
              subValue={stat.subValue}
              loading={loadingStats}
              icon={stat.icon}
              gradientColors={stat.gradientColors}
              accentColor={stat.accentColor}
              index={index}
              isNodeStat={stat.isNodeStat}
              copyLink={stat.copyLink}
              onCopy={showSnackbar}
            />
          </Grid>
        ))}
      </Grid>

      {/* 机场流量概览卡片 */}
      <AirportUsageCard airports={airports} loading={loadingStats} />

      {/* 国家和协议统计 */}
      <Grid container spacing={3} sx={{ mb: 4, alignItems: 'stretch' }}>
        {/* 国家统计卡片 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card
            sx={{
              borderRadius: 4,
              height: '100%',
              background: isDark
                ? `linear-gradient(145deg, ${alpha('#6366f1', 0.12)} 0%, ${alpha('#8b5cf6', 0.06)} 100%)`
                : `linear-gradient(145deg, ${alpha('#6366f1', 0.06)} 0%, ${alpha('#fff', 0.95)} 100%)`,
              backdropFilter: 'blur(20px)',
              border: `1px solid ${isDark ? alpha('#6366f1', 0.2) : alpha('#6366f1', 0.12)}`,
              overflow: 'hidden',
              position: 'relative'
            }}
          >
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2.5 }}>
                <Box
                  sx={{
                    width: 40,
                    height: 40,
                    borderRadius: 2,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)'
                  }}
                >
                  <PublicIcon sx={{ color: '#fff', fontSize: 22 }} />
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 600 }}>
                  节点国家分布
                </Typography>
              </Box>

              {loadingStats ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {[1, 2, 3, 4, 5].map((i) => (
                    <Skeleton key={i} variant="rounded" width={80} height={36} sx={{ borderRadius: 2 }} />
                  ))}
                </Box>
              ) : Object.keys(countryStats).length > 0 ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {Object.entries(countryStats)
                    .sort((a, b) => b[1] - a[1])
                    .map(([country, count]) => {
                      // 国家代码转国旗 emoji
                      const getFlagEmoji = (code) => {
                        if (!code || code === '未知') return '🌐';
                        code = code.toUpperCase() === 'TW' ? 'CN' : code;
                        const codePoints = code
                          .toUpperCase()
                          .split('')
                          .map((char) => 127397 + char.charCodeAt(0));
                        return String.fromCodePoint(...codePoints);
                      };
                      return (
                        <Chip
                          key={country}
                          label={
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                              <Typography sx={{ fontSize: '1rem' }}>{getFlagEmoji(country)}</Typography>
                              <Typography sx={{ fontWeight: 600, fontSize: '0.8rem' }}>{country}</Typography>
                              <Typography sx={{ color: 'text.secondary', fontSize: '0.75rem', ml: 0.5 }}>({count})</Typography>
                            </Box>
                          }
                          sx={{
                            bgcolor: isDark ? alpha('#6366f1', 0.15) : alpha('#6366f1', 0.08),
                            border: `1px solid ${alpha('#6366f1', 0.2)}`,
                            borderRadius: 2,
                            height: 36,
                            '&:hover': {
                              bgcolor: isDark ? alpha('#6366f1', 0.25) : alpha('#6366f1', 0.15)
                            }
                          }}
                        />
                      );
                    })}
                </Box>
              ) : (
                <Typography color="text.secondary" sx={{ fontSize: '0.875rem' }}>
                  暂无国家统计数据
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* 协议统计卡片 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card
            sx={{
              borderRadius: 4,
              height: '100%',
              background: isDark
                ? `linear-gradient(145deg, ${alpha('#10b981', 0.12)} 0%, ${alpha('#059669', 0.06)} 100%)`
                : `linear-gradient(145deg, ${alpha('#10b981', 0.06)} 0%, ${alpha('#fff', 0.95)} 100%)`,
              backdropFilter: 'blur(20px)',
              border: `1px solid ${isDark ? alpha('#10b981', 0.2) : alpha('#10b981', 0.12)}`,
              overflow: 'hidden',
              position: 'relative'
            }}
          >
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2.5 }}>
                <Box
                  sx={{
                    width: 40,
                    height: 40,
                    borderRadius: 2,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)'
                  }}
                >
                  <SecurityIcon sx={{ color: '#fff', fontSize: 22 }} />
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 600 }}>
                  节点协议分布
                </Typography>
              </Box>

              {loadingStats ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {[1, 2, 3, 4].map((i) => (
                    <Skeleton key={i} variant="rounded" width={100} height={36} sx={{ borderRadius: 2 }} />
                  ))}
                </Box>
              ) : Object.keys(protocolStats).length > 0 ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {Object.entries(protocolStats)
                    .sort((a, b) => b[1] - a[1])
                    .map(([protocol, count]) => {
                      // 协议颜色映射
                      const protocolColors = {
                        Shadowsocks: ['#3b82f6', '#2563eb'],
                        ShadowsocksR: ['#6366f1', '#4f46e5'],
                        VMess: ['#8b5cf6', '#7c3aed'],
                        VLESS: ['#10b981', '#059669'],
                        Trojan: ['#ef4444', '#dc2626'],
                        Hysteria: ['#06b6d4', '#0891b2'],
                        Hysteria2: ['#14b8a6', '#0d9488'],
                        TUIC: ['#f59e0b', '#d97706'],
                        WireGuard: ['#84cc16', '#65a30d'],
                        NaiveProxy: ['#ec4899', '#db2777'],
                        SOCKS5: ['#64748b', '#475569'],
                        HTTP: ['#94a3b8', '#64748b'],
                        HTTPS: ['#22c55e', '#16a34a']
                      };
                      const colors = protocolColors[protocol] || ['#6b7280', '#4b5563'];

                      return (
                        <Chip
                          key={protocol}
                          label={
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                              <Box
                                sx={{
                                  width: 8,
                                  height: 8,
                                  borderRadius: '50%',
                                  background: `linear-gradient(135deg, ${colors[0]} 0%, ${colors[1]} 100%)`
                                }}
                              />
                              <Typography sx={{ fontWeight: 600, fontSize: '0.8rem' }}>{protocol}</Typography>
                              <Typography sx={{ color: 'text.secondary', fontSize: '0.75rem', ml: 0.5 }}>({count})</Typography>
                            </Box>
                          }
                          sx={{
                            bgcolor: isDark ? alpha(colors[0], 0.15) : alpha(colors[0], 0.08),
                            border: `1px solid ${alpha(colors[0], 0.25)}`,
                            borderRadius: 2,
                            height: 36,
                            '&:hover': {
                              bgcolor: isDark ? alpha(colors[0], 0.25) : alpha(colors[0], 0.15)
                            }
                          }}
                        />
                      );
                    })}
                </Box>
              ) : (
                <Typography color="text.secondary" sx={{ fontSize: '0.875rem' }}>
                  暂无协议统计数据
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* 标签、分组、来源统计 */}
      <Grid container spacing={3} sx={{ mb: 4, alignItems: 'stretch' }}>
        {/* 标签统计卡片 */}
        <Grid size={{ xs: 12, md: 4 }}>
          <Card
            sx={{
              borderRadius: 4,
              height: '100%',
              background: isDark
                ? `linear-gradient(145deg, ${alpha('#ec4899', 0.12)} 0%, ${alpha('#db2777', 0.06)} 100%)`
                : `linear-gradient(145deg, ${alpha('#ec4899', 0.06)} 0%, ${alpha('#fff', 0.95)} 100%)`,
              backdropFilter: 'blur(20px)',
              border: `1px solid ${isDark ? alpha('#ec4899', 0.2) : alpha('#ec4899', 0.12)}`,
              overflow: 'hidden',
              position: 'relative'
            }}
          >
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2.5 }}>
                <Box
                  sx={{
                    width: 40,
                    height: 40,
                    borderRadius: 2,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'linear-gradient(135deg, #ec4899 0%, #db2777 100%)'
                  }}
                >
                  <LabelIcon sx={{ color: '#fff', fontSize: 22 }} />
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 600 }}>
                  标签统计
                </Typography>
              </Box>

              {loadingStats ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {[1, 2, 3].map((i) => (
                    <Skeleton key={i} variant="rounded" width={80} height={36} sx={{ borderRadius: 2 }} />
                  ))}
                </Box>
              ) : tagStats.length > 0 ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {tagStats
                    .sort((a, b) => b.count - a.count)
                    .map((tag) => (
                      <Chip
                        key={tag.name}
                        label={
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                            <Box
                              sx={{
                                width: 10,
                                height: 10,
                                borderRadius: '50%',
                                bgcolor: tag.color
                              }}
                            />
                            <Typography sx={{ fontWeight: 600, fontSize: '0.8rem' }}>{tag.name}</Typography>
                            <Typography sx={{ color: 'text.secondary', fontSize: '0.75rem', ml: 0.5 }}>({tag.count})</Typography>
                          </Box>
                        }
                        sx={{
                          bgcolor: isDark ? alpha(tag.color, 0.15) : alpha(tag.color, 0.1),
                          border: `1px solid ${alpha(tag.color, 0.3)}`,
                          borderRadius: 2,
                          height: 36,
                          '&:hover': {
                            bgcolor: isDark ? alpha(tag.color, 0.25) : alpha(tag.color, 0.2)
                          }
                        }}
                      />
                    ))}
                </Box>
              ) : (
                <Typography color="text.secondary" sx={{ fontSize: '0.875rem' }}>
                  暂无标签统计数据
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* 分组统计卡片 */}
        <Grid size={{ xs: 12, md: 4 }}>
          <Card
            sx={{
              borderRadius: 4,
              height: '100%',
              background: isDark
                ? `linear-gradient(145deg, ${alpha('#8b5cf6', 0.12)} 0%, ${alpha('#7c3aed', 0.06)} 100%)`
                : `linear-gradient(145deg, ${alpha('#8b5cf6', 0.06)} 0%, ${alpha('#fff', 0.95)} 100%)`,
              backdropFilter: 'blur(20px)',
              border: `1px solid ${isDark ? alpha('#8b5cf6', 0.2) : alpha('#8b5cf6', 0.12)}`,
              overflow: 'hidden',
              position: 'relative'
            }}
          >
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2.5 }}>
                <Box
                  sx={{
                    width: 40,
                    height: 40,
                    borderRadius: 2,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)'
                  }}
                >
                  <FolderIcon sx={{ color: '#fff', fontSize: 22 }} />
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 600 }}>
                  分组统计
                </Typography>
              </Box>

              {loadingStats ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {[1, 2, 3].map((i) => (
                    <Skeleton key={i} variant="rounded" width={80} height={36} sx={{ borderRadius: 2 }} />
                  ))}
                </Box>
              ) : Object.keys(groupStats).length > 0 ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {Object.entries(groupStats)
                    .sort((a, b) => b[1] - a[1])
                    .map(([group, count]) => (
                      <Chip
                        key={group}
                        label={
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                            <Typography sx={{ fontWeight: 600, fontSize: '0.8rem' }}>{group}</Typography>
                            <Typography sx={{ color: 'text.secondary', fontSize: '0.75rem', ml: 0.5 }}>({count})</Typography>
                          </Box>
                        }
                        sx={{
                          bgcolor: isDark ? alpha('#8b5cf6', 0.15) : alpha('#8b5cf6', 0.08),
                          border: `1px solid ${alpha('#8b5cf6', 0.2)}`,
                          borderRadius: 2,
                          height: 36,
                          '&:hover': {
                            bgcolor: isDark ? alpha('#8b5cf6', 0.25) : alpha('#8b5cf6', 0.15)
                          }
                        }}
                      />
                    ))}
                </Box>
              ) : (
                <Typography color="text.secondary" sx={{ fontSize: '0.875rem' }}>
                  暂无分组统计数据
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* 来源统计卡片 */}
        <Grid size={{ xs: 12, md: 4 }}>
          <Card
            sx={{
              borderRadius: 4,
              height: '100%',
              background: isDark
                ? `linear-gradient(145deg, ${alpha('#f97316', 0.12)} 0%, ${alpha('#ea580c', 0.06)} 100%)`
                : `linear-gradient(145deg, ${alpha('#f97316', 0.06)} 0%, ${alpha('#fff', 0.95)} 100%)`,
              backdropFilter: 'blur(20px)',
              border: `1px solid ${isDark ? alpha('#f97316', 0.2) : alpha('#f97316', 0.12)}`,
              overflow: 'hidden',
              position: 'relative'
            }}
          >
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2.5 }}>
                <Box
                  sx={{
                    width: 40,
                    height: 40,
                    borderRadius: 2,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)'
                  }}
                >
                  <SourceIcon sx={{ color: '#fff', fontSize: 22 }} />
                </Box>
                <Typography variant="h5" sx={{ fontWeight: 600 }}>
                  来源统计
                </Typography>
              </Box>

              {loadingStats ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {[1, 2, 3].map((i) => (
                    <Skeleton key={i} variant="rounded" width={80} height={36} sx={{ borderRadius: 2 }} />
                  ))}
                </Box>
              ) : Object.keys(sourceStats).length > 0 ? (
                <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
                  {Object.entries(sourceStats)
                    .sort((a, b) => b[1] - a[1])
                    .map(([source, count]) => (
                      <Chip
                        key={source}
                        label={
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                            <Typography sx={{ fontWeight: 600, fontSize: '0.8rem' }}>{source}</Typography>
                            <Typography sx={{ color: 'text.secondary', fontSize: '0.75rem', ml: 0.5 }}>({count})</Typography>
                          </Box>
                        }
                        sx={{
                          bgcolor: isDark ? alpha('#f97316', 0.15) : alpha('#f97316', 0.08),
                          border: `1px solid ${alpha('#f97316', 0.2)}`,
                          borderRadius: 2,
                          height: 36,
                          '&:hover': {
                            bgcolor: isDark ? alpha('#f97316', 0.25) : alpha('#f97316', 0.15)
                          }
                        }}
                      />
                    ))}
                </Box>
              ) : (
                <Typography color="text.secondary" sx={{ fontSize: '0.875rem' }}>
                  暂无来源统计数据
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* 更新日志 */}
      <MainCard
        title={
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
            <Box
              sx={{
                width: 36,
                height: 36,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)'
              }}
            >
              <Typography sx={{ fontSize: '1.2rem' }}>📝</Typography>
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 600 }}>
              更新日志
            </Typography>
          </Box>
        }
        secondary={
          <Tooltip title="刷新" arrow>
            <Box component="span" sx={{ display: 'inline-block' }}>
              <IconButton
                onClick={fetchReleases}
                disabled={loadingReleases}
                sx={{
                  transition: 'all 0.3s ease',
                  '&:hover': {
                    transform: 'rotate(180deg)',
                    background: alpha(theme.palette.primary.main, 0.1)
                  }
                }}
              >
                <RefreshIcon />
              </IconButton>
            </Box>
          </Tooltip>
        }
        sx={{
          borderRadius: 4,
          overflow: 'hidden',
          '& .MuiCardHeader-root': {
            borderBottom: `1px solid ${isDark ? alpha('#fff', 0.08) : alpha('#000', 0.06)}`
          }
        }}
      >
        {loadingReleases ? (
          <Box>
            {[1, 2, 3].map((i) => (
              <Box key={i} sx={{ mb: 2.5 }}>
                <Skeleton
                  variant="rectangular"
                  height={140}
                  sx={{
                    borderRadius: 3,
                    bgcolor: isDark ? alpha('#fff', 0.05) : alpha('#000', 0.04)
                  }}
                />
              </Box>
            ))}
          </Box>
        ) : releases.length > 0 ? (
          releases.map((release) => <ReleaseCard key={release.id} release={release} />)
        ) : (
          <Box
            sx={{
              textAlign: 'center',
              py: 8,
              px: 3
            }}
          >
            <Typography
              sx={{
                fontSize: '3rem',
                mb: 2
              }}
            >
              📭
            </Typography>
            <Typography variant="h6" color="textSecondary" sx={{ fontWeight: 500 }}>
              暂无更新日志
            </Typography>
            <Typography variant="body2" color="textSecondary" sx={{ mt: 1 }}>
              请检查网络连接或稍后重试
            </Typography>
          </Box>
        )}
      </MainCard>

      {/* 复制成功提示 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={2000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setSnackbar({ ...snackbar, open: false })} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
