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

// icons
import SubscriptionsIcon from '@mui/icons-material/Subscriptions';
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import RefreshIcon from '@mui/icons-material/Refresh';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import { getSubTotal, getNodeTotal } from 'api/total';

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

const PremiumStatCard = ({ title, value, loading, icon: Icon, gradientColors, accentColor, index }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  return (
    <Card
      sx={{
        position: 'relative',
        overflow: 'hidden',
        borderRadius: 4,
        background: isDark
          ? `linear-gradient(145deg, ${alpha(gradientColors[0], 0.15)} 0%, ${alpha(gradientColors[1], 0.08)} 100%)`
          : `linear-gradient(145deg, ${alpha(gradientColors[0], 0.08)} 0%, ${alpha('#fff', 0.95)} 100%)`,
        backdropFilter: 'blur(20px)',
        border: `1px solid ${isDark ? alpha(gradientColors[0], 0.2) : alpha(gradientColors[0], 0.15)}`,
        transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
        animation: `${float} 6s ease-in-out infinite`,
        animationDelay: `${index * 0.3}s`,
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
      <CardContent sx={{ position: 'relative', zIndex: 1, p: 3 }}>
        {/* 背景装饰圆 */}
        <Box
          sx={{
            position: 'absolute',
            top: -30,
            right: -30,
            width: 120,
            height: 120,
            borderRadius: '50%',
            background: `radial-gradient(circle, ${alpha(gradientColors[0], 0.15)} 0%, transparent 70%)`
          }}
        />
        <Box
          sx={{
            position: 'absolute',
            bottom: -20,
            left: -20,
            width: 80,
            height: 80,
            borderRadius: '50%',
            background: `radial-gradient(circle, ${alpha(gradientColors[1], 0.1)} 0%, transparent 70%)`
          }}
        />

        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <Box sx={{ flex: 1 }}>
            {/* 标题 */}
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
              <Box
                sx={{
                  width: 8,
                  height: 8,
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
                  letterSpacing: 1.2,
                  fontSize: '0.75rem'
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
                fontSize: '2.75rem',
                background: `linear-gradient(135deg, ${gradientColors[0]} 0%, ${gradientColors[1]} 100%)`,
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                transition: 'transform 0.3s ease',
                lineHeight: 1.2
              }}
            >
              {loading ? <Skeleton width={80} sx={{ bgcolor: alpha(gradientColors[0], 0.2) }} /> : value.toLocaleString()}
            </Typography>

            {/* 趋势指示器 */}
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 1.5 }}>
              <TrendingUpIcon sx={{ fontSize: 16, color: theme.palette.success.main }} />
              <Typography
                variant="caption"
                sx={{
                  color: theme.palette.success.main,
                  fontWeight: 600
                }}
              >
                运行中
              </Typography>
            </Box>
          </Box>

          {/* 图标 */}
          <Box
            className="stat-icon"
            sx={{
              width: 72,
              height: 72,
              borderRadius: 3,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: `linear-gradient(145deg, ${alpha(gradientColors[0], 0.2)} 0%, ${alpha(gradientColors[1], 0.1)} 100%)`,
              border: `1px solid ${alpha(gradientColors[0], 0.2)}`,
              transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)'
            }}
          >
            <Icon
              sx={{
                fontSize: 36,
                color: gradientColors[0]
              }}
            />
          </Box>
        </Box>

        {/* 底部进度条装饰 */}
        <Box sx={{ mt: 2.5 }}>
          <LinearProgress
            variant="determinate"
            value={loading ? 0 : 100}
            sx={{
              height: 4,
              borderRadius: 2,
              bgcolor: alpha(gradientColors[0], 0.1),
              '& .MuiLinearProgress-bar': {
                borderRadius: 2,
                background: `linear-gradient(90deg, ${gradientColors[0]} 0%, ${gradientColors[1]} 100%)`
              }
            }}
          />
        </Box>
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
  const [releases, setReleases] = useState([]);
  const [loadingStats, setLoadingStats] = useState(true);
  const [loadingReleases, setLoadingReleases] = useState(true);

  const greeting = useMemo(() => getGreeting(), []);

  // 获取统计数据
  const fetchStats = async () => {
    try {
      setLoadingStats(true);
      const [subRes, nodeRes] = await Promise.all([getSubTotal(), getNodeTotal()]);
      setSubTotal(subRes.data || 0);
      setNodeTotal(nodeRes.data || 0);
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
      title: '节点总数',
      value: nodeTotal,
      icon: CloudQueueIcon,
      gradientColors: ['#06b6d4', '#0891b2'],
      accentColor: '#06b6d4'
    }
  ];

  return (
    <Box sx={{ pb: 3 }}>
      {/* 欢迎横幅 */}
      <WelcomeBanner greeting={greeting} />

      {/* 统计卡片 */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        {statsConfig.map((stat, index) => (
          <Grid key={stat.title} size={{ xs: 12, sm: 6, md: 4 }}>
            <PremiumStatCard
              title={stat.title}
              value={stat.value}
              loading={loadingStats}
              icon={stat.icon}
              gradientColors={stat.gradientColors}
              accentColor={stat.accentColor}
              index={index}
            />
          </Grid>
        ))}
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
    </Box>
  );
}
