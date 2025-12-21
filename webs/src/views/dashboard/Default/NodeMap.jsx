import { useEffect, useRef, useMemo } from 'react';
import * as echarts from 'echarts';
import 'echarts-gl';
import { useTheme } from '@mui/material/styles';
import { Box, Card, Typography, CircularProgress } from '@mui/material';

// Common country coordinates (Longitude, Latitude)
const countryCoordinates = {
  CN: [104.1954, 35.8617], // China
  China: [104.1954, 35.8617],
  US: [-95.7129, 37.0902], // USA
  'United States': [-95.7129, 37.0902],
  JP: [138.2529, 36.2048], // Japan
  Japan: [138.2529, 36.2048],
  HK: [114.1694, 22.3193], // Hong Kong
  'Hong Kong': [114.1694, 22.3193],
  TW: [120.9605, 23.6978], // Taiwan
  Taiwan: [120.9605, 23.6978],
  SG: [103.8198, 1.3521], // Singapore
  Singapore: [103.8198, 1.3521],
  KR: [127.7669, 35.9078], // South Korea
  'South Korea': [127.7669, 35.9078],
  GB: [-3.436, 55.3781], // UK
  'United Kingdom': [-3.436, 55.3781],
  DE: [10.4515, 51.1657], // Germany
  Germany: [10.4515, 51.1657],
  FR: [2.2137, 46.2276], // France
  France: [2.2137, 46.2276],
  CA: [-106.3468, 56.1304], // Canada
  Canada: [-106.3468, 56.1304],
  AU: [133.7751, -25.2744], // Australia
  Australia: [133.7751, -25.2744],
  RU: [105.3188, 61.524], // Russia
  Russia: [105.3188, 61.524],
  IN: [78.9629, 20.5937], // India
  India: [78.9629, 20.5937],
  BR: [-51.9253, -14.235], // Brazil
  Brazil: [-51.9253, -14.235],
  NL: [5.2913, 52.1326], // Netherlands
  Netherlands: [5.2913, 52.1326]
  // Add more as needed
};

