import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useTranslation } from 'react-i18next';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import { getNodeColorChipSx, getNodeDialogPaperSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

/**
 * 添加/编辑节点对话框
 */
export default function NodeDialog({
  open,
  isEdit,
  nodeForm,
  setNodeForm,
  groupOptions,
  proxyNodeOptions,
  loadingProxyNodes,
  tagOptions,
  onClose,
  onSubmit,
  onFetchProxyNodes
}) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth PaperProps={{ sx: getNodeDialogPaperSx(theme, tokens) }}>
      <DialogTitle
        sx={{
          color: tokens.primaryText,
          bgcolor: tokens.mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: tokens.panelBorder
        }}
      >
        {isEdit ? t('nodes.dialog.editTitle') : t('nodes.dialog.addTitle')}
      </DialogTitle>
      <DialogContent dividers sx={{ bgcolor: 'transparent', borderColor: tokens.panelBorder }}>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            fullWidth
            multiline
            rows={4}
            label={t('nodes.dialog.fields.link')}
            value={nodeForm.link}
            onChange={(e) => setNodeForm({ ...nodeForm, link: e.target.value })}
            placeholder={t('nodes.dialog.fields.linkPlaceholder')}
            sx={fieldControlSx}
          />
          {isEdit && (
            <Stack spacing={1.5}>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                <TextField
                  fullWidth
                  label={t('nodes.dialog.fields.remarkName')}
                  value={nodeForm.name}
                  onChange={(e) => setNodeForm({ ...nodeForm, name: e.target.value })}
                  helperText={t('nodes.dialog.fields.remarkNameHelper')}
                  sx={fieldControlSx}
                />
                <TextField
                  select
                  fullWidth
                  label={t('nodes.dialog.fields.nameMode')}
                  value={nodeForm.nameMode || 'link'}
                  onChange={(e) => setNodeForm({ ...nodeForm, nameMode: e.target.value })}
                  SelectProps={{ native: true }}
                  helperText={t('nodes.dialog.fields.nameModeHelper')}
                  sx={fieldControlSx}
                >
                  <option value="link">{t('nodes.dialog.fields.nameModeLink')}</option>
                  <option value="remark">{t('nodes.dialog.fields.nameModeRemark')}</option>
                </TextField>
              </Stack>
              <Typography variant="caption" sx={{ color: tokens.secondaryText }}>
                {t('nodes.dialog.fields.nameModeNotice')}
              </Typography>
            </Stack>
          )}
          <SearchableNodeSelect
            nodes={proxyNodeOptions}
            loading={loadingProxyNodes}
            value={nodeForm.dialerProxyName}
            onChange={(newValue) => {
              const name = typeof newValue === 'string' ? newValue : newValue?.Name || '';
              setNodeForm({ ...nodeForm, dialerProxyName: name });
            }}
            displayField="Name"
            valueField="Name"
            label={t('nodes.dialog.fields.dialerProxy')}
            placeholder={t('nodes.dialog.fields.dialerProxyPlaceholder')}
            helperText={t('nodes.dialog.fields.dialerProxyHelper')}
            freeSolo={true}
            limit={50}
            onFocus={onFetchProxyNodes}
            sx={fieldControlSx}
          />
          <Autocomplete
            freeSolo
            options={groupOptions}
            value={nodeForm.group}
            onChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
            onInputChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
            sx={fieldControlSx}
            renderInput={(params) => (
              <TextField
                {...params}
                label={t('nodes.dialog.fields.group')}
                placeholder={t('nodes.dialog.fields.groupPlaceholder')}
                sx={fieldControlSx}
              />
            )}
          />
          {/* 标签选择 */}
          <Autocomplete
            multiple
            options={tagOptions || []}
            value={nodeForm.tags || []}
            onChange={(e, newValue) => setNodeForm({ ...nodeForm, tags: newValue })}
            getOptionLabel={(option) => option.name || option}
            isOptionEqualToValue={(option, val) => option.name === (val.name || val)}
            renderOption={(props, option) => {
              const { key, ...otherProps } = props;
              return (
                <li key={key} {...otherProps}>
                  <Box
                    sx={{
                      width: 12,
                      height: 12,
                      borderRadius: '50%',
                      backgroundColor: option.color || tokens.palette.primary.main,
                      mr: 1,
                      flexShrink: 0
                    }}
                  />
                  {option.name}
                </li>
              );
            }}
            renderTags={(val, getTagProps) =>
              val.map((option, index) => {
                const { key, ...tagProps } = getTagProps({ index });
                return (
                  <Chip
                    key={key}
                    label={option.name || option}
                    size="small"
                    sx={getNodeColorChipSx(theme, tokens, option.color || theme.palette.primary.main, { deletable: true })}
                    {...tagProps}
                  />
                );
              })
            }
            sx={fieldControlSx}
            renderInput={(params) => (
              <TextField
                {...params}
                label={t('nodes.dialog.fields.tags')}
                placeholder={t('nodes.dialog.fields.tagsPlaceholder')}
                sx={fieldControlSx}
              />
            )}
          />
        </Stack>
      </DialogContent>
      <DialogActions sx={{ bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
        <Button onClick={onClose}>{t('common.close')}</Button>
        <Button variant="contained" onClick={onSubmit}>
          {t('common.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

NodeDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  isEdit: PropTypes.bool.isRequired,
  nodeForm: PropTypes.shape({
    name: PropTypes.string,
    nameMode: PropTypes.string,
    link: PropTypes.string,
    dialerProxyName: PropTypes.string,
    group: PropTypes.string,
    mergeMode: PropTypes.string,
    tags: PropTypes.array
  }).isRequired,
  setNodeForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  tagOptions: PropTypes.array,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onFetchProxyNodes: PropTypes.func.isRequired
};
