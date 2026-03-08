import PropTypes from 'prop-types';

// material-ui
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';

// icons
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import SwapHorizIcon from '@mui/icons-material/SwapHoriz';

/**
 * 节点添加结果汇总弹窗
 * 展示添加成功、重复跳过、失败的统计和详情
 */
export default function NodeAddResultDialog({ open, result, onClose }) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

    if (!result) return null;

    const { added = 0, skipped = [], failed = [] } = result;
    const total = added + skipped.length + failed.length;
    const allSuccess = skipped.length === 0 && failed.length === 0;

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth fullScreen={isMobile}>
            <DialogTitle
                sx={{
                    pb: 1,
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1
                }}
            >
                {allSuccess ? (
                    <CheckCircleOutlineIcon sx={{ color: 'success.main' }} />
                ) : (
                    <ErrorOutlineIcon sx={{ color: skipped.length > 0 ? 'warning.main' : 'error.main' }} />
                )}
                节点添加结果
            </DialogTitle>

            <DialogContent sx={{ pt: 1 }}>
                {/* 统计概览 */}
                <Stack
                    direction="row"
                    spacing={1.5}
                    sx={{
                        mb: 2,
                        p: 1.5,
                        borderRadius: 2,
                        backgroundColor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
                        flexWrap: 'wrap',
                        gap: 1
                    }}
                >
                    <Chip
                        icon={<CheckCircleOutlineIcon />}
                        label={`成功 ${added}`}
                        color="success"
                        variant={added > 0 ? 'filled' : 'outlined'}
                        size={isMobile ? 'small' : 'medium'}
                    />
                    {skipped.length > 0 && (
                        <Chip
                            icon={<SwapHorizIcon />}
                            label={`跳过 ${skipped.length}`}
                            color="warning"
                            variant="filled"
                            size={isMobile ? 'small' : 'medium'}
                        />
                    )}
                    {failed.length > 0 && (
                        <Chip
                            icon={<ErrorOutlineIcon />}
                            label={`失败 ${failed.length}`}
                            color="error"
                            variant="filled"
                            size={isMobile ? 'small' : 'medium'}
                        />
                    )}
                    <Typography variant="body2" color="text.secondary" sx={{ alignSelf: 'center', ml: 'auto' }}>
                        共 {total} 个节点
                    </Typography>
                </Stack>

                {/* 全部成功提示 */}
                {allSuccess && added > 0 && (
                    <Alert severity="success" sx={{ borderRadius: 2 }}>
                        所有 {added} 个节点添加成功！
                    </Alert>
                )}

                {/* 重复跳过详情 */}
                {skipped.length > 0 && (
                    <Box sx={{ mb: 2 }}>
                        <Typography variant="subtitle2" color="warning.main" sx={{ mb: 1, fontWeight: 600 }}>
                            重复跳过的节点
                        </Typography>
                        <List
                            dense
                            sx={{
                                border: `1px solid ${theme.palette.divider}`,
                                borderRadius: 2,
                                overflow: 'hidden',
                                '& .MuiListItem-root:not(:last-child)': {
                                    borderBottom: `1px solid ${theme.palette.divider}`
                                }
                            }}
                        >
                            {skipped.map((item, index) => (
                                <ListItem
                                    key={index}
                                    sx={{
                                        flexDirection: isMobile ? 'column' : 'row',
                                        alignItems: isMobile ? 'flex-start' : 'center',
                                        py: 1.5,
                                        px: 2,
                                        gap: isMobile ? 0.5 : 0
                                    }}
                                >
                                    <ListItemIcon sx={{ minWidth: 32, display: isMobile ? 'none' : 'flex' }}>
                                        <SwapHorizIcon fontSize="small" color="warning" />
                                    </ListItemIcon>
                                    <ListItemText
                                        primary={
                                            <Typography variant="body2" sx={{ fontWeight: 500, wordBreak: 'break-all' }}>
                                                {item.name || '未知节点'}
                                            </Typography>
                                        }
                                        secondary={
                                            <Stack direction="row" spacing={0.5} sx={{ mt: 0.5, flexWrap: 'wrap', gap: 0.5 }}>
                                                <Typography variant="caption" color="text.secondary">
                                                    与已有节点重复：
                                                </Typography>
                                                <Chip label={item.existingName} size="small" variant="outlined" sx={{ height: 20, fontSize: '0.7rem' }} />
                                                <Chip
                                                    label={`来源: ${item.source}`}
                                                    size="small"
                                                    variant="outlined"
                                                    color="info"
                                                    sx={{ height: 20, fontSize: '0.7rem' }}
                                                />
                                                <Chip
                                                    label={`分组: ${item.group}`}
                                                    size="small"
                                                    variant="outlined"
                                                    color="secondary"
                                                    sx={{ height: 20, fontSize: '0.7rem' }}
                                                />
                                            </Stack>
                                        }
                                        sx={{ m: 0 }}
                                    />
                                </ListItem>
                            ))}
                        </List>
                    </Box>
                )}

                {/* 失败详情 */}
                {failed.length > 0 && (
                    <Box>
                        {skipped.length > 0 && <Divider sx={{ mb: 2 }} />}
                        <Typography variant="subtitle2" color="error.main" sx={{ mb: 1, fontWeight: 600 }}>
                            添加失败的节点
                        </Typography>
                        <List
                            dense
                            sx={{
                                border: `1px solid ${theme.palette.divider}`,
                                borderRadius: 2,
                                overflow: 'hidden',
                                '& .MuiListItem-root:not(:last-child)': {
                                    borderBottom: `1px solid ${theme.palette.divider}`
                                }
                            }}
                        >
                            {failed.map((item, index) => (
                                <ListItem key={index} sx={{ py: 1.5, px: 2 }}>
                                    <ListItemIcon sx={{ minWidth: 32 }}>
                                        <ErrorOutlineIcon fontSize="small" color="error" />
                                    </ListItemIcon>
                                    <ListItemText
                                        primary={
                                            <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                                                {item.link}
                                            </Typography>
                                        }
                                        secondary={item.error}
                                        sx={{ m: 0 }}
                                    />
                                </ListItem>
                            ))}
                        </List>
                    </Box>
                )}
            </DialogContent>

            <DialogActions sx={{ px: 3, pb: 2 }}>
                <Button variant="contained" onClick={onClose} fullWidth={isMobile}>
                    确定
                </Button>
            </DialogActions>
        </Dialog>
    );
}

NodeAddResultDialog.propTypes = {
    open: PropTypes.bool.isRequired,
    result: PropTypes.shape({
        added: PropTypes.number,
        skipped: PropTypes.arrayOf(
            PropTypes.shape({
                name: PropTypes.string,
                source: PropTypes.string,
                group: PropTypes.string,
                existingName: PropTypes.string
            })
        ),
        failed: PropTypes.arrayOf(
            PropTypes.shape({
                link: PropTypes.string,
                error: PropTypes.string
            })
        )
    }),
    onClose: PropTypes.func.isRequired
};
