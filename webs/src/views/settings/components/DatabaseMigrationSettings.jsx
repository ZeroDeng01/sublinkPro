import { useEffect, useRef, useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import StorageIcon from '@mui/icons-material/Storage';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

import { importDatabaseMigration } from 'api/settings';
import { useTaskProgress } from 'contexts/TaskProgressContext';

export default function DatabaseMigrationSettings({ showMessage }) {
  const { t } = useTranslation();
  const fileInputRef = useRef(null);
  const { registerOnComplete, unregisterOnComplete } = useTaskProgress();

  const [selectedFile, setSelectedFile] = useState(null);
  const [includeAccessKeys, setIncludeAccessKeys] = useState(true);
  const [includeSubLogs, setIncludeSubLogs] = useState(false);
  const [confirmOverwrite, setConfirmOverwrite] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [lastMigrationResult, setLastMigrationResult] = useState(null);

  useEffect(() => {
    const handleTaskComplete = ({ taskType, status, result }) => {
      if (taskType !== 'db_migration') {
        return;
      }

      if (status === 'completed') {
        const warningCount = result?.warnings?.length || 0;
        setLastMigrationResult(result || null);
        showMessage(
          warningCount > 0
            ? t('settings.databaseMigration.messages.completedWithWarnings', { count: warningCount })
            : t('settings.databaseMigration.messages.completed'),
          warningCount > 0 ? 'warning' : 'success'
        );
        return;
      }

      if (status === 'error') {
        showMessage(t('settings.databaseMigration.messages.taskFailed'), 'error');
      }
    };

    registerOnComplete(handleTaskComplete);
    return () => {
      unregisterOnComplete(handleTaskComplete);
    };
  }, [registerOnComplete, showMessage, t, unregisterOnComplete]);

  const handleChooseFile = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event) => {
    const file = event.target.files?.[0] || null;
    setSelectedFile(file);
  };

  const handleStartImport = async () => {
    if (!selectedFile) {
      showMessage(t('settings.databaseMigration.messages.fileRequired'), 'error');
      return;
    }

    if (!confirmOverwrite) {
      showMessage(t('settings.databaseMigration.messages.confirmRequired'), 'error');
      return;
    }

    setSubmitting(true);
    try {
      const formData = new FormData();
      formData.append('file', selectedFile);
      formData.append('includeAccessKeys', String(includeAccessKeys));
      formData.append('includeSubLogs', String(includeSubLogs));

      setLastMigrationResult(null);
      await importDatabaseMigration(formData);
      showMessage(t('settings.databaseMigration.messages.started'));
    } catch (error) {
      showMessage(error.message || t('settings.databaseMigration.messages.startFailed'), 'error');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Card variant="outlined">
      <CardHeader
        avatar={<StorageIcon color="primary" />}
        title={t('settings.databaseMigration.title')}
        subheader={t('settings.databaseMigration.subheader')}
      />
      <CardContent>
        <Stack spacing={2.5}>
          <Alert
            severity="error"
            icon={<WarningAmberIcon fontSize="inherit" />}
            sx={{
              '& .MuiAlert-icon': {
                color: 'error.main'
              },
              '& .MuiAlert-message': {
                color: 'error.dark',
                fontWeight: 600
              }
            }}
          >
            {t('settings.databaseMigration.alerts.overwrite')}
          </Alert>

          <Alert severity="info">
            <Trans i18nKey="settings.databaseMigration.alerts.recommendBackup" components={{ strong: <strong /> }} />
          </Alert>

          <Alert severity="info">
            <Trans i18nKey="settings.databaseMigration.alerts.exportGuide" components={{ strong: <strong /> }} />
          </Alert>

          {lastMigrationResult?.warnings?.length > 0 && (
            <Alert severity="warning">
              <Stack spacing={1}>
                <Typography variant="body2" fontWeight={600}>
                  {t('settings.databaseMigration.warnings.title', { count: lastMigrationResult.warnings.length })}
                </Typography>
                <Box component="ul" sx={{ m: 0, pl: 2.5 }}>
                  {lastMigrationResult.warnings.map((warning, index) => (
                    <Box component="li" key={`${warning}-${index}`} sx={{ mb: 0.5 }}>
                      <Typography variant="body2">{warning}</Typography>
                    </Box>
                  ))}
                </Box>
                <Typography variant="caption" color="text.secondary">
                  {t('settings.databaseMigration.warnings.helper')}
                </Typography>
              </Stack>
            </Alert>
          )}

          <Box>
            <input ref={fileInputRef} type="file" hidden accept=".zip,.db,.sqlite,.sqlite3" onChange={handleFileChange} />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} alignItems={{ xs: 'stretch', sm: 'center' }}>
              <Button variant="outlined" startIcon={<CloudUploadIcon />} onClick={handleChooseFile} disabled={submitting}>
                {t('settings.databaseMigration.actions.chooseFile')}
              </Button>
              {selectedFile ? (
                <Chip
                  color="primary"
                  variant="outlined"
                  label={`${selectedFile.name} · ${(selectedFile.size / 1024 / 1024).toFixed(2)} MB`}
                  sx={{ maxWidth: '100%' }}
                />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  {t('settings.databaseMigration.fileSupport')}
                </Typography>
              )}
            </Stack>
          </Box>

          <Divider />

          <Stack spacing={1}>
            <FormControlLabel
              control={<Checkbox checked={includeAccessKeys} onChange={(event) => setIncludeAccessKeys(event.target.checked)} />}
              label={t('settings.databaseMigration.options.accessKeys')}
            />
            <Typography variant="caption" color="text.secondary" sx={{ ml: 4 }}>
              {t('settings.databaseMigration.options.accessKeysHelper')}
            </Typography>

            <FormControlLabel
              control={<Checkbox checked={includeSubLogs} onChange={(event) => setIncludeSubLogs(event.target.checked)} />}
              label={t('settings.databaseMigration.options.subLogs')}
            />
            <Typography variant="caption" color="text.secondary" sx={{ ml: 4 }}>
              {t('settings.databaseMigration.options.subLogsHelper')}
            </Typography>

            <FormControlLabel
              control={
                <Checkbox checked={confirmOverwrite} onChange={(event) => setConfirmOverwrite(event.target.checked)} color="error" />
              }
              label={
                <Typography color="error.main" fontWeight={600}>
                  {t('settings.databaseMigration.options.confirmOverwrite')}
                </Typography>
              }
            />
            <Typography variant="caption" color="error.main" sx={{ ml: 4, fontWeight: 600 }}>
              {t('settings.databaseMigration.options.restartNotice')}
            </Typography>
          </Stack>

          <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Button variant="contained" color="error" startIcon={<StorageIcon />} onClick={handleStartImport} disabled={submitting}>
              {submitting ? t('settings.databaseMigration.actions.submitting') : t('settings.databaseMigration.actions.start')}
            </Button>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}