const NodeMap = ({ data = {}, loading = false }) => {
  const chartRef = useRef(null);
  const theme = useTheme();
  // We force dark mode styles for the map to look "Sci-Fi" even if the app is in light mode,
  // or we blend it. Generally 3D Sci-fi looks best on dark.
  const isDark = theme.palette.mode === 'dark';
  const chartInstance = useRef(null);

  // Filter valid data and separate unknown useMemo to avoid recalc and dependency issues
  const { points, lines, unknownCount } = useMemo(() => {
    const pts = [];
    const lns = [];
    let unk = 0;
    const chinaCoords = countryCoordinates['CN'];

    Object.entries(data).forEach(([country, count]) => {
      // Try to find coords by key or mapping
      let coords = countryCoordinates[country];

      // If not found and it's a known alias map (simple logic)
      if (!coords) {
        // Maybe handle case-insensitive?
        const key = Object.keys(countryCoordinates).find((k) => k.toLowerCase() === country.toLowerCase());
        if (key) coords = countryCoordinates[key];
      }

      if (coords) {
        pts.push({
          name: country,
          value: [...coords, count], // [lon, lat, count]
          itemStyle: {
            color: country === 'CN' || country === 'China' ? '#ff3333' : '#00ffd0'
          },
          label: {
            show: count > 0,
            formatter: '{b}: {c}',
            position: 'top',
            textStyle: {
              color: '#fff',
              fontSize: 12
            }
          }
        });

        // Add line to China if it's not China
        if (coords[0] !== chinaCoords[0] || coords[1] !== chinaCoords[1]) {
          lns.push({
            coords: [coords, chinaCoords],
            lineStyle: {
              color: '#00ffd0',
              opacity: 0.4
            }
          });
        }
      } else {
        unk += count;
      }
    });

    return { points: pts, lines: lns, unknownCount: unk };
  }, [data]);

  useEffect(() => {
    if (!chartRef.current || loading) return;

    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    const option = {
      backgroundColor: 'transparent',
      tooltip: {
        show: true,
        trigger: 'item',
        backgroundColor: 'rgba(0,0,0,0.7)',
        borderColor: '#00ffd0',
        textStyle: {
          color: '#fff'
        }
      },
      globe: {
        baseTexture: null, // Use environment color or texture
        // Using a gradient environment for sci-fi look
        environment: '#000000',

        heightTexture: null,
        displacementScale: 0.1,

        shading: 'lambert',

        // Sphere material
        globeRadius: 80,
        globeOuterRadius: 100,

        itemStyle: {
          color: '#1a1a2e', // Dark blue/black earth
          borderColor: '#4d5057',
          borderWidth: 1
        },

        // Atmosphere
        atmosphere: {
          show: true,
          offset: 5,
          color: '#00ffd0',
          glowPower: 1,
          innerGlowPower: 1
        },

        viewControl: {
          autoRotate: true,
          autoRotateSpeed: 5,
          distance: 200,
          minDistance: 150,
          maxDistance: 400
        },

        light: {
          ambient: {
            intensity: 0.6
          },
          main: {
            intensity: 0.4,
            shadow: true
          }
        },

        postEffect: {
          enable: true,
          bloom: {
            enable: true,
            intensity: 0.2
          }
        }
      },
      series: [
        {
          type: 'scatter3D',
          coordinateSystem: 'globe',
          data: points,
          symbolSize: (val) => Math.max(5, Math.min(20, Math.log2(val[2] + 1) * 4)),
          itemStyle: {
            color: '#00ffd0',
            opacity: 1
          },
          emphasis: {
            label: {
              show: true
            }
          }
        },
        {
          type: 'lines3D',
          coordinateSystem: 'globe',
          effect: {
            show: true,
            period: 2,
            trailWidth: 3,
            trailLength: 0.5,
            trailOpacity: 1,
            trailColor: '#00ffd0'
          },
          blendMode: 'lighter',
          lineStyle: {
            width: 1,
            color: '#00ffd0',
            opacity: 0.1
          },
          data: lines
        }
      ]
    };

    chartInstance.current.setOption(option);

    const handleResize = () => {
      chartInstance.current?.resize();
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chartInstance.current?.dispose();
      chartInstance.current = null;
    };
  }, [points, lines, loading, isDark]);

  return (
    <Card
      sx={{
        height: '100%',
        minHeight: 500,
        background: 'linear-gradient(135deg, #0f172a 0%, #1e1b4b 100%)', // Deep space blue/purple
        color: '#fff',
        position: 'relative',
        overflow: 'hidden',
        border: '1px solid rgba(0, 255, 208, 0.2)',
        boxShadow: '0 0 20px rgba(0, 255, 208, 0.1)'
      }}
    >
      <Box sx={{ p: 3, position: 'absolute', top: 0, left: 0, zIndex: 10 }}>
        <Typography variant="h3" sx={{ color: '#00ffd0', textShadow: '0 0 10px rgba(0, 255, 208, 0.5)' }}>
          全球节点分布
        </Typography>
        <Typography variant="subtitle2" sx={{ color: 'rgba(255,255,255,0.7)' }}>
          Global Node Distribution
        </Typography>

        {unknownCount > 0 && (
          <Box sx={{ mt: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
            <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: 'rgba(255,255,255,0.5)' }} />
            <Typography variant="body2" sx={{ color: 'rgba(255,255,255,0.7)' }}>
              未知地区: {unknownCount}
            </Typography>
          </Box>
        )}
      </Box>

      <Box
        ref={chartRef}
        sx={{
          width: '100%',
          height: '100%',
          minHeight: 500,
          opacity: loading ? 0 : 1,
          transition: 'opacity 0.5s'
        }}
      />

      {loading && (
        <Box
          sx={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)'
          }}
        >
          <CircularProgress sx={{ color: '#00ffd0' }} />
        </Box>
      )}
    </Card>
  );
};

export default NodeMap;
