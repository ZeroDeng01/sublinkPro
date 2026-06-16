import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import CountryRules from './components/CountryRules';

// ==============================|| COUNTRY RULES MANAGEMENT ||============================== //

export default function CountryRulesManagement() {
  const { t } = useTranslation();
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  const handleSnackbarClose = useCallback(() => {
    setSnackbar((prev) => ({ ...prev, open: false }));
  }, []);

  return (
    <>
      <MainCard title={t('navigation.items.countryRules')}>
        <CountryRules showMessage={showMessage} />
      </MainCard>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleSnackbarClose}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>
    </>
  );
}
