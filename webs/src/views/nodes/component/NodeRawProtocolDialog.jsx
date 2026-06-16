import React from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Slide from '@mui/material/Slide';

// icons
import CloseIcon from '@mui/icons-material/Close';

// project imports
import NodeRawInfoEditor from './NodeRawInfoEditor';
import { resolveProtocolPresentationFromLink } from 'utils/protocolPresentation';
import { getNodeColorChipSx, getNodeThemeTokens } from '../nodeTheme';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// Slide transition for mobile
const SlideTransition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

/**
 * 节点原始协议参数独立编辑对话框
 */
export default function NodeRawProtocolDialog({ open, node, protocolMeta, onClose, onUpdate, showMessage }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);

  if (!node) return null;

  // 获取协议信息
  const presentation = resolveProtocolPresentationFromLink(node.Link, protocolMeta);
  const protocolColor = presentation.color || theme.palette.warning.main;

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      fullScreen={isMobile}
      TransitionComponent={isMobile ? SlideTransition : undefined}
      PaperProps={{
        sx: {
          borderRadius: isMobile ? 0 : 2,
          maxHeight: isMobile ? '100%' : 'calc(100vh - 64px)'
        }
      }}
    >
      <DialogTitle
        sx={{
          pb: 2,
          pt: isMobile ? 2 : 3,
          px: 3,
          borderBottom: '1px solid',
          borderColor: 'divider'
        }}
      >
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Stack direction="row" alignItems="center" spacing={1.5}>
            <Typography variant="h5" fontWeight={700}>
              {t('nodes.rawProtocolDialog.title')}
            </Typography>
            <Chip label={presentation.label} size="small" sx={getNodeColorChipSx(theme, tokens, protocolColor)} />
          </Stack>
          <IconButton
            onClick={onClose}
            size="small"
            aria-label={t('common.close')}
            sx={{
              color: 'text.secondary',
              '&:hover': {
                bgcolor: 'action.hover'
              }
            }}
          >
            <CloseIcon />
          </IconButton>
        </Stack>
        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 0.5 }}>
          {t('nodes.rawProtocolDialog.subtitle')}: {node.EffectiveName || node.Name || node.LinkName}
        </Typography>
      </DialogTitle>
      <DialogContent sx={{ p: 3, pt: 4 }}>
        <NodeRawInfoEditor
          node={node}
          protocolMeta={protocolMeta}
          onUpdate={() => {
            onUpdate?.();
            onClose();
          }}
          showMessage={showMessage}
        />
      </DialogContent>
    </Dialog>
  );
}

NodeRawProtocolDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  node: PropTypes.object,
  protocolMeta: PropTypes.array,
  onClose: PropTypes.func.isRequired,
  onUpdate: PropTypes.func,
  showMessage: PropTypes.func
};
