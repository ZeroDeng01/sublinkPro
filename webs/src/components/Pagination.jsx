import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

// icons
import FirstPageIcon from '@mui/icons-material/FirstPage';
import LastPageIcon from '@mui/icons-material/LastPage';
import NavigateBeforeIcon from '@mui/icons-material/NavigateBefore';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';

/**
 * 通用分页组件
 * @param {number} page - 当前页码（0-indexed 兼容 MUI）
 * @param {number} pageSize - 每页条数
 * @param {number} totalItems - 总条数
 * @param {function} onPageChange - 页码变化回调 (event, newPage) => void
 * @param {function} onPageSizeChange - 每页条数变化回调 (event) => void
 * @param {array} pageSizeOptions - 每页条数选项
 * @param {boolean} disabled - 是否禁用
 */
export default function Pagination({
  page = 0,
  pageSize = 20,
  totalItems = 0,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [10, 20, 50, 100],
  disabled = false,
  showPageInput = true,
  showFirstLast = true,
  labelRowsPerPage = '每页',
  ...props
}) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const totalPages = pageSize > 0 ? Math.ceil(totalItems / pageSize) : 0;
  const currentPage = page + 1; // 显示给用户的是1-indexed

  const handleFirstPage = () => {
    if (!disabled && page > 0) {
      onPageChange?.(null, 0);
    }
  };

  const handlePrevPage = () => {
    if (!disabled && page > 0) {
      onPageChange?.(null, page - 1);
    }
  };

  const handleNextPage = () => {
    if (!disabled && page < totalPages - 1) {
      onPageChange?.(null, page + 1);
    }
  };

  const handleLastPage = () => {
    if (!disabled && page < totalPages - 1) {
      onPageChange?.(null, totalPages - 1);
    }
  };

  const handlePageInputChange = (e) => {
    const value = parseInt(e.target.value, 10);
    if (!isNaN(value) && value >= 1 && value <= totalPages) {
      onPageChange?.(null, value - 1);
    }
  };

  const handlePageInputBlur = (e) => {
    const value = parseInt(e.target.value, 10);
    if (isNaN(value) || value < 1) {
      // 重置为第一页
      e.target.value = 1;
    } else if (value > totalPages) {
      // 重置为最后一页
      e.target.value = totalPages || 1;
    }
  };

  const handlePageSizeSelect = (e) => {
    // 只调用 onPageSizeChange，页码重置由调用方负责
    // 避免同时调用 onPageSizeChange 和 onPageChange 导致两次请求
    onPageSizeChange?.(e);
  };

  // 计算显示范围
  const from = totalItems === 0 ? 0 : page * pageSize + 1;
  const to = Math.min((page + 1) * pageSize, totalItems);

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: isMobile ? 'column' : 'row',
        alignItems: 'center',
        justifyContent: 'flex-end',
        gap: isMobile ? 1 : 2,
        py: 1,
        px: 2,
        flexWrap: 'wrap',
        ...props.sx
      }}
    >
      {/* 显示范围 */}
      <Typography variant="body2" color="textSecondary" sx={{ whiteSpace: 'nowrap' }}>
        {totalItems > 0 ? `${from}-${to} / ${totalItems}` : '0 条'}
      </Typography>

      {/* 每页条数选择 */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <Typography variant="body2" color="textSecondary">
          {labelRowsPerPage}
        </Typography>
        <Select
          size="small"
          value={pageSize}
          onChange={handlePageSizeSelect}
          disabled={disabled}
          sx={{ minWidth: 70, '& .MuiSelect-select': { py: 0.5 } }}
        >
          {pageSizeOptions.map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </Select>
      </Box>

      {/* 分页控制 */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
        {showFirstLast && (
          <IconButton size="small" onClick={handleFirstPage} disabled={disabled || page === 0} title="首页">
            <FirstPageIcon fontSize="small" />
          </IconButton>
        )}

        <IconButton size="small" onClick={handlePrevPage} disabled={disabled || page === 0} title="上一页">
          <NavigateBeforeIcon fontSize="small" />
        </IconButton>

        {/* 页码输入 */}
        {showPageInput && totalPages > 0 && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mx: 0.5 }}>
            <TextField
              size="small"
              type="number"
              defaultValue={currentPage}
              key={currentPage}
              onBlur={handlePageInputBlur}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  handlePageInputChange(e);
                  e.target.blur();
                }
              }}
              disabled={disabled}
              inputProps={{
                min: 1,
                max: totalPages,
                style: { textAlign: 'center', width: '40px', padding: '4px 8px' }
              }}
              sx={{
                width: '60px',
                '& .MuiOutlinedInput-root': {
                  '& input': {
                    MozAppearance: 'textfield',
                    '&::-webkit-outer-spin-button, &::-webkit-inner-spin-button': {
                      WebkitAppearance: 'none',
                      margin: 0
                    }
                  }
                }
              }}
            />
            <Typography variant="body2" color="textSecondary">
              / {totalPages}
            </Typography>
          </Box>
        )}

        <IconButton size="small" onClick={handleNextPage} disabled={disabled || page >= totalPages - 1} title="下一页">
          <NavigateNextIcon fontSize="small" />
        </IconButton>

        {showFirstLast && (
          <IconButton size="small" onClick={handleLastPage} disabled={disabled || page >= totalPages - 1} title="末页">
            <LastPageIcon fontSize="small" />
          </IconButton>
        )}
      </Box>
    </Box>
  );
}

Pagination.propTypes = {
  page: PropTypes.number,
  pageSize: PropTypes.number,
  totalItems: PropTypes.number,
  onPageChange: PropTypes.func,
  onPageSizeChange: PropTypes.func,
  pageSizeOptions: PropTypes.arrayOf(PropTypes.number),
  disabled: PropTypes.bool,
  showPageInput: PropTypes.bool,
  showFirstLast: PropTypes.bool,
  labelRowsPerPage: PropTypes.string,
  sx: PropTypes.object
};
