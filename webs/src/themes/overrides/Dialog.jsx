// ==============================|| OVERRIDES - DIALOG ||============================== //

export default function Dialog() {
  return {
    MuiDialog: {
      styleOverrides: {
        paper: {
          padding: 0,
          borderRadius: '16px',
          // 使用更柔和的阴影效果，避免过于明显的发光感
          boxShadow: '0px 8px 24px rgba(0, 0, 0, 0.12)',
          backgroundImage: 'none'
        },
        // 全屏模式下移除圆角，避免四角出现黑色区域
        paperFullScreen: {
          borderRadius: 0
        }
      }
    },
    MuiBackdrop: {
      styleOverrides: {
        root: {
          backgroundColor: 'rgba(0, 0, 0, 0.45)',
          backdropFilter: 'blur(6px)'
        }
      }
    }
  };
}
