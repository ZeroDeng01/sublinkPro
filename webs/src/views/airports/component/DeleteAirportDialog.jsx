import PropTypes from 'prop-types';

import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import { useTranslation } from 'react-i18next';

/**
 * 删除机场确认对话框
 */
export default function DeleteAirportDialog({ open, airport, withNodes, setWithNodes, onClose, onConfirm }) {
  const { t } = useTranslation();

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm">
      <DialogTitle>{t('airports.deleteDialog.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText>{t('airports.deleteDialog.confirm', { name: airport?.name || t('common.unknown') })}</DialogContentText>
        <FormControlLabel
          control={<Switch checked={withNodes} onChange={(e) => setWithNodes(e.target.checked)} />}
          label={t('airports.deleteDialog.withNodes')}
          sx={{ mt: 2 }}
        />
        {withNodes && <DialogContentText sx={{ mt: 1, color: 'error.main' }}>{t('airports.deleteDialog.warning')}</DialogContentText>}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" color="error" onClick={onConfirm}>
          {t('airports.deleteDialog.confirmAction')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

DeleteAirportDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  airport: PropTypes.shape({
    id: PropTypes.number,
    name: PropTypes.string
  }),
  withNodes: PropTypes.bool.isRequired,
  setWithNodes: PropTypes.func.isRequired,
  onClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired
};
