import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';

// icons - 动态导入通用图标
import * as TablerIcons from '@tabler/icons-react';

/**
 * 机场Logo显示组件
 * 支持三种类型：URL图片、icon:图标名、emoji字符
 * 未设置时显示名称首字
 */
export default function AirportLogo({ logo, name, size = 'medium' }) {
  const theme = useTheme();

  // 尺寸配置
  const sizeMap = {
    small: { width: 28, height: 28, fontSize: 12, iconSize: 16 },
    medium: { width: 36, height: 36, fontSize: 14, iconSize: 20 },
    large: { width: 48, height: 48, fontSize: 18, iconSize: 28 }
  };

  const { width, height, fontSize, iconSize } = sizeMap[size] || sizeMap.medium;

  // 获取名称首字
  const getInitial = () => {
    if (!name) return '?';
    // 处理中文和英文
    return name.charAt(0).toUpperCase();
  };

  // 根据名称生成一个稳定的颜色
  const getColorFromName = (str) => {
    if (!str) return theme.palette.primary.main;
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const colors = [
      theme.palette.primary.main,
      theme.palette.secondary.main,
      theme.palette.success.main,
      theme.palette.info.main,
      theme.palette.warning.main,
      '#9c27b0', // purple
      '#00bcd4', // cyan
      '#ff5722', // deep orange
      '#607d8b', // blue grey
      '#e91e63' // pink
    ];
    return colors[Math.abs(hash) % colors.length];
  };

  // 解析logo类型
  const parseLogoType = () => {
    if (!logo) return { type: 'initial' };

    // URL类型（包括http/https和base64格式）
    if (logo.startsWith('http://') || logo.startsWith('https://') || logo.startsWith('data:image')) {
      return { type: 'url', value: logo };
    }

    // Icon类型
    if (logo.startsWith('icon:')) {
      return { type: 'icon', value: logo.substring(5) };
    }

    // Emoji类型（其他情况视为emoji）
    return { type: 'emoji', value: logo };
  };

  const { type, value } = parseLogoType();
  const bgColor = getColorFromName(name);

  // 渲染URL图片
  if (type === 'url') {
    return (
      <Box
        sx={{
          width,
          height,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderRadius: 1.5,
          bgcolor: alpha(theme.palette.divider, 0.08),
          overflow: 'hidden',
          flexShrink: 0
        }}
      >
        <Box
          component="img"
          src={value}
          alt={name}
          referrerPolicy="no-referrer"
          sx={{
            maxWidth: '100%',
            maxHeight: '100%',
            objectFit: 'contain'
          }}
          onError={(e) => {
            // 图片加载失败时隐藏img，显示首字母
            e.target.style.display = 'none';
          }}
        />
      </Box>
    );
  }

  // 渲染图标
  if (type === 'icon') {
    const IconComponent = TablerIcons[value];
    if (IconComponent) {
      return (
        <Avatar
          sx={{
            width,
            height,
            bgcolor: alpha(bgColor, 0.1),
            color: bgColor
          }}
        >
          <IconComponent size={iconSize} stroke={1.5} />
        </Avatar>
      );
    }
    // 图标不存在则回退到首字母
  }

  // 渲染Emoji
  if (type === 'emoji') {
    return (
      <Box
        sx={{
          width,
          height,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: fontSize + 4,
          lineHeight: 1,
          borderRadius: '50%',
          bgcolor: alpha(bgColor, 0.08)
        }}
      >
        {value}
      </Box>
    );
  }

  // 默认：显示名称首字
  return (
    <Avatar
      sx={{
        width,
        height,
        fontSize,
        fontWeight: 600,
        bgcolor: alpha(bgColor, 0.15),
        color: bgColor
      }}
    >
      {getInitial()}
    </Avatar>
  );
}

AirportLogo.propTypes = {
  logo: PropTypes.string,
  name: PropTypes.string,
  size: PropTypes.oneOf(['small', 'medium', 'large'])
};
