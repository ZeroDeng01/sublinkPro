import PropTypes from 'prop-types';

import { useTheme } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControlLabel from '@mui/material/FormControlLabel';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { useTranslation } from 'react-i18next';

import CronExpressionGenerator from 'components/CronExpressionGenerator';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import AirportDialogSection from './AirportDialogSection';

export default function AirportBatchEditDialog({
  open,
  selectedCount,
  batchForm,
  setBatchForm,
  groupOptions,
  onClose,
  onSubmit,
  submitting
}) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);
  const { t } = useTranslation();

  const summaryItems = [];

  if (batchForm.applyGroup) {
    summaryItems.push({
      label: t('airports.batchEdit.summary.group'),
      value: batchForm.group.trim() ? batchForm.group.trim() : t('airports.batchEdit.summary.clearGroup')
    });
  }
  if (batchForm.applySchedule) {
    summaryItems.push({
      label: t('airports.batchEdit.summary.schedule'),
      value: batchForm.cronExpr.trim() || t('airports.batchEdit.summary.unset')
    });
  }

  const controlRowSx = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 2,
    px: 1.5,
    py: 1.25,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.paper, 0.2) : withAlpha(palette.background.paper, 0.92),
    border: '1px solid',
    borderColor: panelBorder
  };

  return (
    <Dialog
      open={open}
      onClose={submitting ? undefined : onClose}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: {
          borderRadius: 2.5,
          border: '1px solid',
          borderColor: panelBorder,
          bgcolor: dialogSurface,
          backgroundImage: dialogSurfaceGradient
        }
      }}
    >
      <DialogTitle
        sx={{
          pb: 1.5,
          color: primaryText,
          bgcolor: mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: panelBorder
        }}
      >
        {t('airports.batchEdit.title')}
      </DialogTitle>
      <DialogContent dividers sx={{ pt: 2.5, pb: 2, bgcolor: 'transparent', borderColor: panelBorder }}>
        <Stack spacing={2.5}>
          <Alert severity={summaryItems.length > 0 ? 'info' : 'warning'}>
            {summaryItems.length > 0
              ? t('airports.batchEdit.messages.selectedWithSummary', {
                  count: selectedCount,
                  summary: summaryItems.map((item) => `${item.label}${item.value}`).join(t('airports.batchEdit.summarySeparator'))
                })
              : t('airports.batchEdit.messages.selectedOnly', { count: selectedCount })}
          </Alert>

          <AirportDialogSection
            title={t('airports.batchEdit.sections.group.title')}
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <FormControlLabel
              control={
                <Checkbox checked={batchForm.applyGroup} onChange={(e) => setBatchForm({ ...batchForm, applyGroup: e.target.checked })} />
              }
              label={t('airports.batchEdit.sections.group.switch')}
              sx={{ color: primaryText, alignItems: 'flex-start', m: 0 }}
            />
            <Collapse in={batchForm.applyGroup}>
              <Stack spacing={1.5} sx={{ mt: 1.5 }}>
                <Autocomplete
                  freeSolo
                  size="small"
                  options={groupOptions}
                  value={batchForm.group}
                  onChange={(e, newValue) => setBatchForm({ ...batchForm, group: newValue || '' })}
                  onInputChange={(e, newValue) => setBatchForm({ ...batchForm, group: newValue ?? '' })}
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      label={t('airports.batchEdit.sections.group.fieldLabel')}
                      placeholder={t('airports.batchEdit.sections.group.placeholder')}
                    />
                  )}
                />
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  {t('airports.batchEdit.sections.group.helper')}
                </Typography>
              </Stack>
            </Collapse>
          </AirportDialogSection>

          <AirportDialogSection
            title={t('airports.batchEdit.sections.schedule.title')}
            surface={nestedPanelSurface}
            borderColor={panelBorder}
            titleColor={primaryText}
          >
            <Stack spacing={2}>
              <Box>
                <Box sx={controlRowSx}>
                  <Box sx={{ pr: 2 }}>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      {t('airports.batchEdit.sections.schedule.fieldLabel')}
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      {t('airports.batchEdit.sections.schedule.description')}
                    </Typography>
                  </Box>
                  <Switch
                    checked={batchForm.applySchedule}
                    onChange={(e) => setBatchForm({ ...batchForm, applySchedule: e.target.checked })}
                  />
                </Box>
                <Collapse in={batchForm.applySchedule}>
                  <Box sx={{ mt: 1.5 }}>
                    <CronExpressionGenerator
                      value={batchForm.cronExpr}
                      onChange={(value) => setBatchForm({ ...batchForm, cronExpr: value })}
                      label=""
                      helperText={t('airports.batchEdit.sections.schedule.helper')}
                    />
                  </Box>
                </Collapse>
              </Box>
            </Stack>
          </AirportDialogSection>
        </Stack>
      </DialogContent>
      <DialogActions
        sx={{
          px: 3,
          py: 2,
          bgcolor: mutedPanelSurface,
          borderTop: '1px solid',
          borderColor: panelBorder
        }}
      >
        <Button onClick={onClose} disabled={submitting}>
          {t('common.cancel')}
        </Button>
        <Button variant="contained" onClick={onSubmit} disabled={submitting}>
          {submitting ? t('common.saving') : t('airports.batchEdit.actions.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

AirportBatchEditDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  batchForm: PropTypes.shape({
    applyGroup: PropTypes.bool.isRequired,
    group: PropTypes.string.isRequired,
    applySchedule: PropTypes.bool.isRequired,
    cronExpr: PropTypes.string.isRequired
  }).isRequired,
  setBatchForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  submitting: PropTypes.bool
};

AirportBatchEditDialog.defaultProps = {
  submitting: false
};
