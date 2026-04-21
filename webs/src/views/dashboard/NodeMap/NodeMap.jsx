import { useEffect, useRef, useMemo } from 'react';
import * as echarts from 'echarts';
// import 'echarts-gl'; // 3D Not needed for flat map
import { alpha, useColorScheme, useTheme } from '@mui/material/styles';
import { Box, Card, Typography, CircularProgress } from '@mui/material';
import worldJson from 'assets/world.json';
import { withAlpha } from 'utils/colorUtils';
import { COUNTRY_COORDINATES, COUNTRY_NAME_MAP } from './countryData';

// Register standard world map
echarts.registerMap('world', worldJson);

// Helper: Get Flag Emoji from Country Code
const getFlagEmoji = (countryCode) => {
  if (!countryCode) return '';
  // Special case for Taiwan -> China flag as requested
  if (countryCode.toUpperCase() === 'TW') return '🇨🇳'; // CN Flag

  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
};

// Helper: Attempt to find country code by name (case insensitive)
const findCountryCode = (name) => {
  if (!name) return null;
  const lowerName = name.toLowerCase().trim();

  // 1. Direct check in Coordinates keys
  if (COUNTRY_COORDINATES[name.toUpperCase()]) return name.toUpperCase();

  // 2. Map lookup
  if (COUNTRY_NAME_MAP[lowerName]) return COUNTRY_NAME_MAP[lowerName];

  // 3. 2-letter code check
  if (name.length === 2 && COUNTRY_COORDINATES[name.toUpperCase()]) return name.toUpperCase();

  return null;
};

