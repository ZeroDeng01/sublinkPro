import { QRCodeSVG } from 'qrcode.react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';

/**
 * 二维码展示对话框
 */
export default function QrCodeDialog({ open, title, url, onClose, onCopy }) {
  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent sx={{ textAlign: 'center', pt: 2 }}>
        <QRCodeSVG value={url} size={200} />
        <TextField fullWidth value={url} sx={{ mt: 2 }} size="small" InputProps={{ readOnly: true }} />
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onCopy(url)}>复制</Button>
        <Button onClick={() => window.open(url)}>打开</Button>
        <Button onClick={onClose}>关闭</Button>
      </DialogActions>
    </Dialog>
  );
}
