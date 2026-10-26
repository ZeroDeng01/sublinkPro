import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';

// flag-icons 提供随构建打包的本地 SVG 旗帜；Windows 等平台不渲染国旗 emoji，需要位图/矢量图替代
import 'flag-icons/css/flag-icons.min.css';

const COUNTRY_CODE_PATTERN = /^[a-z]{2}$/;

/**
 * 国家旗帜图标
 * 基于 flag-icons 的本地 SVG，避免国旗 emoji 在部分平台退化为字母
 */
export default function CountryFlag({ code, height = 14, sx }) {
  const normalizedCode = String(code || '')
    .trim()
    .toLowerCase();

  if (!COUNTRY_CODE_PATTERN.test(normalizedCode)) return null;

  return (
    <Box
      component="span"
      className={`fi fi-${normalizedCode}`}
      sx={{
        display: 'inline-block',
        width: Math.round((height * 4) / 3),
        height,
        borderRadius: '2px',
        backgroundSize: 'contain',
        flexShrink: 0,
        ...sx
      }}
    />
  );
}

CountryFlag.propTypes = {
  code: PropTypes.string,
  height: PropTypes.number,
  sx: PropTypes.object
};
