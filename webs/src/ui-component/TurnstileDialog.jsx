import { useState, useRef, useEffect, forwardRef, useImperativeHandle } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

// MUI
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import Fade from '@mui/material/Fade';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// Icons
import CloseIcon from '@mui/icons-material/Close';
import RefreshIcon from '@mui/icons-material/Refresh';
import SecurityIcon from '@mui/icons-material/Security';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';

// 组件
import TurnstileWidget from './TurnstileWidget';

/**
 * Turnstile 验证弹窗组件
 * 用于在用户点击登录后弹出验证面板
 */
const TurnstileDialog = forwardRef(function TurnstileDialog({ open, onClose, onSuccess, siteKey, title }, ref) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const turnstileRef = useRef(null);
  const { t } = useTranslation();

  const dialogTitle = title || t('turnstile.title');

  // 状态：idle | loading | verified | error
  const [status, setStatus] = useState('idle');
  const [errorMessage, setErrorMessage] = useState('');
  const [retryCount, setRetryCount] = useState(0);

  // 暴露 reset 方法给父组件
  useImperativeHandle(ref, () => ({
    reset: () => {
      setStatus('idle');
      setErrorMessage('');
      if (turnstileRef.current?.reset) {
        turnstileRef.current.reset();
      }
    }
  }));

  // 弹窗打开时重置状态
  useEffect(() => {
    if (open) {
      setStatus('loading');
      setErrorMessage('');
    }
  }, [open]);

  // 验证成功回调
  const handleVerify = (token) => {
    setStatus('verified');
    // 短暂延迟显示成功状态后触发回调
    setTimeout(() => {
      onSuccess?.(token);
    }, 500);
  };

  // 验证失败回调
  const handleError = () => {
    setStatus('error');
    setErrorMessage(t('turnstile.error'));
  };

  // Token 过期回调
  const handleExpire = () => {
    setStatus('error');
    setErrorMessage(t('turnstile.expired'));
  };

  // 重试验证
  const handleRetry = () => {
    if (retryCount >= 3) {
      setErrorMessage(t('turnstile.retryLimit'));
      return;
    }
    setRetryCount((prev) => prev + 1);
    setStatus('loading');
    setErrorMessage('');
    if (turnstileRef.current?.reset) {
      turnstileRef.current.reset();
    }
  };

  // 关闭弹窗
  const handleClose = () => {
    setStatus('idle');
    setErrorMessage('');
    setRetryCount(0);
    onClose?.();
  };

  // 渲染内容
  const renderContent = () => {
    switch (status) {
      case 'verified':
        return (
          <Fade in>
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                py: 3,
                gap: 1.5
              }}
            >
              <CheckCircleIcon sx={{ fontSize: 48, color: 'success.main' }} />
              <Typography variant="body1" color="success.main" fontWeight={500}>
                {t('turnstile.success')}
              </Typography>
            </Box>
          </Fade>
        );

      case 'error':
        return (
          <Fade in>
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                py: 3,
                gap: 2
              }}
            >
              <ErrorOutlineIcon sx={{ fontSize: 48, color: 'error.main' }} />
              <Typography variant="body2" color="error.main" textAlign="center">
                {errorMessage}
              </Typography>
              {retryCount < 3 && (
                <Button variant="outlined" color="primary" size="small" startIcon={<RefreshIcon />} onClick={handleRetry}>
                  {t('turnstile.retry')}
                </Button>
              )}
            </Box>
          </Fade>
        );

      case 'loading':
      case 'idle':
      default:
        return (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              py: 2,
              gap: 2,
              minHeight: 120
            }}
          >
            {/* Turnstile 控件容器 */}
            <Box
              sx={{
                display: 'flex',
                justifyContent: 'center',
                minHeight: 65,
                '& > div': {
                  display: 'flex',
                  justifyContent: 'center'
                }
              }}
            >
              <TurnstileWidget
                ref={turnstileRef}
                siteKey={siteKey}
                onVerify={handleVerify}
                onError={handleError}
                onExpire={handleExpire}
                onLoad={() => setStatus('idle')}
                showSuccessMessage={false}
              />
            </Box>
          </Box>
        );
    }
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      fullScreen={isMobile}
      maxWidth="xs"
      fullWidth
      PaperProps={{
        sx: {
          borderRadius: isMobile ? 0 : 3,
          overflow: 'hidden'
        }
      }}
      TransitionComponent={Fade}
      transitionDuration={200}
    >
      {/* 标题栏 */}
      <DialogTitle
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          py: 1.5,
          px: 2,
          background: (theme) =>
            isDark
              ? 'linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(139, 92, 246, 0.1))'
              : 'linear-gradient(135deg, rgba(99, 102, 241, 0.08), rgba(139, 92, 246, 0.05))',
          borderBottom: '1px solid',
          borderColor: 'divider'
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <SecurityIcon sx={{ color: 'primary.main', fontSize: 20 }} />
          <Typography variant="subtitle1" fontWeight={600}>
            {dialogTitle}
          </Typography>
        </Box>
        <IconButton size="small" onClick={handleClose} sx={{ color: 'text.secondary' }}>
          <CloseIcon fontSize="small" />
        </IconButton>
      </DialogTitle>

      {/* 内容区 */}
      <DialogContent
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          py: 3,
          px: 2,
          minHeight: isMobile ? 'auto' : 180
        }}
      >
        {renderContent()}
      </DialogContent>
    </Dialog>
  );
});

TurnstileDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func,
  onSuccess: PropTypes.func,
  siteKey: PropTypes.string.isRequired,
  title: PropTypes.string
};

export default TurnstileDialog;
