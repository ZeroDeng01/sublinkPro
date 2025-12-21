import { useEffect, useRef, useMemo } from 'react';
import * as echarts from 'echarts';
// import 'echarts-gl'; // 3D Not needed for flat map
// import { useTheme } from '@mui/material/styles'; // Unused
import { Box, Card, Typography, CircularProgress } from '@mui/material';
import worldJson from 'assets/world.json';
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
  // const theme = useTheme(); // Unused
  const chartInstance = useRef(null);

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
          color: '#fbbf24' // Gold
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
          color: '#0cecdb' // Cyan
        }
      });

      // Flying Lines to China
      if (chinaCoords) {
        lns.push({
          coords: [coords, chinaCoords],
          lineStyle: {
            color: '#38bdf8' // Light Blue
          }
        });
      }
    });

    return { points: pts, lines: lns, targetPoint: tPoint, unknownCount: unk, maxCount: max };
  }, [data]);

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
        backgroundColor: 'rgba(15, 23, 42, 0.95)',
        borderColor: '#0cecdb',
        borderWidth: 1,
        textStyle: {
          color: '#fff',
          fontFamily: '"Noto Sans SC", sans-serif'
        },
        padding: [12, 16],
        formatter: (params) => {
          // const code = params.value && (params.value[3] || params.value[2]); // Unused
          // In 2D scatter: value is [lon, lat, count, code] -> index 3
          // In Target: value is [lon, lat, 0] -> no code? target has name 'CN'

          if (params.seriesType === 'lines') return ''; // No tooltip for lines

          // Target check
          if (params.name === 'CN' || (params.data && params.data.name === 'CN')) {
            return `<div style="display:flex;align-items:center;gap:12px">
                             <span style="font-size:24px;line-height:1">🇨🇳</span>
                             <span style="font-size:16px;font-weight:bold;color:#fff">CN (本地区域)</span>
                          </div>`;
          }

          const valCode = params.value && params.value[3];
          if (!valCode) return '';

          const count = params.value[2];
          const flagEmoji = getFlagEmoji(valCode);

          return `<div style="display:flex;align-items:center;gap:12px">
                    <span style="font-size:24px;line-height:1">${flagEmoji}</span>
                    <span style="font-size:16px;font-weight:bold;color:#fff">${valCode}</span>
                 </div>
                 <div style="margin-top:8px;font-size:12px;opacity:0.9;display:flex;justify-content:space-between;width:120px">
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
            areaColor: '#1e293b', // slate-800
            borderColor: '#0f172a', // slate-900
            borderWidth: 1.5
          },
          emphasis: {
            areaColor: '#334155' // slate-700
          }
        },
        regions: [
          {
            name: 'China',
            itemStyle: {
              areaColor: '#334155' // Highlight China slightly base map
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
            color: '#fff',
            symbolSize: 3
          },
          lineStyle: {
            normal: {
              color: '#0cecdb',
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
              color: '#0cecdb',
              width: 1,
              opacity: 0.4,
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
            scale: 3
          },
          label: {
            show: true,
            position: 'right',
            formatter: (params) => {
              const code = params.value[3];
              return getFlagEmoji(code);
            },
            fontSize: 14,
            distance: 5
          },
          symbolSize: (val) => Math.max(6, Math.min(20, Math.log2(val[2] + 1) * 5)),
          itemStyle: {
            color: '#0cecdb'
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
                  color: '#fbbf24'
                },
                symbol: 'pin',
                symbolSize: 20,
                itemStyle: {
                  color: '#fbbf24'
                },
                label: {
                  show: true,
                  formatter: '🇨🇳 CN',
                  position: 'top',
                  fontWeight: 'bold',
                  color: '#fbbf24',
                  fontSize: 14,
                  backgroundColor: 'rgba(0,0,0,0.5)',
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
  }, [points, lines, targetPoint, loading]);

  return (
    <Card
      sx={{
        height: '100%',
        width: '100%',
        background: 'radial-gradient(circle at center, #0f172a 0%, #020617 100%)', // Simpler gradient for 2D map
        color: '#fff',
        position: 'relative',
        overflow: 'hidden',
        boxShadow: 'none',
        border: 'none',
        borderRadius: 0
      }}
    >
      {/* Holographic Grid Overlay - Kept for aesthetics */}
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          opacity: 0.1,
          backgroundImage: `
                linear-gradient(rgba(14, 165, 233, 0.3) 1px, transparent 1px),
                linear-gradient(90deg, rgba(14, 165, 233, 0.3) 1px, transparent 1px)
            `,
          backgroundSize: '40px 40px',
          pointerEvents: 'none'
        }}
      />

      <Box sx={{ p: 4, position: 'absolute', top: 0, left: 0, zIndex: 10 }}>
        <Typography
          variant="h3"
          sx={{
            color: '#fff',
            fontWeight: 800,
            textShadow: '0 0 30px rgba(14, 165, 233, 0.8)',
            letterSpacing: '4px',
            mb: 1,
            fontFamily: '"Orbitron", sans-serif'
          }}
        >
          全球节点分布
        </Typography>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: '#22c55e', boxShadow: '0 0 10px #22c55e' }} />
          <Typography variant="subtitle2" sx={{ color: '#0ea5e9', letterSpacing: '2px', fontFamily: '"Noto Sans SC", monospace' }}>
            以获取到落地IP的数据为参考
          </Typography>
        </Box>

        {unknownCount > 0 && (
          <Box
            sx={{
              mt: 3,
              display: 'inline-flex',
              alignItems: 'center',
              gap: 2,
              bgcolor: 'rgba(15, 23, 42, 0.6)',
              py: 1.5,
              px: 3,
              borderRadius: 0,
              borderLeft: '4px solid #0ea5e9',
              backdropFilter: 'blur(4px)'
            }}
          >
            <Typography variant="body2" sx={{ color: '#94a3b8', fontFamily: '"Noto Sans SC", monospace', fontSize: '0.9rem' }}>
              未知区域 &gt;&gt; <span style={{ color: '#fff', fontWeight: 'bold' }}>{unknownCount}</span>
            </Typography>
          </Box>
        )}
      </Box>

      {/* Stats overlay */}
      <Box sx={{ p: 6, position: 'absolute', bottom: 0, right: 0, zIndex: 10, textAlign: 'right', pointerEvents: 'none' }}>
        <Typography
          variant="overline"
          sx={{ color: 'rgba(255,255,255,0.7)', letterSpacing: 2, display: 'block', fontFamily: '"Noto Sans SC"' }}
        >
          覆盖区域
        </Typography>
        <Typography
          variant="h1"
          sx={{ color: '#0ea5e9', fontWeight: '900', fontSize: '4rem', lineHeight: 1, textShadow: '0 0 40px rgba(14, 165, 233, 0.6)' }}
        >
          {String(points.length + (targetPoint ? 1 : 0)).padStart(2, '0')}
        </Typography>

        <Box sx={{ mt: 2, height: 2, width: 100, bgcolor: '#0ea5e9', ml: 'auto', opacity: 0.8 }} />
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
          <CircularProgress size={60} sx={{ color: '#0ea5e9' }} />
          <Typography sx={{ color: '#0ea5e9', letterSpacing: 4, fontFamily: '"Noto Sans SC", monospace' }}>正在初始化地图...</Typography>
        </Box>
      )}
    </Card>
  );
};

export default NodeMap;
