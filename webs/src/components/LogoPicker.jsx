import { useState, useMemo } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// icons
import * as TablerIcons from '@tabler/icons-react';
import ClearIcon from '@mui/icons-material/Clear';
import SearchIcon from '@mui/icons-material/Search';

// local components
import AirportLogo from 'views/airports/component/AirportLogo';

// 精选的 Tabler 图标列表
const ICON_LIST = [
  'IconPlane',
  'IconPlaneTilt',
  'IconPlaneArrival',
  'IconPlaneDeparture',
  'IconRocket',
  'IconRocketOff',
  'IconSatellite',
  'IconWorld',
  'IconWorldWww',
  'IconCloud',
  'IconCloudComputing',
  'IconServer',
  'IconServerBolt',
  'IconDatabase',
  'IconShield',
  'IconShieldCheck',
  'IconShieldLock',
  'IconLock',
  'IconLockOpen',
  'IconKey',
  'IconFingerprint',
  'IconEye',
  'IconEyeOff',
  'IconBolt',
  'IconFlame',
  'IconZap',
  'IconStar',
  'IconStarFilled',
  'IconHeart',
  'IconDiamond',
  'IconCrown',
  'IconTrophy',
  'IconMedal',
  'IconAward',
  'IconFlag',
  'IconFlagFilled',
  'IconBookmark',
  'IconTag',
  'IconTags',
  'IconHome',
  'IconBuilding',
  'IconBuildingSkyscraper',
  'IconCastle',
  'IconNetwork',
  'IconTopology',
  'IconRouter',
  'IconWifi',
  'IconAntenna',
  'IconArrowsExchange',
  'IconRefresh',
  'IconDownload',
  'IconUpload',
  'IconBrandGithub',
  'IconBrandGoogle',
  'IconBrandTelegram',
  'IconBrandDiscord',
  'IconCode',
  'IconTerminal',
  'IconApi',
  'IconBraces',
  'IconBrackets',
  'IconCpu',
  'IconDevices',
  'IconDeviceDesktop',
  'IconDeviceMobile',
  'IconGlobe',
  'IconMap',
  'IconMapPin',
  'IconCompass',
  'IconNavigation',
  'IconSun',
  'IconMoon',
  'IconCloudSun',
  'IconSnowflake',
  'IconRainbow',
  'IconLeaf',
  'IconTree',
  'IconFlower',
  'IconSeeding',
  'IconCat',
  'IconDog',
  'IconFish',
  'IconBird',
  'IconButterfly',
  'IconMusic',
  'IconHeadphones',
  'IconMicrophone',
  'IconVolume',
  'IconPhoto',
  'IconCamera',
  'IconVideo',
  'IconMovie',
  'IconPalette',
  'IconBrush',
  'IconPencil',
  'IconPen',
  'IconMessage',
  'IconMail',
  'IconBell',
  'IconBellRinging',
  'IconSettings',
  'IconAdjustments',
  'IconTool',
  'IconWrench',
  'IconUser',
  'IconUsers',
  'IconUserCircle',
  'IconUserShield'
];

// 常用 Emoji 列表
const EMOJI_LIST = [
  // 交通
  '✈️',
  '🛫',
  '🛬',
  '🚀',
  '🛸',
  '🚁',
  '🛩️',
  '🎈',
  '🪂',
  // 地球与自然
  '🌍',
  '🌎',
  '🌏',
  '🌐',
  '🗺️',
  '🧭',
  '🏔️',
  '⛰️',
  '🌋',
  '🏝️',
  // 天气
  '☀️',
  '🌙',
  '⭐',
  '🌟',
  '✨',
  '💫',
  '⚡',
  '🔥',
  '🌈',
  '☁️',
  '❄️',
  // 动物
  '🦅',
  '🦋',
  '🐝',
  '🐬',
  '🦈',
  '🐙',
  '🦑',
  '🦀',
  '🐢',
  '🦎',
  '🐉',
  '🦄',
  '🐺',
  '🦊',
  '🐱',
  '🐶',
  '🐰',
  '🐻',
  '🐼',
  '🦁',
  // 植物
  '🌸',
  '🌺',
  '🌹',
  '🌻',
  '🌼',
  '🍀',
  '🌿',
  '🌲',
  '🌳',
  '🍃',
  // 食物
  '🍎',
  '🍊',
  '🍋',
  '🍇',
  '🍓',
  '🍒',
  '🥝',
  '🍑',
  '🥭',
  '🍍',
  // 心形
  '❤️',
  '🧡',
  '💛',
  '💚',
  '💙',
  '💜',
  '🖤',
  '🤍',
  '💖',
  '💝',
  // 符号
  '⚡',
  '💎',
  '🔮',
  '🎯',
  '🎪',
  '🎭',
  '🎨',
  '🎬',
  '🎮',
  '🎲',
  '🏆',
  '🥇',
  '🥈',
  '🥉',
  '🏅',
  '🎖️',
  '👑',
  '💍',
  '🔑',
  '🗝️',
  '🛡️',
  '⚔️',
  '🔒',
  '🔓',
  '🔐',
  '💡',
  '🔦',
  '📡',
  '🖥️',
  '💻',
  '📱',
  '⌚',
  '🎧',
  '🎤',
  '📷',
  '📹',
  '🔭',
  '🔬',
  '💉',
  '💊',
  // 国旗（常用）
  '🇺🇸',
  '🇯🇵',
  '🇰🇷',
  '🇨🇳',
  '🇭🇰',
  '🇹🇼',
  '🇸🇬',
  '🇬🇧',
  '🇩🇪',
  '🇫🇷',
  '🇨🇦',
  '🇦🇺',
  '🇳🇿',
  '🇮🇳',
  '🇷🇺',
  '🇧🇷',
  '🇲🇽',
  '🇮🇹',
  '🇪🇸',
  '🇳🇱'
];