const NodeMap = ({ data = {}, loading = false }) => {
  const chartRef = useRef(null);
  const theme = useTheme();
  const { mode } = useColorScheme();
  const palette = theme.vars?.palette || theme.palette;
  const isDark = mode === 'dark';
  const chartInstance = useRef(null);
  const accentColor = theme.palette.info.main;
  const accentSoft = theme.palette.info.light;
  const successColor = theme.palette.success.main;
  const warningColor = theme.palette.warning.main;
  const darkText = palette.text?.dark || theme.palette.common.white;
  const readablePrimaryTextColor = isDark ? withAlpha(darkText, 0.94) : theme.palette.text.primary;
  const readableSecondaryTextColor = isDark ? withAlpha(darkText, 0.78) : theme.palette.text.secondary;
  const readableTertiaryTextColor = isDark ? withAlpha(darkText, 0.68) : alpha(theme.palette.text.primary, 0.72);
  const cardBackground = isDark
    ? 'radial-gradient(circle at center, rgba(17, 42, 78, 0.98) 0%, rgba(6, 18, 38, 0.99) 72%, rgba(3, 10, 24, 1) 100%)'
    : `radial-gradient(circle at center, ${alpha(theme.palette.background.paper, 1)} 0%, ${alpha(theme.palette.background.default, 0.94)} 100%)`;
  const cardBorderColor = isDark ? alpha('#4dd0e1', 0.24) : 'divider';
  const overlaySurface = isDark
    ? 'linear-gradient(180deg, rgba(10, 30, 58, 0.78) 0%, rgba(5, 18, 38, 0.84) 100%)'
    : alpha(theme.palette.background.paper, 0.72);
  const overlayBorderColor = isDark ? alpha('#63f2ff', 0.24) : alpha(accentColor, 0.16);
  const overlayInsetHighlight = isDark
    ? `0 0 0 1px ${alpha('#63f2ff', 0.08)}, inset 0 1px 0 ${alpha(theme.palette.common.white, 0.06)}`
    : 'none';
  const tooltipSurface = isDark ? 'rgba(8, 24, 46, 0.92)' : alpha(theme.palette.background.paper, 0.96);
  const tooltipBorderColor = isDark ? alpha('#63f2ff', 0.36) : alpha(accentColor, 0.6);
  const tooltipShadow = isDark ? `0 0 0 1px ${alpha('#63f2ff', 0.14)}, 0 18px 40px ${alpha('#04101f', 0.72)}` : 'none';
  const mapAreaColor = isDark ? 'rgba(7, 28, 54, 0.94)' : alpha(theme.palette.background.paper, 0.88);
  const mapBorderColor = isDark ? alpha('#56d9ff', 0.24) : alpha(theme.palette.divider, 0.9);
  const mapEmphasisColor = isDark ? alpha('#31d7ff', 0.16) : alpha(theme.palette.primary.main, 0.12);
  const chinaAreaColor = isDark ? alpha('#ffd166', 0.24) : alpha(theme.palette.warning.main, 0.14);
  const lineTrailColor = isDark ? '#8df8ff' : theme.palette.common.white;
  const lineStrokeColor = isDark ? '#33ddff' : accentColor;
  const lineOpacity = isDark ? 0.62 : 0.4;
  const titleGlow = isDark ? alpha('#63f2ff', 0.42) : alpha(accentColor, 0.3);
  const coverageGlow = isDark ? alpha('#3ce6ff', 0.46) : alpha(accentColor, 0.4);
  const statusDotGlow = isDark ? alpha('#63f2ff', 0.58) : alpha(successColor, 0.65);
  const targetLabelBackground = isDark ? 'rgba(8, 20, 40, 0.78)' : alpha(theme.palette.background.default, 0.72);
  const darkNodeColor = '#52f7ff';
  const darkTargetColor = '#ffd166';

  // Reuse processLines logic if needed, but for 2D map we normally rely on the map json borders.
  // However, if we want custom glowing borders on top, we can keep using mapLines but transformed to 2D.
  // simpler to just use geo itemStyle for borders first.

  // Memoize data processing
  const { points, lines, targetPoint, unknownCount } = useMemo(() => {
    const pts = [];
    const lns = [];
    let tPoint = null;
    let unk = 0;
    let max = 0;
    const chinaCoords = COUNTRY_COORDINATES['CN'];

    // Create Target Point (China)
    if (chinaCoords) {
      tPoint = {
        name: 'CN',
        value: [...chinaCoords, 0], // Flat coords
        itemStyle: {
          color: warningColor
        },
        rippleEffect: {
          brushType: 'stroke',
          scale: 4,
          period: 4
        }
      };
    }

    Object.entries(data).forEach(([key, count]) => {
      const countryCode = findCountryCode(key);

      if (!countryCode || !COUNTRY_COORDINATES[countryCode]) {
        unk += count;
        return;
      }

      if (count > max) max = count;
      const coords = COUNTRY_COORDINATES[countryCode];

      // Specific handling for China
      if (countryCode === 'CN') {
        // Just update label or target info if needed
        return;
      }

      // Normal Points - using effectScatter for 2D
      pts.push({
        name: countryCode,
        value: [...coords, count, countryCode], // [lon, lat, count, code]
        itemStyle: {
          color: isDark ? darkNodeColor : accentColor
        }
      });

      // Flying Lines to China
      if (chinaCoords) {
        lns.push({
          coords: [coords, chinaCoords],
          lineStyle: {
            color: isDark ? darkNodeColor : accentSoft
          }
        });
      }
    });

    return { points: pts, lines: lns, targetPoint: tPoint, unknownCount: unk };
  }, [data, accentColor, accentSoft, warningColor, isDark]);
  const coveredRegionCount = useMemo(
    () => points.length + (Object.prototype.hasOwnProperty.call(data, 'CN') ? 1 : 0),
    [data, points.length]
  );

  useEffect(() => {
    if (!chartRef.current || loading) return;

    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    const option = {
      backgroundColor: 'transparent', // Let Card background show
      tooltip: {
        show: true,
        trigger: 'item',
        backgroundColor: tooltipSurface,
        borderColor: tooltipBorderColor,
        borderWidth: 1,
        textStyle: {
          color: readablePrimaryTextColor,
          fontFamily: '"Noto Sans SC", sans-serif'
        },
        padding: [12, 16],
        extraCssText: `backdrop-filter: blur(${isDark ? 10 : 6}px); box-shadow: ${tooltipShadow}; border-radius: 12px;`,
        formatter: (params) => {
          // const code = params.value && (params.value[3] || params.value[2]); // Unused
          // In 2D scatter: value is [lon, lat, count, code] -> index 3
          // In Target: value is [lon, lat, 0] -> no code? target has name 'CN'

          if (params.seriesType === 'lines') return ''; // No tooltip for lines

          // Target check
          if (params.name === 'CN' || (params.data && params.data.name === 'CN')) {
            return `<div style="display:flex;align-items:center;gap:12px">
                             <span style="font-size:24px;line-height:1">🇨🇳</span>
                              <span style="font-size:16px;font-weight:bold;color:${readablePrimaryTextColor}">CN (本地区域)</span>
                           </div>`;
          }

          const valCode = params.value && params.value[3];
          if (!valCode) return '';

          const count = params.value[2];
          const flagEmoji = getFlagEmoji(valCode);

          return `<div style="display:flex;align-items:center;gap:12px">
                    <span style="font-size:24px;line-height:1">${flagEmoji}</span>
                    <span style="font-size:16px;font-weight:bold;color:${readablePrimaryTextColor}">${valCode}</span>
                 </div>
                 <div style="margin-top:8px;font-size:12px;color:${readableSecondaryTextColor};display:flex;justify-content:space-between;width:120px">
                     <span>节点数量</span>
                     <span style="color:${params.color};font-weight:bold;font-family:monospace;font-size:14px">${count}</span>
                   </div>`;
        }
      },
      geo: {
        map: 'world',
        roam: true, // Allow zoom/pan
        zoom: 1.2,
        label: {
          emphasis: {
            show: false
          }
        },
        itemStyle: {
          normal: {
            areaColor: mapAreaColor,
            borderColor: mapBorderColor,
            borderWidth: 1.5
          },
          emphasis: {
            areaColor: mapEmphasisColor
          }
        },
        regions: [
          {
            name: 'China',
            itemStyle: {
              areaColor: chinaAreaColor
            }
          }
        ]
      },
      series: [
        // 1. Flying Lines
        {
          type: 'lines',
          zlevel: 1, // Keep lines below points
          effect: {
            show: true,
            period: 6,
            trailLength: 0.7,
            color: lineTrailColor,
            symbolSize: 3
          },
          lineStyle: {
            normal: {
              color: lineStrokeColor,
              width: 0,
              curveness: 0.2
            }
          },
          data: lines
        },
        // 2. Flying Lines (Trail base)
        {
          type: 'lines',
          zlevel: 2,
          symbol: ['none', 'arrow'],
          symbolSize: 5,
          effect: {
            show: true,
            period: 6,
            trailLength: 0,
            symbol: 'arrow', // Arrow animation
            symbolSize: 6
          },
          lineStyle: {
            normal: {
              color: lineStrokeColor,
              width: 1,
              opacity: lineOpacity,
              curveness: 0.2
            }
          },
          data: lines
        },
        // 3. Effect Scatter (Nodes)
        {
          type: 'effectScatter',
          coordinateSystem: 'geo',
          zlevel: 2,
          rippleEffect: {
            brushType: 'stroke',
            scale: isDark ? 4 : 3
          },
          label: {
            show: true,
            position: 'right',
            formatter: (params) => {
              const code = params.value[3];
              return getFlagEmoji(code);
            },
            fontSize: 14,
            distance: 5,
            color: isDark ? alpha('#caffff', 0.96) : undefined,
            textShadowColor: isDark ? alpha('#38e6ff', 0.55) : 'transparent',
            textShadowBlur: isDark ? 8 : 0
          },
          symbolSize: (val) => Math.max(6, Math.min(20, Math.log2(val[2] + 1) * 5)),
          itemStyle: {
            color: isDark ? darkNodeColor : accentColor,
            shadowBlur: isDark ? 14 : 0,
            shadowColor: isDark ? alpha(darkNodeColor, 0.55) : 'transparent'
          },
          data: points
        },
        // 4. Target Point (China)
        ...(targetPoint
          ? [
              {
                type: 'effectScatter',
                coordinateSystem: 'geo',
                zlevel: 3,
                rippleEffect: {
                  brushType: 'stroke',
                  scale: 5,
                  period: 3,
                  color: isDark ? darkTargetColor : warningColor
                },
                symbol: 'pin',
                symbolSize: 20,
                itemStyle: {
                  color: isDark ? darkTargetColor : warningColor,
                  shadowBlur: isDark ? 16 : 0,
                  shadowColor: isDark ? alpha(darkTargetColor, 0.52) : 'transparent'
                },
                label: {
                  show: true,
                  formatter: '🇨🇳 CN',
                  position: 'top',
                  fontWeight: 'bold',
                  color: isDark ? alpha('#ffe8a3', 0.98) : warningColor,
                  fontSize: 14,
                  backgroundColor: targetLabelBackground,
                  padding: [4, 6],
                  borderRadius: 4
                },
                data: [targetPoint]
              }
            ]
          : [])
      ]
    };

    chartInstance.current.setOption(option, true);

    const handleResize = () => {
      chartInstance.current?.resize();
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chartInstance.current?.dispose();
      chartInstance.current = null;
    };
  }, [
    points,
    lines,
    targetPoint,
    loading,
    theme,
    accentColor,
    accentSoft,
    warningColor,
    tooltipSurface,
    tooltipBorderColor,
    tooltipShadow,
    readablePrimaryTextColor,
    readableSecondaryTextColor,
    mapAreaColor,
    mapBorderColor,
    mapEmphasisColor,
    chinaAreaColor,
    lineTrailColor,
    lineStrokeColor,
    lineOpacity,
    isDark,
    targetLabelBackground,
    darkNodeColor,
    darkTargetColor
  ]);

  return (
    <Card
      sx={{
        height: '100%',
        width: '100%',
        background: cardBackground,
        color: 'text.primary',
        position: 'relative',
        overflow: 'hidden',
        boxShadow: isDark ? `0 22px 44px ${alpha('#020915', 0.56)}, inset 0 0 0 1px ${alpha('#63f2ff', 0.08)}` : 'none',
        border: '1px solid',
        borderColor: cardBorderColor,
        borderRadius: 0
      }}
    >
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          opacity: isDark ? 0.16 : 0.1,
          backgroundImage: `
                linear-gradient(${alpha(isDark ? '#3ce6ff' : accentColor, isDark ? 0.12 : 0.24)} 1px, transparent 1px),
                linear-gradient(90deg, ${alpha(isDark ? '#3ce6ff' : accentColor, isDark ? 0.12 : 0.24)} 1px, transparent 1px)
             `,
          backgroundSize: '40px 40px',
          pointerEvents: 'none'
        }}
      />

      <Box sx={{ p: 4, position: 'absolute', top: 0, left: 0, zIndex: 10 }}>
        <Box
          sx={{
            display: 'inline-flex',
            flexDirection: 'column',
            gap: 1.25,
            px: 3,
            py: 2.5,
            background: overlaySurface,
            border: isDark ? '1px solid' : 'none',
            borderColor: overlayBorderColor,
            boxShadow: overlayInsetHighlight,
            backdropFilter: `blur(${isDark ? 10 : 4}px)`
          }}
        >
          <Typography
            variant="h3"
            sx={{
              color: readablePrimaryTextColor,
              fontWeight: 800,
              textShadow: `0 0 ${isDark ? 30 : 24}px ${titleGlow}`,
              letterSpacing: '4px',
              mb: 0.25,
              fontFamily: '"Orbitron", sans-serif'
            }}
          >
            全球节点分布
          </Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
            <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: successColor, boxShadow: `0 0 10px ${statusDotGlow}` }} />
            <Typography
              variant="subtitle2"
              sx={{ color: isDark ? alpha('#98f7ff', 0.9) : accentColor, letterSpacing: '2px', fontFamily: '"Noto Sans SC", monospace' }}
            >
              以获取到落地IP的数据为参考
            </Typography>
          </Box>
        </Box>

        {unknownCount > 0 && (
          <Box
            sx={{
              mt: 3,
              display: 'inline-flex',
              alignItems: 'center',
              gap: 2,
              background: overlaySurface,
              border: isDark ? '1px solid' : 'none',
              borderColor: overlayBorderColor,
              py: 1.5,
              px: 3,
              borderRadius: 0,
              borderLeft: `4px solid ${accentColor}`,
              boxShadow: overlayInsetHighlight,
              backdropFilter: `blur(${isDark ? 10 : 4}px)`
            }}
          >
            <Typography
              variant="body2"
              sx={{ color: readableSecondaryTextColor, fontFamily: '"Noto Sans SC", monospace', fontSize: '0.9rem' }}
            >
              未知区域 &gt;&gt; <span style={{ color: readablePrimaryTextColor, fontWeight: 'bold' }}>{unknownCount}</span>
            </Typography>
          </Box>
        )}
      </Box>

      {/* Stats overlay */}
      <Box sx={{ p: 6, position: 'absolute', bottom: 0, right: 0, zIndex: 10, textAlign: 'right', pointerEvents: 'none' }}>
        <Box
          sx={{
            display: 'inline-flex',
            flexDirection: 'column',
            alignItems: 'flex-end',
            px: 3,
            py: 2.5,
            background: overlaySurface,
            border: isDark ? '1px solid' : 'none',
            borderColor: overlayBorderColor,
            boxShadow: overlayInsetHighlight,
            backdropFilter: `blur(${isDark ? 10 : 4}px)`
          }}
        >
          <Typography
            variant="overline"
            sx={{ color: readableTertiaryTextColor, letterSpacing: 2, display: 'block', fontFamily: '"Noto Sans SC"' }}
          >
            覆盖区域
          </Typography>
          <Typography
            variant="h1"
            sx={{
              color: isDark ? alpha('#7ff9ff', 0.98) : accentColor,
              fontWeight: '900',
              fontSize: '4rem',
              lineHeight: 1,
              textShadow: `0 0 ${isDark ? 34 : 28}px ${coverageGlow}`
            }}
          >
            {String(coveredRegionCount).padStart(2, '0')}
          </Typography>

          <Box sx={{ mt: 2, height: 2, width: 100, bgcolor: alpha(isDark ? '#63f2ff' : accentColor, isDark ? 0.9 : 0.8), ml: 'auto' }} />
        </Box>
      </Box>

      <Box
        ref={chartRef}
        sx={{
          width: '100%',
          height: '100%',
          opacity: loading ? 0 : 1,
          transition: 'opacity 1s ease-in-out'
        }}
      />

      {loading && (
        <Box
          sx={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 2
          }}
        >
          <CircularProgress size={60} sx={{ color: accentColor }} />
          <Typography sx={{ color: readableSecondaryTextColor, letterSpacing: 4, fontFamily: '"Noto Sans SC", monospace' }}>
            正在初始化地图...
          </Typography>
        </Box>
      )}
    </Card>
  );
};

export default NodeMap;
