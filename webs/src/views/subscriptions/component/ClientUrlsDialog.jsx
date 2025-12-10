import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';

/**
 * 客户端链接对话框
 */
export default function ClientUrlsDialog({ open, clientUrls, onClose, onQrCode, onCopy }) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>客户端（点击二维码获取地址）</DialogTitle>
      <DialogContent>
        <Stack spacing={2}>
          {Object.entries(clientUrls).map(([name, url]) => (
            <Stack key={name} direction="row" alignItems="center" spacing={2}>
              <Chip label={name} color="success" sx={{ minWidth: 100 }} />
              <Button variant="outlined" onClick={() => onQrCode(name === '自动识别' ? url : `${url}&client=${name}`, name)}>
                二维码
              </Button>
              <IconButton size="small" onClick={() => onCopy(name === '自动识别' ? url : `${url}&client=${name}`)}>
                <ContentCopyIcon fontSize="small" />
              </IconButton>
            </Stack>
          ))}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>关闭</Button>
      </DialogActions>
    </Dialog>
  );
}
