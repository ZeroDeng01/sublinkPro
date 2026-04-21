import { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import FormControl from '@mui/material/FormControl';
import Grid from '@mui/material/Grid';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { alpha, useColorScheme, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import { withAlpha } from 'utils/colorUtils';

// icons
import ScheduleIcon from '@mui/icons-material/Schedule';
import EditIcon from '@mui/icons-material/Edit';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';

// ==================== Cron 工具函数 ====================

/**
 * 解析 cron 字段，返回匹配的值数组
 * @param {string} field - cron 字段值
 * @param {number} min - 最小值
 * @param {number} max - 最大值
 * @returns {number[]} - 匹配的值数组
 */
const parseCronField = (field, min, max) => {
  const values = [];

  // 处理逗号分隔的多个值
  const parts = field.split(',');

  for (const part of parts) {
    // 处理 */n 格式
    if (part.startsWith('*/')) {
      const step = parseInt(part.slice(2), 10);
      for (let i = min; i <= max; i += step) {
        values.push(i);
      }
    }
    // 处理 n-m/s 格式 (范围+步长)
    else if (part.includes('-') && part.includes('/')) {
      const [range, stepStr] = part.split('/');
      const [start, end] = range.split('-').map((n) => parseInt(n, 10));
      const step = parseInt(stepStr, 10);
      for (let i = start; i <= end; i += step) {
        values.push(i);
      }
    }
    // 处理 n-m 格式 (范围)
    else if (part.includes('-')) {
      const [start, end] = part.split('-').map((n) => parseInt(n, 10));
      for (let i = start; i <= end; i++) {
        values.push(i);
      }
    }
    // 处理 * 格式
    else if (part === '*') {
      for (let i = min; i <= max; i++) {
        values.push(i);
      }
    }
    // 处理单个数字
    else {
      const num = parseInt(part, 10);
      if (!isNaN(num) && num >= min && num <= max) {
        values.push(num);
      }
    }
  }

  return [...new Set(values)].sort((a, b) => a - b);
};

/**
 * 验证 cron 表达式格式
 * @param {string} cron - Cron 表达式
 * @returns {boolean} - 是否有效
 */
export const validateCronExpression = (cron) => {
  if (!cron) return false;
  const parts = cron.trim().split(/\s+/);
  if (parts.length !== 5) return false;

  const ranges = [
    { min: 0, max: 59 }, // 分钟
    { min: 0, max: 23 }, // 小时
    { min: 1, max: 31 }, // 日
    { min: 1, max: 12 }, // 月
    { min: 0, max: 7 } // 星期 (0和7都表示周日)
  ];

  for (let i = 0; i < 5; i++) {
    const part = parts[i];
    const patterns = [
      /^\*$/, // *
      /^\*\/\d+$/, // */n
      /^\d+$/, // n
      /^\d+-\d+$/, // n-m
      /^[\d,]+$/, // n,m,o
      /^\d+-\d+\/\d+$/ // n-m/s
    ];

    if (!patterns.some((p) => p.test(part))) {
      return false;
    }

    const numbers = part.match(/\d+/g);
    if (numbers) {
      for (const num of numbers) {
        const n = parseInt(num, 10);
        if (n < ranges[i].min || n > ranges[i].max) {
          return false;
        }
      }
    }
  }
  return true;
};

/**
 * 检查给定时间是否匹配 cron 表达式
 * @param {Date} date - 要检查的时间
 * @param {object} cronParts - 解析后的 cron 各字段
 * @returns {boolean} - 是否匹配
 */
const matchesCron = (date, cronParts) => {
  const minute = date.getMinutes();
  const hour = date.getHours();
  const dayOfMonth = date.getDate();
  const month = date.getMonth() + 1;
  let dayOfWeek = date.getDay(); // 0 = 周日

  // 检查各字段是否匹配
  if (!cronParts.minutes.includes(minute)) return false;
  if (!cronParts.hours.includes(hour)) return false;
  if (!cronParts.months.includes(month)) return false;

  // 日期和星期的特殊处理：如果两者都不是 *，则只需匹配其一
  const dayOfMonthMatch = cronParts.daysOfMonth.includes(dayOfMonth);
  const dayOfWeekMatch = cronParts.daysOfWeek.includes(dayOfWeek) || cronParts.daysOfWeek.includes(dayOfWeek === 0 ? 7 : dayOfWeek);

  // 如果日期字段全覆盖（等同于 *），只检查星期
  if (cronParts.daysOfMonth.length === 31) {
    if (!dayOfWeekMatch) return false;
  }
  // 如果星期字段全覆盖（等同于 *），只检查日期
  else if (cronParts.daysOfWeek.length >= 7) {
    if (!dayOfMonthMatch) return false;
  }
  // 否则两者满足其一即可
  else {
    if (!dayOfMonthMatch && !dayOfWeekMatch) return false;
  }

  return true;
};

/**
 * 计算 cron 表达式的下 N 次运行时间
 * @param {string} cronExpr - Cron 表达式 (分 时 日 月 周)
 * @param {number} count - 要计算的次数
 * @param {Date} startFrom - 起始时间，默认当前时间
 * @returns {Date[]} - 下次运行时间数组
 */
export const getNextCronRuns = (cronExpr, count = 3, startFrom = new Date()) => {
  if (!validateCronExpression(cronExpr)) {
    return [];
  }

  const parts = cronExpr.trim().split(/\s+/);
  const cronParts = {
    minutes: parseCronField(parts[0], 0, 59),
    hours: parseCronField(parts[1], 0, 23),
    daysOfMonth: parseCronField(parts[2], 1, 31),
    months: parseCronField(parts[3], 1, 12),
    daysOfWeek: parseCronField(parts[4], 0, 7)
  };

  const results = [];
  // 从下一分钟开始检查
  const current = new Date(startFrom);
  current.setSeconds(0);
  current.setMilliseconds(0);
  current.setMinutes(current.getMinutes() + 1);

  // 最多检查两年的分钟数，防止无限循环
  const maxIterations = 2 * 365 * 24 * 60;
  let iterations = 0;

  while (results.length < count && iterations < maxIterations) {
    if (matchesCron(current, cronParts)) {
      results.push(new Date(current));
    }
    current.setMinutes(current.getMinutes() + 1);
    iterations++;
  }

  return results;
};

/**
 * 格式化相对时间
 * @param {Date} date - 目标时间
 * @param {Date} now - 当前时间
 * @returns {string} - 相对时间描述
 */
const formatRelativeTime = (date, now = new Date()) => {
  const diffMs = date.getTime() - now.getTime();
  const diffMinutes = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffDays > 0) {
    const remainHours = diffHours % 24;
    return remainHours > 0 ? `${diffDays}天${remainHours}小时后` : `${diffDays}天后`;
  }
  if (diffHours > 0) {
    const remainMinutes = diffMinutes % 60;
    return remainMinutes > 0 ? `${diffHours}小时${remainMinutes}分钟后` : `${diffHours}小时后`;
  }
  return `${diffMinutes}分钟后`;
};

/**
 * 格式化日期时间
 * @param {Date} date - 日期对象
 * @returns {string} - 格式化后的字符串
 */
const formatDateTime = (date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  return `${year}-${month}-${day} ${hours}:${minutes}`;
};

/**
 * 格式化日期时间（简短版，用于移动端）
 * @param {Date} date - 日期对象
 * @returns {string} - 格式化后的字符串
 */
const formatDateTimeShort = (date) => {
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  return `${month}-${day} ${hours}:${minutes}`;
};

// ==================== 预设选项 ====================

const CRON_PRESETS = [
  { label: '每30分钟', value: '*/30 * * * *', icon: '⏱️' },
  { label: '每1小时', value: '0 * * * *', icon: '🕐' },
  { label: '每6小时', value: '0 */6 * * *', icon: '🕕' },
  { label: '每12小时', value: '0 */12 * * *', icon: '🕛' },
  { label: '每天0点', value: '0 0 * * *', icon: '🌙' },
  { label: '每周一', value: '0 0 * * 1', icon: '📅' }
];

const FREQUENCY_OPTIONS = [
  { value: 'interval', label: '每隔固定时间' },
  { value: 'daily', label: '每天指定时间' },
  { value: 'weekly', label: '每周指定日期' }
];

const INTERVAL_OPTIONS = [
  { value: 5, label: '5分钟' },
  { value: 10, label: '10分钟' },
  { value: 15, label: '15分钟' },
  { value: 30, label: '30分钟' },
  { value: 60, label: '1小时' },
  { value: 120, label: '2小时' },
  { value: 180, label: '3小时' },
  { value: 360, label: '6小时' },
  { value: 720, label: '12小时' }
];

const WEEKDAY_OPTIONS = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 0, label: '周日' }
];

