import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useTranslation } from 'react-i18next';

// icons
import DeleteIcon from '@mui/icons-material/Delete';
import ClearIcon from '@mui/icons-material/Clear';
import FileDownloadIcon from '@mui/icons-material/FileDownload';
import { getNodeActionButtonSx, getNodeThemeTokens } from '../nodeTheme';
import { withAlpha } from 'utils/colorUtils';

/**
 * 批量操作栏
 */
export default function BatchActions({
  selectedCount,
  totalCount,
  onSelectAll,
  onClearSelection,
  onDelete,
  onGroup,
  onSource,
  onCountry,
  onDialerProxy,
  onExport,
  onTag,
  onRemoveTag
}) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const hasSelection = selectedCount > 0;

  // 是否全选（当前页全选或所有符合条件节点全选）
  const isAllSelected = hasSelection && selectedCount >= totalCount;
  const isIndeterminate = hasSelection && selectedCount < totalCount;

  return (
    <Stack
      direction="row"
      spacing={1}
      sx={{
        mb: 2,
        p: 1.5,
        borderRadius: 2.5,
        bgcolor: tokens.toolbarSurface,
        border: '1px solid',
        borderColor: hasSelection ? tokens.selectedBorder : tokens.softBorder,
        flexWrap: 'wrap',
        gap: 1,
        alignItems: 'center'
      }}
    >
      {/* 全选复选框 */}
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <Checkbox checked={isAllSelected} indeterminate={isIndeterminate} onChange={(e) => onSelectAll(e)} size="small" sx={{ p: 0.5 }} />
        <Typography
          variant="body2"
          sx={{
            ml: 0.5,
            cursor: 'pointer',
            userSelect: 'none',
            whiteSpace: 'nowrap'
          }}
          onClick={() => onSelectAll({ target: { checked: !isAllSelected } })}
        >
          {t('nodes.batch.selectAll')}
        </Typography>
      </Box>

      {/* 已选择数量显示 */}
      <Chip
        label={t('nodes.batch.selectedCount', { count: selectedCount })}
        size="small"
        sx={{
          fontWeight: 600,
          ...(hasSelection && {
            bgcolor: isDark ? withAlpha(tokens.palette.primary.main, 0.18) : tokens.palette.primary.main,
            color: isDark ? tokens.palette.primary.light : tokens.palette.primary.contrastText,
            border: '1px solid',
            borderColor: withAlpha(tokens.palette.primary.main, isDark ? 0.34 : 0.2)
          })
        }}
        color={hasSelection ? undefined : 'default'}
      />

      {/* 清除选择按钮 */}
      <Button
        size="small"
        color="inherit"
        startIcon={<ClearIcon />}
        onClick={onClearSelection}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.text.secondary)}
      >
        {t('common.cancel')}
      </Button>

      <Button
        size="small"
        color="error"
        startIcon={<DeleteIcon />}
        onClick={onDelete}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.error.main)}
      >
        {t('nodes.batch.delete')}
      </Button>
      <Button
        size="small"
        color="primary"
        variant="outlined"
        startIcon={<FileDownloadIcon />}
        onClick={onExport}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.primary.main)}
      >
        {t('nodes.batch.exportLinks')}
      </Button>
      <Button
        size="small"
        color="primary"
        variant="outlined"
        onClick={onGroup}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.warning.main)}
      >
        {t('nodes.batch.changeGroup')}
      </Button>
      <Button
        size="small"
        color="info"
        variant="outlined"
        onClick={onSource}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.info.main)}
      >
        {t('nodes.batch.changeSource')}
      </Button>
      <Button
        size="small"
        color="secondary"
        variant="outlined"
        onClick={onCountry}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.secondary.main)}
      >
        {t('nodes.batch.changeCountry')}
      </Button>
      <Button
        size="small"
        color="primary"
        variant="outlined"
        onClick={onDialerProxy}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.primary.main)}
      >
        {t('nodes.batch.changeDialerProxy')}
      </Button>
      <Button
        size="small"
        color="secondary"
        variant="outlined"
        onClick={onTag}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.success.main)}
      >
        {t('nodes.batch.setTags')}
      </Button>
      <Button
        size="small"
        color="error"
        variant="outlined"
        onClick={onRemoveTag}
        disabled={!hasSelection}
        sx={getNodeActionButtonSx(theme, tokens, tokens.palette.error.main)}
      >
        {t('nodes.batch.removeTags')}
      </Button>
    </Stack>
  );
}

BatchActions.propTypes = {
  selectedCount: PropTypes.number.isRequired,
  totalCount: PropTypes.number.isRequired,
  onSelectAll: PropTypes.func.isRequired,
  onClearSelection: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onGroup: PropTypes.func.isRequired,
  onSource: PropTypes.func.isRequired,
  onCountry: PropTypes.func.isRequired,
  onDialerProxy: PropTypes.func.isRequired,
  onExport: PropTypes.func.isRequired,
  onTag: PropTypes.func.isRequired,
  onRemoveTag: PropTypes.func.isRequired
};
