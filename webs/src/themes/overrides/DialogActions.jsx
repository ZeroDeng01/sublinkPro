// ==============================|| OVERRIDES - DIALOG ACTIONS ||============================== //

export default function DialogActions() {
  return {
    MuiDialogActions: {
      styleOverrides: {
        root: {
          padding: '16px 24px 24px',
          '& > :not(:first-of-type)': {
            marginLeft: '16px'
          }
        }
      }
    }
  };
}
