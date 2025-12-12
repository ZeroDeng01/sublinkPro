import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Grid';

// icons
import PlayArrowIcon from '@mui/icons-material/PlayArrow';

// constants
import { CRON_OPTIONS, SPEED_TEST_TCP_OPTIONS, SPEED_TEST_MIHOMO_OPTIONS, LATENCY_TEST_URL_OPTIONS } from '../utils';

/**
 * 测速设置对话框
 */
export default function SpeedTestDialog({
  open,
  speedTestForm,
  setSpeedTestForm,
  groupOptions,
  tagOptions,
  onClose,
  onSubmit,
  onRunSpeedTest,
  onModeChange
}) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>测速设置</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <FormControlLabel
            control={
              <Switch checked={speedTestForm.enabled} onChange={(e) => setSpeedTestForm({ ...speedTestForm, enabled: e.target.checked })} />
            }
            label="启用自动测速"
          />
          <Autocomplete
            freeSolo
            options={CRON_OPTIONS}
            getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
            value={speedTestForm.cron}
            onChange={(e, newValue) => {
              const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
              setSpeedTestForm({ ...speedTestForm, cron: value });
            }}
            onInputChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, cron: newValue || '' })}
            renderOption={(props, option) => (
              <Box component="li" {...props} key={option.value}>
                <Box>
                  <Typography variant="body2">{option.label}</Typography>
                  <Typography variant="caption" color="textSecondary">
                    {option.value}
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Cron表达式"
                placeholder="分 时 日 月 周"
                helperText="格式: 分 时 日 月 周 (例如: 0 */1 * * * 表示每小时执行一次)"
              />
            )}
          />
          <FormControl fullWidth>
            <InputLabel>测速模式</InputLabel>
            <Select variant={'outlined'} value={speedTestForm.mode} label="测速模式" onChange={(e) => onModeChange(e.target.value)}>
              <MenuItem value="tcp">Mihomo - 仅延迟测试 (更快)</MenuItem>
              <MenuItem value="mihomo">Mihomo - 真速度测试 (延迟+下载速度)</MenuItem>
            </Select>
          </FormControl>
          <Box>
            <Autocomplete
              freeSolo
              options={speedTestForm.mode === 'mihomo' ? SPEED_TEST_MIHOMO_OPTIONS : SPEED_TEST_TCP_OPTIONS}
              getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
              value={speedTestForm.url}
              onChange={(e, newValue) => {
                const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
                setSpeedTestForm({ ...speedTestForm, url: value });
              }}
              onInputChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, url: newValue || '' })}
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.value}>
                  <Box>
                    <Typography variant="body2">{option.label}</Typography>
                    <Typography variant="caption" color="textSecondary" sx={{ wordBreak: 'break-all' }}>
                      {option.value}
                    </Typography>
                  </Box>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="测速URL"
                  placeholder={speedTestForm.mode === 'mihomo' ? '请选择或输入下载测速URL' : '请选择或输入204测速URL'}
                />
              )}
            />
            <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
              可以自定义测速URL。
              {speedTestForm.mode === 'mihomo'
                ? '真速度测试使用可下载资源地址，例如: https://speed.cloudflare.com/__down?bytes=10000000'
                : '延迟测试使用更轻量的204测试地址，例如: http://cp.cloudflare.com/generate_204'}
            </Typography>
          </Box>
          {/* 延迟测试URL - 仅在Mihomo模式显示 */}
          {speedTestForm.mode === 'mihomo' && (
            <Box>
              <Autocomplete
                freeSolo
                options={LATENCY_TEST_URL_OPTIONS}
                getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
                value={speedTestForm.latency_url || ''}
                onChange={(e, newValue) => {
                  const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
                  setSpeedTestForm({ ...speedTestForm, latency_url: value });
                }}
                onInputChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, latency_url: newValue || '' })}
                renderOption={(props, option) => (
                  <Box component="li" {...props} key={option.value}>
                    <Box>
                      <Typography variant="body2">{option.label}</Typography>
                      <Typography variant="caption" color="textSecondary" sx={{ wordBreak: 'break-all' }}>
                        {option.value}
                      </Typography>
                    </Box>
                  </Box>
                )}
                renderInput={(params) => (
                  <TextField {...params} label="延迟测试URL" placeholder="用于延迟测试的轻量级URL（留空使用速度测试URL）" />
                )}
              />
              <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                延迟测试阶段使用。推荐使用204轻量地址以获得更准确的延迟测量。留空则使用速度测试URL。
              </Typography>
            </Box>
          )}
          <TextField
            fullWidth
            label="超时时间"
            type="number"
            value={speedTestForm.timeout}
            onChange={(e) => setSpeedTestForm({ ...speedTestForm, timeout: Number(e.target.value) })}
            InputProps={{ endAdornment: <InputAdornment position="end">秒</InputAdornment> }}
          />

          {/* 并发与采样配置 */}
          <Typography variant="subtitle2" color="textSecondary" sx={{ mt: 1 }}>
            并发与采样配置
          </Typography>
          <Grid container spacing={2}>
            <Grid item size={{ xs: 12, sm: 4 }}>
              <TextField
                fullWidth
                label="延迟测试并发"
                type="number"
                value={speedTestForm.latency_concurrency || ''}
                placeholder="自动"
                onChange={(e) =>
                  setSpeedTestForm({ ...speedTestForm, latency_concurrency: e.target.value === '' ? 0 : Number(e.target.value) })
                }
                inputProps={{ min: 0, max: 1000 }}
                helperText="0=自动"
              />
            </Grid>
            <Grid item size={{ xs: 12, sm: 4 }}>
              <TextField
                fullWidth
                label="速度测试并发"
                type="number"
                value={speedTestForm.speed_concurrency || 1}
                onChange={(e) => setSpeedTestForm({ ...speedTestForm, speed_concurrency: Math.max(1, Number(e.target.value) || 1) })}
                inputProps={{ min: 1, max: 128 }}
                helperText="建议1-3"
              />
            </Grid>
            <Grid item size={{ xs: 12, sm: 4 }}>
              <TextField
                fullWidth
                label="延迟采样次数"
                type="number"
                value={speedTestForm.latency_samples || 3}
                onChange={(e) => setSpeedTestForm({ ...speedTestForm, latency_samples: Math.max(1, Number(e.target.value) || 3) })}
                inputProps={{ min: 1, max: 10 }}
                helperText="建议3次"
              />
            </Grid>
          </Grid>
          <Typography variant="caption" color="textSecondary" sx={{ mt: -1 }}>
            测速分两阶段：先并发测延迟（多次采样取平均），再低并发测速度。速度并发建议设为1以获得准确结果。
          </Typography>

          <Autocomplete
            multiple
            freeSolo
            options={groupOptions}
            value={speedTestForm.groups || []}
            onChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, groups: newValue })}
            renderInput={(params) => <TextField {...params} label="测速分组" placeholder="留空则测试全部分组" />}
          />
          <Autocomplete
            multiple
            options={tagOptions || []}
            getOptionLabel={(option) => option.name || option}
            value={speedTestForm.tags || []}
            onChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, tags: newValue.map((t) => t.name || t) })}
            isOptionEqualToValue={(option, value) => (option.name || option) === value}
            renderTags={(value, getTagProps) =>
              value.map((option, index) => {
                const tagObj = (tagOptions || []).find((t) => t.name === option);
                const { key, ...tagProps } = getTagProps({ index });
                return (
                  <Chip
                    key={key}
                    label={option}
                    size="small"
                    sx={{
                      backgroundColor: tagObj?.color || '#1976d2',
                      color: '#fff',
                      '& .MuiChip-deleteIcon': { color: 'rgba(255,255,255,0.7)' }
                    }}
                    {...tagProps}
                  />
                );
              })
            }
            renderOption={(props, option) => (
              <Box component="li" {...props} key={option.name}>
                <Box
                  sx={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    backgroundColor: option.color || '#1976d2',
                    mr: 1
                  }}
                />
                {option.name}
              </Box>
            )}
            renderInput={(params) => <TextField {...params} label="测速标签" placeholder="留空则不按标签过滤" />}
          />
          <Typography variant="caption" color="textSecondary" sx={{ mt: -1 }}>
            分组优先级高于标签：选了分组则先按分组筛选，再按标签过滤；只选标签则直接按标签筛选；都不选则测全部。
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={speedTestForm.detect_country}
                onChange={(e) => setSpeedTestForm({ ...speedTestForm, detect_country: e.target.checked })}
              />
            }
            label="检测落地IP国家"
          />
          <Typography variant="caption" color="textSecondary" sx={{ mt: -1 }}>
            开启后，测速时会通过代理获取落地IP并解析对应的国家代码，会降低测速效率。IP通过https://api.ip.sb/ip获取。
          </Typography>
          <Button variant="outlined" startIcon={<PlayArrowIcon />} onClick={onRunSpeedTest}>
            立即测速
          </Button>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={onSubmit}>
          保存
        </Button>
      </DialogActions>
    </Dialog>
  );
}

SpeedTestDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  speedTestForm: PropTypes.shape({
    cron: PropTypes.string,
    enabled: PropTypes.bool,
    mode: PropTypes.string,
    url: PropTypes.string,
    latency_url: PropTypes.string,
    timeout: PropTypes.number,
    groups: PropTypes.array,
    tags: PropTypes.array,
    detect_country: PropTypes.bool,
    latency_concurrency: PropTypes.number,
    speed_concurrency: PropTypes.number,
    latency_samples: PropTypes.number
  }).isRequired,
  setSpeedTestForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  tagOptions: PropTypes.array,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onRunSpeedTest: PropTypes.func.isRequired,
  onModeChange: PropTypes.func.isRequired
};
