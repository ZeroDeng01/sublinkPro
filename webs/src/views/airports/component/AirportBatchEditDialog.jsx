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

  const summaryItems = [];

  if (batchForm.applyGroup) {
    summaryItems.push(`分组：${batchForm.group.trim() ? batchForm.group.trim() : '清空分组'}`);
  }
  if (batchForm.applySchedule) {
    summaryItems.push(`调度：${batchForm.cronExpr.trim() || '未设置'}`);
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
        批量设置机场
      </DialogTitle>
      <DialogContent dividers sx={{ pt: 2.5, pb: 2, bgcolor: 'transparent', borderColor: panelBorder }}>
        <Stack spacing={2.5}>
          <Alert severity={summaryItems.length > 0 ? 'info' : 'warning'}>
            {summaryItems.length > 0
              ? `将更新 ${selectedCount} 个机场：${summaryItems.join('；')}`
              : `已选择 ${selectedCount} 个机场，请先勾选本次要修改的字段`}
          </Alert>

          <AirportDialogSection title="节点分组" surface={nestedPanelSurface} borderColor={panelBorder} titleColor={primaryText}>
            <FormControlLabel
              control={
                <Checkbox checked={batchForm.applyGroup} onChange={(e) => setBatchForm({ ...batchForm, applyGroup: e.target.checked })} />
              }
              label="统一设置节点分组"
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
                  renderInput={(params) => <TextField {...params} label="节点分组" placeholder="输入或选择分组，留空表示清空分组" />}
                />
                <Typography variant="caption" sx={{ color: secondaryText }}>
                  会同步更新这些机场已导入节点的分组，留空表示清空分组。
                </Typography>
              </Stack>
            </Collapse>
          </AirportDialogSection>

          <AirportDialogSection title="定时更新" surface={nestedPanelSurface} borderColor={panelBorder} titleColor={primaryText}>
            <Stack spacing={2}>
              <Box>
                <Box sx={controlRowSx}>
                  <Box sx={{ pr: 2 }}>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      更新 Cron 表达式
                    </Typography>
                    <Typography variant="caption" sx={{ color: secondaryText }}>
                      批量编辑仅会更新定时规则，不会改变已选机场当前的启用或禁用状态。
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
                      helperText="可批量修改或清空 Cron 表达式；此处不会同步切换机场的定时更新开关状态。"
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
          取消
        </Button>
        <Button variant="contained" onClick={onSubmit} disabled={submitting}>
          {submitting ? '保存中...' : '确认批量更新'}
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
