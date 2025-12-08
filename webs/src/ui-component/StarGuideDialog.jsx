import { useState, useEffect } from 'react';

// material-ui
import { useTheme, alpha, keyframes } from '@mui/material/styles';
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import Stack from '@mui/material/Stack';

// icons
import CloseIcon from '@mui/icons-material/Close';
import StarIcon from '@mui/icons-material/Star';
import GitHubIcon from '@mui/icons-material/GitHub';
import BugReportIcon from '@mui/icons-material/BugReport';
import FavoriteIcon from '@mui/icons-material/Favorite';

const STORAGE_KEY = 'sublinkpro_star_guide_shown';

// 动画定义
const float = keyframes`
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
`;

const pulse = keyframes`
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.1);
    opacity: 0.8;
  }
`;

const sparkle = keyframes`
  0%, 100% {
    opacity: 0;
    transform: scale(0) rotate(0deg);
  }
  50% {
    opacity: 1;
    transform: scale(1) rotate(180deg);
  }
`;

export default function StarGuideDialog() {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [open, setOpen] = useState(false);
  const [dontShowAgain, setDontShowAgain] = useState(false);

  useEffect(() => {
    // 检查是否已经显示过
    const hasShown = localStorage.getItem(STORAGE_KEY);
    if (!hasShown) {
      // 延迟显示弹窗，让用户先看到页面
      const timer = setTimeout(() => {
        setOpen(true);
      }, 1500);
      return () => clearTimeout(timer);
    }
  }, []);

  const handleClose = () => {
    setOpen(false);
    if (dontShowAgain) {
      localStorage.setItem(STORAGE_KEY, 'true');
    }
  };

  const handleStar = () => {
    window.open('https://github.com/ZeroDeng01/sublinkPro', '_blank');
    localStorage.setItem(STORAGE_KEY, 'true');
    setOpen(false);
  };

  const handleFeedback = () => {
    window.open('https://github.com/ZeroDeng01/sublinkPro/issues', '_blank');
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          borderRadius: 4,
          overflow: 'hidden',
          background: isDark
            ? `linear-gradient(145deg, ${alpha('#1e1b4b', 0.98)} 0%, ${alpha('#312e81', 0.95)} 100%)`
            : `linear-gradient(145deg, ${alpha('#eef2ff', 0.98)} 0%, ${alpha('#e0e7ff', 0.95)} 100%)`,
          backdropFilter: 'blur(20px)',
          border: `1px solid ${isDark ? alpha('#6366f1', 0.3) : alpha('#6366f1', 0.2)}`
        }
      }}
    >
      {/* 关闭按钮 */}
      <IconButton
        onClick={handleClose}
        sx={{
          position: 'absolute',
          top: 12,
          right: 12,
          color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
          zIndex: 1
        }}
      >
        <CloseIcon />
      </IconButton>

      <DialogContent sx={{ p: 0 }}>
        {/* 头部背景 */}
        <Box
          sx={{
            position: 'relative',
            pt: 5,
            pb: 4,
            px: 4,
            textAlign: 'center',
            background: `linear-gradient(135deg, ${alpha('#6366f1', isDark ? 0.3 : 0.15)} 0%, ${alpha('#a855f7', isDark ? 0.2 : 0.1)} 100%)`,
            overflow: 'hidden'
          }}
        >
          {/* 背景装饰 */}
          <Box
            sx={{
              position: 'absolute',
              top: '50%',
              left: '50%',
              transform: 'translate(-50%, -50%)',
              width: 200,
              height: 200,
              borderRadius: '50%',
              background: `radial-gradient(circle, ${alpha('#6366f1', 0.2)} 0%, transparent 70%)`,
              animation: `${pulse} 3s ease-in-out infinite`
            }}
          />

          {/* 星星图标 */}
          <Box
            sx={{
              position: 'relative',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 100,
              height: 100,
              borderRadius: '50%',
              background: `linear-gradient(145deg, ${alpha('#fbbf24', 0.3)} 0%, ${alpha('#f59e0b', 0.2)} 100%)`,
              border: `2px solid ${alpha('#fbbf24', 0.5)}`,
              animation: `${float} 3s ease-in-out infinite`,
              mb: 2
            }}
          >
            <StarIcon sx={{ fontSize: 56, color: '#fbbf24' }} />

            {/* 闪烁装饰 */}
            {[...Array(4)].map((_, i) => (
              <Box
                key={i}
                sx={{
                  position: 'absolute',
                  width: 8,
                  height: 8,
                  borderRadius: '50%',
                  bgcolor: '#fbbf24',
                  animation: `${sparkle} 2s ease-in-out infinite`,
                  animationDelay: `${i * 0.5}s`,
                  top: i === 0 ? -10 : i === 2 ? 'auto' : '50%',
                  bottom: i === 2 ? -10 : 'auto',
                  left: i === 3 ? -10 : i === 1 ? 'auto' : '50%',
                  right: i === 1 ? -10 : 'auto',
                  transform: i === 0 || i === 2 ? 'translateX(-50%)' : 'translateY(-50%)'
                }}
              />
            ))}
          </Box>

          <Typography
            variant="h3"
            sx={{
              fontWeight: 700,
              background: isDark ? 'linear-gradient(135deg, #fff 0%, #c7d2fe 100%)' : 'linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%)',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              position: 'relative'
            }}
          >
            感谢使用 SublinkPro！
          </Typography>
        </Box>

        {/* 内容区 */}
        <Box sx={{ px: 4, py: 3 }}>
          <Typography
            variant="body1"
            sx={{
              textAlign: 'center',
              color: isDark ? alpha('#fff', 0.85) : theme.palette.text.primary,
              mb: 3,
              lineHeight: 1.8
            }}
          >
            如果您觉得这个项目对您有帮助，欢迎在 GitHub 上给我们一个
            <Box
              component="span"
              sx={{
                display: 'inline-flex',
                alignItems: 'center',
                mx: 0.5,
                color: '#fbbf24',
                fontWeight: 600
              }}
            >
              <StarIcon sx={{ fontSize: 18, mr: 0.25 }} />
              Star
            </Box>
            ！您的支持是我们持续改进的动力
            <FavoriteIcon sx={{ fontSize: 16, color: '#ef4444', ml: 0.5, verticalAlign: 'middle' }} />
          </Typography>

          {/* 按钮区 */}
          <Stack direction="row" spacing={2} justifyContent="center" sx={{ mb: 3 }}>
            <Button
              variant="contained"
              size="large"
              startIcon={<GitHubIcon />}
              endIcon={<StarIcon />}
              onClick={handleStar}
              sx={{
                px: 4,
                py: 1.5,
                borderRadius: 3,
                background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                fontWeight: 600,
                fontSize: '1rem',
                textTransform: 'none',
                boxShadow: `0 8px 24px ${alpha('#6366f1', 0.4)}`,
                '&:hover': {
                  background: 'linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%)',
                  boxShadow: `0 12px 32px ${alpha('#6366f1', 0.5)}`,
                  transform: 'translateY(-2px)'
                },
                transition: 'all 0.3s ease'
              }}
            >
              前往 Star
            </Button>
          </Stack>

          {/* 反馈渠道 */}
          <Box
            sx={{
              p: 2,
              borderRadius: 3,
              bgcolor: isDark ? alpha('#fff', 0.05) : alpha('#6366f1', 0.05),
              border: `1px solid ${isDark ? alpha('#fff', 0.1) : alpha('#6366f1', 0.1)}`
            }}
          >
            <Typography
              variant="subtitle2"
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 1,
                color: isDark ? alpha('#fff', 0.9) : theme.palette.text.primary,
                fontWeight: 600,
                mb: 1
              }}
            >
              <BugReportIcon sx={{ fontSize: 20, color: theme.palette.primary.main }} />
              问题反馈
            </Typography>
            <Typography
              variant="body2"
              sx={{
                color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary,
                mb: 1.5
              }}
            >
              如果您在使用过程中遇到任何问题或有功能建议，欢迎通过以下方式反馈：
            </Typography>
            <Button
              variant="outlined"
              size="small"
              startIcon={<GitHubIcon />}
              onClick={handleFeedback}
              sx={{
                borderRadius: 2,
                textTransform: 'none',
                borderColor: isDark ? alpha('#fff', 0.2) : alpha('#6366f1', 0.3),
                color: isDark ? '#a5b4fc' : '#6366f1',
                '&:hover': {
                  borderColor: '#6366f1',
                  bgcolor: alpha('#6366f1', 0.1)
                }
              }}
            >
              GitHub Issues
            </Button>
          </Box>

          {/* 不再显示选项 */}
          <Box sx={{ mt: 2, textAlign: 'center' }}>
            <FormControlLabel
              control={
                <Checkbox
                  checked={dontShowAgain}
                  onChange={(e) => setDontShowAgain(e.target.checked)}
                  size="small"
                  sx={{
                    color: isDark ? alpha('#fff', 0.5) : theme.palette.text.secondary,
                    '&.Mui-checked': {
                      color: theme.palette.primary.main
                    }
                  }}
                />
              }
              label={
                <Typography variant="caption" sx={{ color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary }}>
                  下次不再显示
                </Typography>
              }
            />
          </Box>
        </Box>
      </DialogContent>
    </Dialog>
  );
}
