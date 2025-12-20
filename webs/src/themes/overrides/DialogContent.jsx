// ==============================|| OVERRIDES - DIALOG CONTENT ||============================== //

export default function DialogContent() {
  return {
    MuiDialogContent: {
      styleOverrides: {
        root: {
          padding: '0 24px 24px'
        },
        // 移除 dividers 变体的边框样式，避免与内部 Card 组件边框重叠
        dividers: {
          borderTop: 'none',
          borderBottom: 'none'
        }
      }
    }
  };
}
