// ==============================|| OVERRIDES - DIALOG ||============================== //

export default function Dialog() {
  return {
    MuiDialog: {
      styleOverrides: {
        paper: {
          padding: 0,
          borderRadius: '16px',
          boxShadow: '0px 24px 48px -12px rgba(0, 0, 0, 0.18)',
          backgroundImage: 'none'
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
