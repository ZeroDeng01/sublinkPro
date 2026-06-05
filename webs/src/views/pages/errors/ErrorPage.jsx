import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

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
    title: 'Page not found',
    titleKey: 'errorPage.404.title',
    description: 'The page you requested could not be found. Check the URL or return home.',
    descriptionKey: 'errorPage.404.description',
    icon: '🚀',
    isAstronaut: true,
    bgClass: 'error-bg-404',
    showStars: true
  },
  401: {
    code: '401',
    title: 'Authorization required',
    titleKey: 'errorPage.401.title',
    description: 'You need to sign in before accessing this page.',
    descriptionKey: 'errorPage.401.description',
    Icon: LockOutlinedIcon,
    bgClass: 'error-bg-401'
  },
  403: {
    code: '403',
    title: 'Access forbidden',
    titleKey: 'errorPage.403.title',
    description: 'You do not have permission to access this resource.',
    descriptionKey: 'errorPage.403.description',
    Icon: BlockIcon,
    bgClass: 'error-bg-403'
  },
  500: {
    code: '500',
    title: 'Server error',
    titleKey: 'errorPage.500.title',
    description: 'The server encountered a problem. Try again later.',
    descriptionKey: 'errorPage.500.description',
    Icon: ErrorOutlineIcon,
    bgClass: 'error-bg-500'
  },
  503: {
    code: '503',
    title: 'Service under maintenance',
    titleKey: 'errorPage.503.title',
    description: 'The system is under maintenance and should be back soon.',
    descriptionKey: 'errorPage.503.description',
    Icon: BuildCircleIcon,
    bgClass: 'error-bg-503'
  },
  default: {
    code: '???',
    title: 'Something went wrong',
    titleKey: 'errorPage.default.title',
    description: 'The system encountered a problem. Try again later or contact an administrator.',
    descriptionKey: 'errorPage.default.description',
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
  const { t } = useTranslation();

  // Get error configuration
  const config = ERROR_CONFIG[statusCode] || ERROR_CONFIG.default;
  const { code, title, titleKey, description, descriptionKey, icon, Icon, isAstronaut, bgClass, showStars } = config;

  // Use custom title/description if provided
  const displayTitle = customTitle || t(titleKey, title);
  const displayDescription = customDescription || t(descriptionKey, description);

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
            {t('errorPage.actions.home')}
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
              {t('errorPage.actions.refresh')}
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
              {t('errorPage.actions.back')}
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
          {t('errorPage.404.footer')}
        </Box>
      )}
    </Box>
  );
}
