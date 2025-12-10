import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
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

// icons
import PlayArrowIcon from '@mui/icons-material/PlayArrow';

// constants
import { CRON_OPTIONS, SPEED_TEST_TCP_OPTIONS, SPEED_TEST_MIHOMO_OPTIONS } from '../utils';

/**
 * 测速设置对话框
 */
export default function SpeedTestDialog({
  open,
  speedTestForm,
  setSpeedTestForm,
  groupOptions,
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
          <TextField
            fullWidth
            label="超时时间"
            type="number"
            value={speedTestForm.timeout}
            onChange={(e) => setSpeedTestForm({ ...speedTestForm, timeout: Number(e.target.value) })}
            InputProps={{ endAdornment: <InputAdornment position="end">秒</InputAdornment> }}
          />
          <Box>
            <TextField
              fullWidth
              label="并发数"
              type="number"
              value={speedTestForm.concurrency || ''}
              placeholder="留空自动设置"
              onChange={(e) => setSpeedTestForm({ ...speedTestForm, concurrency: e.target.value === '' ? 0 : Number(e.target.value) })}
              inputProps={{ min: 0, max: 100 }}
            />
            <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
              设置测速并发数量。留空或设为0时，系统将根据CPU核心数自动设置（2倍核心数，最小2，最大建议不超过核心数的2倍，可以自行按需调整）。
            </Typography>
          </Box>
          <Autocomplete
            multiple
            freeSolo
            options={groupOptions}
            value={speedTestForm.groups || []}
            onChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, groups: newValue })}
            renderInput={(params) => <TextField {...params} label="测速分组" placeholder="留空则测试全部分组" />}
          />
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
    timeout: PropTypes.number,
    groups: PropTypes.array,
    detect_country: PropTypes.bool,
    concurrency: PropTypes.number
  }).isRequired,
  setSpeedTestForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onRunSpeedTest: PropTypes.func.isRequired,
  onModeChange: PropTypes.func.isRequired
};
