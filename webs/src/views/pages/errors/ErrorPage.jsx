import { useNavigate } from 'react-router-dom';

// MUI Components
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';

// Icons
import HomeIcon from '@mui/icons-material/Home';
import RefreshIcon from '@mui/icons-material/Refresh';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import BlockIcon from '@mui/icons-material/Block';
import SettingsSuggestIcon from '@mui/icons-material/SettingsSuggest';
import BuildCircleIcon from '@mui/icons-material/BuildCircle';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';

// Styles
import 'assets/scss/error.css';

// Error configurations for different error types
const ERROR_CONFIG = {
  404: {
    code: '404',
    title: '页面未找到',
    description: '抱歉，您访问的页面似乎已经迷失在数字宇宙中了。请检查网址是否正确，或返回首页继续探索。',
    icon: '🚀',
    isAstronaut: true,
    bgClass: 'error-bg-404',
    showStars: true
  },
  401: {
    code: '401',
    title: '需要授权',
    description: '您需要登录才能访问此页面。请先登录您的账户，或联系管理员获取访问权限。',
    Icon: LockOutlinedIcon,
    bgClass: 'error-bg-401'
  },
  403: {
    code: '403',
    title: '禁止访问',
    description: '很抱歉，您没有权限访问此资源。如果您认为这是一个错误，请联系系统管理员。',
    Icon: BlockIcon,
    bgClass: 'error-bg-403'
  },
  500: {
    code: '500',
    title: '服务器错误',
    description: '服务器遇到了一些问题，我们的工程师正在紧急修复中。请稍后再试。',
    Icon: ErrorOutlineIcon,
    bgClass: 'error-bg-500'
  },
  503: {
    code: '503',
    title: '服务维护中',
    description: '系统正在进行维护升级，预计很快就会恢复。感谢您的耐心等待！',
    Icon: BuildCircleIcon,
    bgClass: 'error-bg-503'
  },
  default: {
    code: '???',
    title: '发生错误',
    description: '系统遇到了一些问题，请稍后再试或联系管理员。',
    Icon: SettingsSuggestIcon,
    bgClass: 'error-bg-default'
  }
};

// Floating particles component
function FloatingParticles() {
  return (
    <Box className="error-particles">
      {[...Array(5)].map((_, i) => (
        <Box
          key={i}
          className="error-particle"
          sx={{
            top: `${-20 + Math.random() * 20}%`,
            animationDelay: `${i * 2}s`
          }}
        />
      ))}
    </Box>
  );
}

// Stars background component (for 404)
function StarsBackground() {
  const stars = [...Array(50)].map((_, i) => ({
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    size: Math.random() * 3 + 1,
    delay: Math.random() * 3
  }));

  return (
    <Box className="error-stars">
      {stars.map((star, i) => (
        <Box
          key={i}
          className="error-star"
          sx={{
            left: star.left,
            top: star.top,
            width: star.size,
            height: star.size,
            animationDelay: `${star.delay}s`
          }}
        />
      ))}
    </Box>
  );
}

// Main ErrorPage component
export default function ErrorPage({ statusCode = 404, customTitle, customDescription }) {
  const navigate = useNavigate();

  // Get error configuration
  const config = ERROR_CONFIG[statusCode] || ERROR_CONFIG.default;
  const { code, title, description, icon, Icon, isAstronaut, bgClass, showStars } = config;

  // Use custom title/description if provided
  const displayTitle = customTitle || title;
  const displayDescription = customDescription || description;

  const handleGoHome = () => {
    navigate('/dashboard/default');
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  const handleRefresh = () => {
    window.location.reload();
  };

  return (
    <Box className={`error-page-container ${bgClass}`}>
      {/* Floating particles background */}
      <FloatingParticles />

      {/* Stars for 404 page */}
      {showStars && <StarsBackground />}

      {/* Main error card */}
      <Box className="error-card">
        {/* Icon or Astronaut */}
        <Box className="error-icon-container">
          {isAstronaut ? (
            <Box className="error-astronaut">{icon}</Box>
          ) : Icon ? (
            <Icon className="error-icon" sx={{ fontSize: 64, color: '#fff' }} />
          ) : (
            <Box className="error-astronaut">{icon}</Box>
          )}
        </Box>

        {/* Error Code */}
        <Typography className="error-code" component="h1">
          {code}
        </Typography>

        {/* Error Title */}
        <Typography className="error-title" component="h2">
          {displayTitle}
        </Typography>

        {/* Error Description */}
        <Typography className="error-description">{displayDescription}</Typography>

        {/* Action Buttons */}
        <Box className="error-buttons">
          <Button
            className="error-btn error-btn-primary"
            onClick={handleGoHome}
            startIcon={<HomeIcon />}
            sx={{
              textTransform: 'none',
              borderRadius: '50px',
              px: 3.5,
              py: 1.5,
              fontWeight: 600,
              fontSize: '15px',
              backgroundColor: '#fff',
              color: '#333',
              boxShadow: '0 4px 15px rgba(0, 0, 0, 0.1)',
              '&:hover': {
                backgroundColor: '#fff',
                transform: 'translateY(-3px)',
                boxShadow: '0 8px 25px rgba(0, 0, 0, 0.15)'
              }
            }}
          >
            返回首页
          </Button>

          {statusCode === 500 || statusCode === 503 ? (
            <Button
              className="error-btn error-btn-secondary"
              onClick={handleRefresh}
              startIcon={<RefreshIcon />}
              sx={{
                textTransform: 'none',
                borderRadius: '50px',
                px: 3.5,
                py: 1.5,
                fontWeight: 600,
                fontSize: '15px',
                backgroundColor: 'rgba(255, 255, 255, 0.2)',
                color: '#fff',
                border: '2px solid rgba(255, 255, 255, 0.3)',
                '&:hover': {
                  backgroundColor: 'rgba(255, 255, 255, 0.3)',
                  transform: 'translateY(-3px)'
                }
              }}
            >
              刷新页面
            </Button>
          ) : (
            <Button
              className="error-btn error-btn-secondary"
              onClick={handleGoBack}
              startIcon={<ArrowBackIcon />}
              sx={{
                textTransform: 'none',
                borderRadius: '50px',
                px: 3.5,
                py: 1.5,
                fontWeight: 600,
                fontSize: '15px',
                backgroundColor: 'rgba(255, 255, 255, 0.2)',
                color: '#fff',
                border: '2px solid rgba(255, 255, 255, 0.3)',
                '&:hover': {
                  backgroundColor: 'rgba(255, 255, 255, 0.3)',
                  transform: 'translateY(-3px)'
                }
              }}
            >
              返回上一页
            </Button>
          )}
        </Box>
      </Box>

      {/* Additional decorative elements */}
      {statusCode === 404 && (
        <Box
          sx={{
            position: 'absolute',
            bottom: 24,
            left: '50%',
            transform: 'translateX(-50%)',
            color: 'rgba(255, 255, 255, 0.6)',
            fontSize: '14px',
            textAlign: 'center',
            zIndex: 10
          }}
        >
          🌌 在浩瀚的互联网中，总有一些路径通往未知...
        </Box>
      )}
    </Box>
  );
}
