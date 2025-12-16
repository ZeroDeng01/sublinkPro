import { useState, useEffect, useCallback } from 'react';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';

// icons
import CloseIcon from '@mui/icons-material/Close';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import SettingsIcon from '@mui/icons-material/Settings';

// project imports
import { getGeoIPStatus } from 'api/geoip';
import GeoIPSettingsDialog from 'views/settings/components/GeoIPSettingsDialog';

const STORAGE_KEY = 'sublinkpro_geoip_warning_dismissed';

// ==============================|| GeoIP 缺失提示对话框 ||============================== //

export default function GeoIPWarningDialog() {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [open, setOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [dontShowAgain, setDontShowAgain] = useState(false);

  // 检查 GeoIP 状态
  const checkStatus = useCallback(async () => {
    // 检查是否已经忽略
    const dismissed = localStorage.getItem(STORAGE_KEY);
    if (dismissed === 'true') {
      return;
    }

    try {
      const res = await getGeoIPStatus();
      if (res.code === 200 && res.data) {
        // 如果数据库不可用且没有正在下载，显示提示
        if (!res.data.available && !res.data.downloading) {
          // 延迟显示，让用户先看到页面
          setTimeout(() => {
            setOpen(true);
          }, 2000);
        }
      }
    } catch (error) {
      console.error('检查 GeoIP 状态失败:', error);
    }
  }, []);

  useEffect(() => {
    checkStatus();
  }, [checkStatus]);

  const handleClose = () => {
    setOpen(false);
    if (dontShowAgain) {
      localStorage.setItem(STORAGE_KEY, 'true');
    }
  };

  const handleOpenSettings = () => {
    setOpen(false);
    setSettingsOpen(true);
  };

  const handleSettingsClose = () => {
    setSettingsOpen(false);
    // 重新检查状态
    checkStatus();
  };

  return (
    <>
      <Dialog
        open={open}
        onClose={handleClose}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: 3,
            overflow: 'hidden',
            background: isDark
              ? `linear-gradient(145deg, ${alpha('#451a03', 0.98)} 0%, ${alpha('#78350f', 0.95)} 100%)`
              : `linear-gradient(145deg, ${alpha('#fef3c7', 0.98)} 0%, ${alpha('#fde68a', 0.95)} 100%)`,
            border: `1px solid ${isDark ? alpha('#f59e0b', 0.3) : alpha('#f59e0b', 0.4)}`
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
          {/* 头部 */}
          <Box
            sx={{
              pt: 4,
              pb: 3,
              px: 4,
              textAlign: 'center',
              background: `linear-gradient(135deg, ${alpha('#f59e0b', isDark ? 0.3 : 0.2)} 0%, ${alpha('#d97706', isDark ? 0.2 : 0.1)} 100%)`
            }}
          >
            {/* 图标 */}
            <Box
              sx={{
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: 80,
                height: 80,
                borderRadius: '50%',
                background: `linear-gradient(145deg, ${alpha('#f59e0b', 0.3)} 0%, ${alpha('#d97706', 0.2)} 100%)`,
                border: `2px solid ${alpha('#f59e0b', 0.5)}`,
                mb: 2
              }}
            >
              <WarningAmberIcon sx={{ fontSize: 48, color: '#f59e0b' }} />
            </Box>

            <Typography
              variant="h4"
              sx={{
                fontWeight: 700,
                color: isDark ? '#fef3c7' : '#92400e',
                position: 'relative'
              }}
            >
              GeoIP 数据库未安装
            </Typography>
          </Box>

          {/* 内容区 */}
          <Box sx={{ px: 4, py: 3 }}>
            <Typography
              variant="body1"
              sx={{
                textAlign: 'center',
                color: isDark ? alpha('#fff', 0.85) : theme.palette.text.primary,
                mb: 2,
                lineHeight: 1.8
              }}
            >
              系统检测到 GeoIP 数据库尚未安装。以下功能将受到影响：
            </Typography>

            <Box
              sx={{
                p: 2,
                borderRadius: 2,
                bgcolor: isDark ? alpha('#fff', 0.05) : alpha('#f59e0b', 0.1),
                border: `1px solid ${isDark ? alpha('#fff', 0.1) : alpha('#f59e0b', 0.2)}`,
                mb: 3
              }}
            >
              <Typography
                variant="body2"
                component="ul"
                sx={{
                  m: 0,
                  pl: 2,
                  color: isDark ? alpha('#fff', 0.8) : theme.palette.text.secondary,
                  '& li': { mb: 0.5 }
                }}
              >
                <li>IP 地理位置信息查询</li>
                <li>节点落地 IP 国家/地区检测</li>
                <li>登录位置信息显示</li>
              </Typography>
            </Box>

            <Typography
              variant="body2"
              sx={{
                textAlign: 'center',
                color: isDark ? alpha('#fff', 0.7) : theme.palette.text.secondary
              }}
            >
              点击下方按钮进入设置页面下载 GeoIP 数据库。
            </Typography>
          </Box>
        </DialogContent>

        <DialogActions sx={{ px: 4, py: 2, justifyContent: 'space-between' }}>
          <FormControlLabel
            control={
              <Checkbox
                checked={dontShowAgain}
                onChange={(e) => setDontShowAgain(e.target.checked)}
                size="small"
                sx={{
                  color: isDark ? alpha('#fff', 0.5) : theme.palette.text.secondary,
                  '&.Mui-checked': {
                    color: '#f59e0b'
                  }
                }}
              />
            }
            label={
              <Typography variant="caption" sx={{ color: isDark ? alpha('#fff', 0.6) : theme.palette.text.secondary }}>
                不再提示
              </Typography>
            }
          />
          <Button
            variant="contained"
            startIcon={<SettingsIcon />}
            onClick={handleOpenSettings}
            sx={{
              borderRadius: 2,
              background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
              '&:hover': {
                background: 'linear-gradient(135deg, #d97706 0%, #b45309 100%)'
              }
            }}
          >
            前往设置
          </Button>
        </DialogActions>
      </Dialog>

      {/* GeoIP 设置对话框 */}
      <GeoIPSettingsDialog open={settingsOpen} onClose={handleSettingsClose} />
    </>
  );
}
