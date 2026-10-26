import { useEffect, useRef, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import * as echarts from 'echarts';
// import 'echarts-gl'; // 3D Not needed for flat map
import { alpha, useTheme } from '@mui/material/styles';
import { Autocomplete, Box, Card, Typography, CircularProgress, TextField } from '@mui/material';
import worldJson from 'assets/world.json';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useLocalStorage } from 'hooks/useLocalStorage';
import { withAlpha } from 'utils/colorUtils';
import { isoToFlag } from 'utils/countryDisplay';
import { COUNTRY_COORDINATES, COUNTRY_NAME_MAP } from './countryData';

// Register standard world map
echarts.registerMap('world', worldJson);

// Helper: Get Flag Emoji from Country Code
const getFlagEmoji = (countryCode) => {
  return isoToFlag(countryCode);
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

const LOCAL_REGION_STORAGE_KEY = 'nodeMapLocalRegion';
const DEFAULT_LOCAL_REGION = 'CN';

// world.json uses older English region names; map them where Intl.DisplayNames differs.
const MAP_REGION_ALIASES = {
  BA: 'Bosnia and Herz.',
  CD: 'Dem. Rep. Congo',
  CF: 'Central African Rep.',
  CG: 'Congo',
  CV: 'Cape Verde',
  CZ: 'Czech Rep.',
  DO: 'Dominican Rep.',
  EH: 'W. Sahara',
  FK: 'Falkland Is.',
  FO: 'Faeroe Is.',
  GS: 'S. Geo. and S. Sandw. Is.',
  HM: 'Heard I. and McDonald Is.',
  KR: 'Korea',
  LA: 'Lao PDR',
  LC: 'Saint Lucia',
  MK: 'Macedonia',
  MM: 'Myanmar',
  MP: 'N. Mariana Is.',
  PF: 'Fr. Polynesia',
  SB: 'Solomon Is.',
  SH: 'Saint Helena',
  SS: 'S. Sudan',
  ST: 'São Tomé and Principe',
  SZ: 'Swaziland',
  TC: 'Turks and Caicos Is.',
  TR: 'Turkey',
  TT: 'Trinidad and Tobago',
  VI: 'U.S. Virgin Is.'
};

const WORLD_REGION_NAMES = new Set((worldJson.features || []).map((feature) => feature.properties?.name).filter(Boolean));

const englishRegionDisplayNames = (() => {
  try {
    return new Intl.DisplayNames(['en'], { type: 'region' });
  } catch {
    return null;
  }
})();

// Resolve the world.json region name for a country code (null when the map has no shape for it).
const resolveMapRegionName = (countryCode) => {
  if (!countryCode) return null;
  if (Object.prototype.hasOwnProperty.call(MAP_REGION_ALIASES, countryCode)) {
    return MAP_REGION_ALIASES[countryCode];
  }
  const displayName = englishRegionDisplayNames?.of(countryCode);
  return displayName && WORLD_REGION_NAMES.has(displayName) ? displayName : null;
};

const NodeMap = ({ data = {}, loading = false }) => {
  const { t, i18n } = useTranslation();
  const chartRef = useRef(null);
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const palette = theme.vars?.palette || theme.palette;
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
  const targetAreaColor = isDark ? alpha('#ffd166', 0.24) : alpha(theme.palette.warning.main, 0.14);
  const lineTrailColor = isDark ? '#8df8ff' : theme.palette.common.white;
  const lineStrokeColor = isDark ? '#33ddff' : accentColor;
  const lineOpacity = isDark ? 0.62 : 0.4;
  const titleGlow = isDark ? alpha('#63f2ff', 0.42) : alpha(accentColor, 0.3);
  const coverageGlow = isDark ? alpha('#3ce6ff', 0.46) : alpha(accentColor, 0.4);
  const statusDotGlow = isDark ? alpha('#63f2ff', 0.58) : alpha(successColor, 0.65);
  const targetLabelBackground = isDark ? 'rgba(8, 20, 40, 0.78)' : alpha(theme.palette.background.paper, 0.92);
  const darkNodeColor = '#52f7ff';
  const darkTargetColor = '#ffd166';

  // Persisted local region used as the map hub
  const { state: storedLocalRegion, setState: setStoredLocalRegion } = useLocalStorage(LOCAL_REGION_STORAGE_KEY, DEFAULT_LOCAL_REGION);
  const storedRegionCode = typeof storedLocalRegion === 'string' ? storedLocalRegion.toUpperCase() : '';
  const localRegion = COUNTRY_COORDINATES[storedRegionCode] ? storedRegionCode : DEFAULT_LOCAL_REGION;

  const countryOptions = useMemo(() => {
    let displayNames = null;
    try {
      displayNames = new Intl.DisplayNames([i18n.language], { type: 'region' });
    } catch {
      displayNames = null;
    }

    return Object.keys(COUNTRY_COORDINATES)
      .map((code) => {
        const name = displayNames?.of(code) || code;
        return { code, name, label: `${getFlagEmoji(code)} ${name} (${code})` };
      })
      .sort((a, b) => a.name.localeCompare(b.name, i18n.language));
  }, [i18n.language]);

  const selectedCountryOption = useMemo(
    () => countryOptions.find((option) => option.code === localRegion) || null,
    [countryOptions, localRegion]
  );

  // Localized name of the hub region, used in the pin label/tooltip (flag emoji alone degrades to letters on some platforms)
  const localRegionName = useMemo(() => {
    try {
      return new Intl.DisplayNames([i18n.language], { type: 'region' }).of(localRegion) || localRegion;
    } catch {
      return localRegion;
    }
  }, [i18n.language, localRegion]);

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
    // The local region itself is rendered as the target pin; keep its node count for the pin label
    let localCount = 0;
    const targetCoords = COUNTRY_COORDINATES[localRegion];

    Object.entries(data).forEach(([key, count]) => {
      const countryCode = findCountryCode(key);

      if (!countryCode || !COUNTRY_COORDINATES[countryCode]) {
        unk += count;
        return;
      }

      if (count > max) max = count;
      const coords = COUNTRY_COORDINATES[countryCode];

      // The local region is rendered as the target pin instead of a regular node point
      if (countryCode === localRegion) {
        localCount += count;
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

      // Flying Lines to the local region
      if (targetCoords) {
        lns.push({
          coords: [coords, targetCoords],
          lineStyle: {
            color: isDark ? darkNodeColor : accentSoft
          }
        });
      }
    });

    // Create the target point for the selected local region (value keeps the node count for the label/tooltip)
    if (targetCoords) {
      tPoint = {
        name: localRegion,
        value: [...targetCoords, localCount], // [lon, lat, count]
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

    return { points: pts, lines: lns, targetPoint: tPoint, unknownCount: unk };
  }, [data, accentColor, accentSoft, warningColor, isDark, localRegion]);
  const coveredRegionCount = useMemo(
    () => points.length + (Object.keys(data).some((key) => findCountryCode(key) === localRegion) ? 1 : 0),
    [data, localRegion, points.length]
  );
  const mapRegionName = resolveMapRegionName(localRegion);

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
          // In Target: value is [lon, lat, 0] -> no code; target has the local region name

          if (params.seriesType === 'lines') return ''; // No tooltip for lines

          // Target check
          if (params.name === localRegion || (params.data && params.data.name === localRegion)) {
            const targetCount = (params.value && params.value[2]) || 0;
            const countRow =
              targetCount > 0
                ? `<div style="margin-top:8px;font-size:12px;color:${readableSecondaryTextColor};display:flex;justify-content:space-between;width:120px">
                     <span>${t('dashboard.nodeMap.nodeCount')}</span>
                     <span style="color:${warningColor};font-weight:bold;font-family:monospace;font-size:14px">${targetCount}</span>
                   </div>`
                : '';
            return `<div style="display:flex;align-items:center;gap:12px">
                             <span style="font-size:24px;line-height:1">${getFlagEmoji(localRegion)}</span>
                              <span style="font-size:16px;font-weight:bold;color:${readablePrimaryTextColor}">${localRegionName} (${t('dashboard.nodeMap.localRegion')})</span>
                           </div>${countRow}`;
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
                     <span>${t('dashboard.nodeMap.nodeCount')}</span>
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
        regions: mapRegionName
          ? [
              {
                name: mapRegionName,
                itemStyle: {
                  areaColor: targetAreaColor
                }
              }
            ]
          : []
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
        // 4. Target Point (local region)
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
                symbolSize: (val) => {
                  const count = (val && val[2]) || 0;
                  return count > 0 ? 20 + Math.min(12, Math.log2(count + 1) * 3) : 20;
                },
                itemStyle: {
                  color: isDark ? darkTargetColor : warningColor,
                  borderColor: isDark ? alpha('#04101f', 0.9) : theme.palette.common.white,
                  borderWidth: 1.5,
                  shadowBlur: isDark ? 16 : 6,
                  shadowColor: isDark ? alpha(darkTargetColor, 0.52) : alpha(warningColor, 0.45)
                },
                label: {
                  show: true,
                  formatter: (params) => {
                    const count = (params.value && params.value[2]) || 0;
                    return `${getFlagEmoji(localRegion)} ${localRegionName}${count > 0 ? ` · ${count}` : ''}`;
                  },
                  position: 'top',
                  fontWeight: 'bold',
                  color: isDark ? alpha('#ffe8a3', 0.98) : theme.palette.warning.dark,
                  fontSize: 13,
                  backgroundColor: targetLabelBackground,
                  borderColor: isDark ? alpha(darkTargetColor, 0.28) : alpha(warningColor, 0.35),
                  borderWidth: 1,
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
    targetAreaColor,
    localRegion,
    mapRegionName,
    localRegionName,
    lineTrailColor,
    lineStrokeColor,
    lineOpacity,
    isDark,
    targetLabelBackground,
    darkNodeColor,
    darkTargetColor,
    t
  ]);

  return (
    <Card
      sx={(theme) => ({
        height: '100%',
        width: '100%',
        background: cardBackground,
        color: 'text.primary',
        position: 'relative',
        overflow: 'hidden',
        boxShadow: isDark
          ? `0 18px 42px ${alpha(theme.palette.common.black, 0.22)}, inset 0 1px 0 ${alpha(theme.palette.common.white, 0.035)}`
          : `0 14px 34px ${alpha(theme.palette.primary.dark, 0.055)}, 0 3px 10px ${alpha(theme.palette.grey[500], 0.075)}`,
        border: '1px solid',
        borderColor: isDark ? alpha(theme.palette.common.white, 0.08) : alpha(theme.palette.primary.main, 0.1),
        borderRadius: 3
      })}
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
            maxWidth: 320,
            pointerEvents: 'none'
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
            {t('dashboard.nodeMap.globalDistribution')}
          </Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.25 }}>
            <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: successColor, boxShadow: `0 0 10px ${statusDotGlow}` }} />
            <Typography
              variant="subtitle2"
              sx={{ color: isDark ? alpha('#98f7ff', 0.9) : accentColor, letterSpacing: '1.6px', fontFamily: '"Noto Sans SC", monospace' }}
            >
              {t('dashboard.nodeMap.dataReference')}
            </Typography>
          </Box>
          <Box sx={{ pointerEvents: 'auto', width: 260, mt: 0.5 }}>
            <Autocomplete
              size="small"
              disableClearable
              options={countryOptions}
              value={selectedCountryOption}
              onChange={(_, option) => {
                if (option) setStoredLocalRegion(option.code);
              }}
              getOptionLabel={(option) => option.label}
              isOptionEqualToValue={(option, value) => option.code === value.code}
              filterOptions={(options, state) => {
                const query = state.inputValue.trim().toLowerCase();
                if (!query) return options;
                return options.filter((option) => option.name.toLowerCase().includes(query) || option.code.toLowerCase().includes(query));
              }}
              slotProps={{ paper: { sx: { minWidth: 320 } } }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('dashboard.nodeMap.localRegion')}
                  sx={{
                    '& .MuiOutlinedInput-root': {
                      background: overlaySurface,
                      backdropFilter: `blur(${isDark ? 10 : 4}px)`,
                      '& fieldset': { borderColor: overlayBorderColor },
                      '&:hover fieldset': { borderColor: accentColor },
                      '&.Mui-focused fieldset': { borderColor: accentColor }
                    },
                    '& .MuiInputLabel-root': { color: readableSecondaryTextColor },
                    '& .MuiInputLabel-root.Mui-focused': { color: accentColor },
                    '& .MuiAutocomplete-input': { color: readablePrimaryTextColor }
                  }}
                />
              )}
            />
          </Box>
        </Box>
      </Box>

      <Box sx={{ p: 4, position: 'absolute', right: 0, bottom: 0, zIndex: 10, pointerEvents: 'none' }}>
        <Box
          sx={{
            display: 'inline-flex',
            flexDirection: 'column',
            gap: 1.25,
            px: 2.5,
            py: 2,
            minWidth: 210,
            background: overlaySurface,
            border: '1px solid',
            borderColor: overlayBorderColor,
            boxShadow: overlayInsetHighlight,
            backdropFilter: `blur(${isDark ? 10 : 4}px)`
          }}
        >
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.25 }}>
            <Box>
              <Typography
                variant="overline"
                sx={{ color: readableTertiaryTextColor, letterSpacing: 2, display: 'block', fontFamily: '"Noto Sans SC"' }}
              >
                {t('dashboard.nodeMap.coveredRegions')}
              </Typography>
              <Typography
                variant="h2"
                sx={{
                  color: isDark ? alpha('#7ff9ff', 0.98) : accentColor,
                  fontWeight: '900',
                  lineHeight: 1,
                  textShadow: `0 0 ${isDark ? 28 : 22}px ${coverageGlow}`,
                  fontFamily: '"Orbitron", sans-serif'
                }}
              >
                {String(coveredRegionCount).padStart(2, '0')}
              </Typography>
            </Box>

            <Box
              sx={{
                px: 1.5,
                py: 1.15,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 1.25,
                background: isDark ? 'transparent' : alpha(theme.palette.background.default, 0.8),
                border: '1px solid',
                borderColor: overlayBorderColor,
                borderTop: `2px solid ${unknownCount > 0 ? alpha(isDark ? '#63f2ff' : accentColor, isDark ? 0.72 : 0.6) : alpha(accentColor, 0.3)}`,
                boxShadow: overlayInsetHighlight,
                backdropFilter: `blur(${isDark ? 8 : 4}px)`
              }}
            >
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.2, minWidth: 0 }}>
                <Typography
                  variant="caption"
                  sx={{ color: readableSecondaryTextColor, letterSpacing: 1.2, fontFamily: '"Noto Sans SC", monospace' }}
                >
                  {t('dashboard.nodeMap.unknownRegions')}
                </Typography>
                <Typography
                  variant="body2"
                  sx={{ color: readableTertiaryTextColor, lineHeight: 1.2, fontFamily: '"Noto Sans SC", monospace' }}
                >
                  {t('dashboard.nodeMap.unmatchedCoordinates')}
                </Typography>
              </Box>

              <Typography
                variant="h6"
                sx={{
                  color: unknownCount > 0 ? readablePrimaryTextColor : readableTertiaryTextColor,
                  fontWeight: 800,
                  lineHeight: 1,
                  fontFamily: '"Orbitron", sans-serif'
                }}
              >
                {String(unknownCount).padStart(2, '0')}
              </Typography>
            </Box>
          </Box>

          <Box sx={{ height: 2, width: 88, bgcolor: alpha(isDark ? '#63f2ff' : accentColor, isDark ? 0.9 : 0.8) }} />
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
            {t('dashboard.nodeMap.initializing')}
          </Typography>
        </Box>
      )}
    </Card>
  );
};

export default NodeMap;