// ==================== 主组件 ====================

/**
 * Cron 表达式生成器组件
 * 提供直观的可视化界面让用户设置定时任务规则
 */
export default function CronExpressionGenerator({ value, onChange, label = 'Cron表达式', helperText, error = false }) {
  const theme = useTheme();
  const { mode } = useColorScheme();
  const palette = theme.vars?.palette || theme.palette;
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const runtimeColorScheme = typeof document !== 'undefined' ? document.documentElement?.getAttribute('data-color-scheme') : null;
  const isDark = (mode || runtimeColorScheme || theme.palette.mode) === 'dark';
  const customPanelBackground = isDark ? withAlpha(palette.background.default, 0.92) : alpha(theme.palette.background.default, 0.5);
  const customPanelBorder = isDark ? withAlpha(palette.divider, 0.84) : alpha(theme.palette.divider, 0.5);
  const nestedFieldBackground = isDark ? withAlpha(palette.background.paper, 0.94) : 'background.paper';
  const nestedFieldBorder = isDark ? withAlpha(palette.divider, 0.82) : undefined;
  const nestedFieldHoverBorder = isDark ? withAlpha(palette.divider, 0.94) : undefined;
  const subtleChipBackground = isDark ? withAlpha(palette.background.paper, 0.92) : undefined;

  // 状态管理
  const [showCustom, setShowCustom] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [frequency, setFrequency] = useState('interval');
  const [interval, setInterval] = useState(60);
  const [hour, setHour] = useState(0);
  const [minute, setMinute] = useState(0);
  const [weekdays, setWeekdays] = useState([1]); // 默认周一

  // 检查当前值是否匹配预设
  const matchedPreset = useMemo(() => {
    return CRON_PRESETS.find((preset) => preset.value === value);
  }, [value]);

  // 计算下次运行时间
  const nextRuns = useMemo(() => {
    if (!value || !validateCronExpression(value)) {
      return [];
    }
    return getNextCronRuns(value, 3);
  }, [value]);

  // 验证表达式是否有效
  const isValid = useMemo(() => {
    return !value || validateCronExpression(value);
  }, [value]);

  // 从自定义配置生成 cron 表达式
  const generateCronFromConfig = () => {
    switch (frequency) {
      case 'interval':
        if (interval < 60) {
          return `*/${interval} * * * *`;
        } else {
          const hours = Math.floor(interval / 60);
          return `0 */${hours} * * *`;
        }
      case 'daily':
        return `${minute} ${hour} * * *`;
      case 'weekly':
        const days = weekdays.length > 0 ? weekdays.join(',') : '1';
        return `${minute} ${hour} * * ${days}`;
      default:
        return '0 * * * *';
    }
  };

  // 当自定义配置变化时自动更新
  useEffect(() => {
    if (showCustom) {
      const newCron = generateCronFromConfig();
      if (newCron !== value) {
        onChange(newCron);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [frequency, interval, hour, minute, weekdays, showCustom, showAdvanced]);

  // 预设按钮点击
  const handlePresetClick = (preset) => {
    onChange(preset.value);
    setShowCustom(false);
    setShowAdvanced(false);
  };

  // 切换自定义模式
  const handleCustomToggle = () => {
    setShowCustom(!showCustom);
    if (!showCustom) {
      // 进入自定义模式时，根据当前值初始化配置
      if (value) {
        // 尝试解析当前值来初始化配置
        const parts = value.trim().split(/\s+/);
        if (parts.length === 5) {
          // 简单解析
          if (parts[0].startsWith('*/')) {
            setFrequency('interval');
            setInterval(parseInt(parts[0].slice(2), 10));
          } else if (parts[1].startsWith('*/')) {
            setFrequency('interval');
            setInterval(parseInt(parts[1].slice(2), 10) * 60);
          } else if (parts[4] !== '*') {
            setFrequency('weekly');
            setHour(parseInt(parts[1], 10) || 0);
            setMinute(parseInt(parts[0], 10) || 0);
            const days = parts[4].split(',').map((d) => parseInt(d, 10));
            setWeekdays(days);
          } else {
            setFrequency('daily');
            setHour(parseInt(parts[1], 10) || 0);
            setMinute(parseInt(parts[0], 10) || 0);
          }
        }
      }
    }
  };

  return (
    <Box>
      {/* 标签 */}
      {label ? (
        <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 1, display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <ScheduleIcon fontSize="small" />
          {label}
        </Typography>
      ) : null}

      {/* 预设快捷选项 */}
      <Box sx={{ mb: 2 }}>
        <Grid container spacing={1}>
          {CRON_PRESETS.map((preset) => (
            <Grid item key={preset.value} size={{ xs: 4, sm: 'auto' }}>
              <Chip
                label={
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <span>{preset.icon}</span>
                    <span>{preset.label}</span>
                  </Box>
                }
                onClick={() => handlePresetClick(preset)}
                variant={matchedPreset?.value === preset.value ? 'filled' : 'outlined'}
                color={matchedPreset?.value === preset.value ? 'primary' : 'default'}
                sx={{
                  width: '100%',
                  height: 36,
                  fontSize: isMobile ? '0.75rem' : '0.875rem',
                  bgcolor:
                    matchedPreset?.value === preset.value ? undefined : isDark ? alpha(theme.palette.background.paper, 0.08) : undefined,
                  '&:hover': {
                    backgroundColor: alpha(theme.palette.primary.main, isDark ? 0.14 : 0.1)
                  }
                }}
              />
            </Grid>
          ))}
          <Grid item size={{ xs: 4, sm: 'auto' }}>
            <Chip
              label={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <EditIcon fontSize="small" />
                  <span>自定义</span>
                </Box>
              }
              onClick={handleCustomToggle}
              variant={showCustom ? 'filled' : 'outlined'}
              color={showCustom ? 'secondary' : 'default'}
              sx={{
                width: '100%',
                height: 36,
                fontSize: isMobile ? '0.75rem' : '0.875rem',
                bgcolor: showCustom ? undefined : isDark ? alpha(theme.palette.background.paper, 0.08) : undefined
              }}
            />
          </Grid>
        </Grid>
      </Box>

      {/* 自定义配置面板 */}
      <Collapse in={showCustom}>
        <Paper
          variant="outlined"
          sx={{
            p: 2,
            mb: 2,
            backgroundColor: customPanelBackground,
            borderColor: customPanelBorder,
            borderRadius: 2,
            boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}` : 'none',
            '& .MuiOutlinedInput-root': {
              backgroundColor: nestedFieldBackground,
              '& fieldset': {
                borderColor: nestedFieldBorder
              },
              '&:hover fieldset': {
                borderColor: nestedFieldHoverBorder
              }
            }
          }}
        >
          <Stack spacing={2}>
            {/* 频率选择 */}
            <FormControl fullWidth size="small">
              <InputLabel>执行频率</InputLabel>
              <Select value={frequency} label="执行频率" onChange={(e) => setFrequency(e.target.value)}>
                {FREQUENCY_OPTIONS.map((opt) => (
                  <MenuItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {/* 间隔时间选择 */}
            {frequency === 'interval' && (
              <FormControl fullWidth size="small">
                <InputLabel>间隔时间</InputLabel>
                <Select value={interval} label="间隔时间" onChange={(e) => setInterval(e.target.value)}>
                  {INTERVAL_OPTIONS.map((opt) => (
                    <MenuItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}

            {/* 时间选择 */}
            {(frequency === 'daily' || frequency === 'weekly') && (
              <Grid container spacing={2}>
                <Grid item size={{ xs: 6 }}>
                  <FormControl fullWidth size="small">
                    <InputLabel>小时</InputLabel>
                    <Select value={hour} label="小时" onChange={(e) => setHour(e.target.value)}>
                      {Array.from({ length: 24 }, (_, i) => (
                        <MenuItem key={i} value={i}>
                          {String(i).padStart(2, '0')}:00
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item size={{ xs: 6 }}>
                  <FormControl fullWidth size="small">
                    <InputLabel>分钟</InputLabel>
                    <Select value={minute} label="分钟" onChange={(e) => setMinute(e.target.value)}>
                      {[0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55].map((m) => (
                        <MenuItem key={m} value={m}>
                          :{String(m).padStart(2, '0')}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
              </Grid>
            )}

            {/* 星期选择 */}
            {frequency === 'weekly' && (
              <Box>
                <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                  选择执行日期（可多选）
                </Typography>
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                  {WEEKDAY_OPTIONS.map((day) => (
                    <Chip
                      key={day.value}
                      label={day.label}
                      size="small"
                      onClick={() => {
                        if (weekdays.includes(day.value)) {
                          setWeekdays(weekdays.filter((d) => d !== day.value));
                        } else {
                          setWeekdays([...weekdays, day.value].sort((a, b) => a - b));
                        }
                      }}
                      variant={weekdays.includes(day.value) ? 'filled' : 'outlined'}
                      color={weekdays.includes(day.value) ? 'primary' : 'default'}
                      sx={{ minWidth: 48 }}
                    />
                  ))}
                </Box>
              </Box>
            )}

            {/* 高级模式切换 */}
            <Button
              size="small"
              onClick={() => setShowAdvanced(!showAdvanced)}
              endIcon={showAdvanced ? <ExpandLessIcon /> : <ExpandMoreIcon />}
              sx={{ alignSelf: 'flex-start' }}
            >
              {showAdvanced ? '隐藏高级选项' : '显示高级选项'}
            </Button>

            {/* 高级模式：直接编辑 cron 表达式 */}
            <Collapse in={showAdvanced}>
              <TextField
                fullWidth
                size="small"
                label="Cron 表达式"
                value={value || ''}
                onChange={(e) => onChange(e.target.value)}
                error={!isValid}
                helperText={!isValid ? '格式错误：分 时 日 月 周' : '格式: 分 时 日 月 周，如 0 */6 * * *'}
                placeholder="分 时 日 月 周"
              />
            </Collapse>
          </Stack>
        </Paper>
      </Collapse>

      {/* 当前表达式显示（非自定义模式或已折叠时显示） */}
      {!showCustom && value && (
        <Paper
          variant="outlined"
          sx={{
            p: 1.5,
            mb: 2,
            display: 'flex',
            alignItems: 'center',
            gap: 1,
            backgroundColor: isDark ? withAlpha(palette.background.default, 0.92) : alpha(theme.palette.background.default, 0.3),
            borderColor: isDark ? withAlpha(palette.divider, 0.84) : alpha(theme.palette.divider, 0.5),
            borderRadius: 1.5
          }}
        >
          <Typography variant="body2" color="text.secondary">
            当前表达式:
          </Typography>
          <Chip
            label={value}
            size="small"
            color={isValid ? 'default' : 'error'}
            icon={isValid ? <CheckCircleIcon /> : <ErrorIcon />}
            sx={{
              bgcolor: subtleChipBackground,
              color: isDark && isValid ? alpha(theme.palette.text.primary, 0.9) : undefined,
              borderColor: isDark ? withAlpha(palette.divider, 0.72) : undefined
            }}
          />
        </Paper>
      )}

      {/* 下次运行时间预览 */}
      {nextRuns.length > 0 && (
        <Paper
          variant="outlined"
          sx={{
            p: isMobile ? 1.5 : 2,
            backgroundColor: isDark ? alpha(theme.palette.success.main, 0.08) : alpha(theme.palette.success.main, 0.05),
            borderColor: alpha(theme.palette.success.main, isDark ? 0.24 : 0.3),
            borderRadius: 2
          }}
        >
          <Typography
            variant="subtitle2"
            sx={{
              mb: 1,
              display: 'flex',
              alignItems: 'center',
              gap: 0.5,
              color: theme.palette.success.main
            }}
          >
            <ScheduleIcon fontSize="small" />
            下次运行时间预览
          </Typography>
          <Stack spacing={0.5}>
            {nextRuns.map((run, index) => (
              <Box
                key={index}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 1,
                  py: 0.5,
                  borderBottom: index < nextRuns.length - 1 ? `1px dashed ${alpha(theme.palette.divider, 0.5)}` : 'none'
                }}
              >
                <Chip
                  label={`第${index + 1}次`}
                  size="small"
                  variant="outlined"
                  sx={{
                    minWidth: 56,
                    fontSize: '0.7rem',
                    bgcolor: subtleChipBackground,
                    borderColor: alpha(theme.palette.success.main, isDark ? 0.22 : 0.18)
                  }}
                />
                <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 500 }}>
                  {isMobile ? formatDateTimeShort(run) : formatDateTime(run)}
                </Typography>
                <Typography variant="caption" color="text.secondary" sx={{ ml: 'auto' }}>
                  {formatRelativeTime(run)}
                </Typography>
              </Box>
            ))}
          </Stack>
        </Paper>
      )}

      {/* 无效表达式提示 */}
      {value && !isValid && (
        <Paper
          variant="outlined"
          sx={{
            p: 1.5,
            mt: 1,
            backgroundColor: isDark ? alpha(theme.palette.error.main, 0.08) : alpha(theme.palette.error.main, 0.05),
            borderColor: alpha(theme.palette.error.main, isDark ? 0.24 : 0.3),
            borderRadius: 1.5
          }}
        >
          <Typography variant="body2" color="error" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <ErrorIcon fontSize="small" />
            表达式格式不正确，请检查输入
          </Typography>
        </Paper>
      )}

      {/* 帮助文本 */}
      {helperText && (
        <Typography variant="caption" color={error ? 'error' : 'text.secondary'} sx={{ mt: 1, display: 'block' }}>
          {helperText}
        </Typography>
      )}
    </Box>
  );
}

CronExpressionGenerator.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  label: PropTypes.string,
  helperText: PropTypes.string,
  error: PropTypes.bool
};