/**
 * Logo选择器组件
 * 支持URL输入、图标选择、Emoji选择三种模式
 */
export default function LogoPicker({ value, onChange, name }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState(0);
  const [search, setSearch] = useState('');
  const [urlInput, setUrlInput] = useState('');

  // 判断是否为图片URL（包括http/https和base64格式）
  const isImageUrl = (val) => {
    if (!val) return false;
    return val.startsWith('http://') || val.startsWith('https://') || val.startsWith('data:image');
  };

  // 打开对话框时初始化
  const handleOpen = () => {
    setOpen(true);
    // 如果当前是URL（包括base64），填入输入框
    if (isImageUrl(value)) {
      setUrlInput(value);
      setTab(0);
    } else if (value && value.startsWith('icon:')) {
      setTab(1);
    } else if (value) {
      setTab(2);
    }
    setSearch('');
  };

  const handleClose = () => {
    setOpen(false);
    setSearch('');
  };

  // 选择图标
  const handleSelectIcon = (iconName) => {
    onChange(`icon:${iconName}`);
    handleClose();
  };

  // 选择Emoji
  const handleSelectEmoji = (emoji) => {
    onChange(emoji);
    handleClose();
  };

  // 确认URL
  const handleConfirmUrl = () => {
    if (urlInput.trim()) {
      onChange(urlInput.trim());
    }
    handleClose();
  };

  // 清除Logo
  const handleClear = () => {
    onChange('');
    setUrlInput('');
    handleClose();
  };

  // 过滤图标
  const filteredIcons = useMemo(() => {
    if (!search) return ICON_LIST;
    const lower = search.toLowerCase();
    return ICON_LIST.filter((name) => name.toLowerCase().includes(lower));
  }, [search]);

  // 过滤Emoji（按搜索词过滤比较困难，这里只做简单显示）
  const filteredEmojis = EMOJI_LIST;

  return (
    <>
      {/* 预览和触发按钮 */}
      <Box
        onClick={handleOpen}
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
          p: 1.5,
          borderRadius: 2,
          border: `1px dashed ${alpha(theme.palette.divider, 0.5)}`,
          cursor: 'pointer',
          transition: 'all 0.2s',
          '&:hover': {
            borderColor: theme.palette.primary.main,
            bgcolor: alpha(theme.palette.primary.main, 0.04)
          }
        }}
      >
        <AirportLogo logo={value} name={name} size="medium" />
        <Box sx={{ flex: 1 }}>
          <Typography variant="body2" fontWeight={500}>
            {value ? t('airports.form.logoPicker.change') : t('airports.form.logoPicker.set')}
          </Typography>
          <Typography variant="caption" color="textSecondary">
            {value
              ? value.startsWith('icon:')
                ? t('airports.form.logoPicker.types.icon')
                : value.startsWith('http') || value.startsWith('data:image')
                  ? t('airports.form.logoPicker.types.urlImage')
                  : t('airports.form.logoPicker.types.emoji')
              : t('airports.form.logoPicker.types.initial')}
          </Typography>
        </Box>
        {value && (
          <Tooltip title={t('common.clear')} arrow>
            <IconButton
              size="small"
              onClick={(e) => {
                e.stopPropagation();
                onChange('');
              }}
            >
              <ClearIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        )}
      </Box>

      {/* 选择对话框 */}
      <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
        <DialogTitle>{t('airports.form.logoPicker.dialogTitle')}</DialogTitle>
        <DialogContent sx={{ pt: 1 }}>
          {/* Tab切换 */}
          <Tabs
            value={tab}
            onChange={(_, v) => {
              setTab(v);
              setSearch('');
            }}
            sx={{ mb: 2, borderBottom: 1, borderColor: 'divider' }}
          >
            <Tab label={t('airports.form.logoPicker.tabs.url')} />
            <Tab label={t('airports.form.logoPicker.tabs.icon')} />
            <Tab label={t('airports.form.logoPicker.tabs.emoji')} />
          </Tabs>

          {/* URL输入 */}
          {tab === 0 && (
            <Box>
              <TextField
                fullWidth
                size="small"
                label={t('airports.form.logoPicker.urlLabel')}
                placeholder={t('airports.form.logoPicker.urlPlaceholder')}
                value={urlInput}
                onChange={(e) => setUrlInput(e.target.value)}
                helperText={t('airports.form.logoPicker.urlHelper')}
              />
              {urlInput && (
                <Box sx={{ mt: 2, display: 'flex', justifyContent: 'center' }}>
                  <Box
                    sx={{
                      p: 2,
                      borderRadius: 2,
                      bgcolor: alpha(theme.palette.divider, 0.1),
                      textAlign: 'center',
                      maxWidth: '100%'
                    }}
                  >
                    <Typography variant="caption" color="textSecondary" sx={{ mb: 1, display: 'block' }}>
                      {t('common.preview')}
                    </Typography>
                    <Box
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        minHeight: 60,
                        maxHeight: 120,
                        overflow: 'hidden'
                      }}
                    >
                      <Box
                        component="img"
                        src={urlInput}
                        alt={t('common.preview')}
                        referrerPolicy="no-referrer"
                        sx={{
                          maxWidth: 200,
                          maxHeight: 100,
                          objectFit: 'contain',
                          borderRadius: 1
                        }}
                        onError={(e) => {
                          e.target.style.display = 'none';
                        }}
                      />
                    </Box>
                  </Box>
                </Box>
              )}
            </Box>
          )}

          {/* 图标选择 */}
          {tab === 1 && (
            <Box>
              <TextField
                fullWidth
                size="small"
                placeholder={t('airports.form.logoPicker.searchIcon')}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                slotProps={{
                  input: {
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchIcon fontSize="small" />
                      </InputAdornment>
                    )
                  }
                }}
                sx={{ mb: 2 }}
              />
              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fill, minmax(48px, 1fr))',
                  gap: 1,
                  maxHeight: 300,
                  overflowY: 'auto',
                  p: 0.5
                }}
              >
                {filteredIcons.map((iconName) => {
                  const IconComponent = TablerIcons[iconName];
                  if (!IconComponent) return null;
                  const isSelected = value === `icon:${iconName}`;
                  return (
                    <Tooltip key={iconName} title={iconName.replace('Icon', '')} arrow>
                      <Box
                        onClick={() => handleSelectIcon(iconName)}
                        sx={{
                          width: 48,
                          height: 48,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          borderRadius: 2,
                          cursor: 'pointer',
                          border: isSelected ? `2px solid ${theme.palette.primary.main}` : '1px solid transparent',
                          bgcolor: isSelected ? alpha(theme.palette.primary.main, 0.1) : 'transparent',
                          transition: 'all 0.15s',
                          '&:hover': {
                            bgcolor: alpha(theme.palette.primary.main, 0.08)
                          }
                        }}
                      >
                        <IconComponent size={24} stroke={1.5} />
                      </Box>
                    </Tooltip>
                  );
                })}
              </Box>
            </Box>
          )}

          {/* Emoji选择 */}
          {tab === 2 && (
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(44px, 1fr))',
                gap: 0.5,
                maxHeight: 320,
                overflowY: 'auto',
                p: 0.5
              }}
            >
              {filteredEmojis.map((emoji, index) => {
                const isSelected = value === emoji;
                return (
                  <Box
                    key={index}
                    onClick={() => handleSelectEmoji(emoji)}
                    sx={{
                      width: 44,
                      height: 44,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: 24,
                      borderRadius: 2,
                      cursor: 'pointer',
                      border: isSelected ? `2px solid ${theme.palette.primary.main}` : '1px solid transparent',
                      bgcolor: isSelected ? alpha(theme.palette.primary.main, 0.1) : 'transparent',
                      transition: 'all 0.15s',
                      '&:hover': {
                        bgcolor: alpha(theme.palette.primary.main, 0.08)
                      }
                    }}
                  >
                    {emoji}
                  </Box>
                );
              })}
            </Box>
          )}
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={handleClear} color="error">
            {t('common.clear')}
          </Button>
          <Box sx={{ flex: 1 }} />
          <Button onClick={handleClose}>{t('common.cancel')}</Button>
          {tab === 0 && (
            <Button variant="contained" onClick={handleConfirmUrl} disabled={!urlInput.trim()}>
              {t('common.confirm')}
            </Button>
          )}
        </DialogActions>
      </Dialog>
    </>
  );
}

LogoPicker.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  name: PropTypes.string
};
