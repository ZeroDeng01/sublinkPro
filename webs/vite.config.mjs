import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import jsconfigPaths from 'vite-jsconfig-paths';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig(({ mode }) => {
  // depending on your application, base can also be "/"
  const env = loadEnv(mode, process.cwd(), '');
  const API_URL = `${env.VITE_APP_BASE_NAME}`;
  const PORT = 3000;

  return {
    server: {
      // this ensures that the browser opens upon server start
      open: true,
      // this sets a default port to 3000
      port: PORT,
      host: true,
      // API proxy configuration
      proxy: {
        '/api': {
          target: env.VITE_API_URL || 'http://localhost:8080',
          changeOrigin: true,
          secure: false
        }
      }
    },
    build: {
      chunkSizeWarningLimit: 1600
    },
    preview: {
      open: true,
      host: true
    },
    define: {
      global: 'window'
    },
    resolve: {
      alias: {
        '@tabler/icons-react': '@tabler/icons-react/dist/esm/icons/index.mjs'
      }
    },
    base: API_URL,
    plugins: [
      react(),
      jsconfigPaths(),
      VitePWA({
        registerType: 'autoUpdate',
        includeAssets: ['images/favicon.ico', 'images/favicon.svg'],
        manifest: {
          name: 'Sublink Pro',
          short_name: 'Sublink',
          description: '支持更多协议和高度自定义的开源订阅转换管理工具',
          theme_color: '#2296f3',
          background_color: '#ffffff',
          display: 'fullscreen',
          start_url: '/',
          icons: [
            {
              src: 'icon-192.png',
              sizes: '192x192',
              type: 'image/png'
            },
            {
              src: 'icon-512.png',
              sizes: '512x512',
              type: 'image/png'
            },
            {
              src: 'icon-512.png',
              sizes: '512x512',
              type: 'image/png',
              purpose: 'maskable'
            }
          ]
        },
        workbox: {
          // 增加文件缓存大小限制到 3MB
          maximumFileSizeToCacheInBytes: 3 * 1024 * 1024,
          // 缓存静态资源（排除 html，因为后端会动态注入配置）
          globPatterns: ['**/*.{js,css,ico,png,svg,woff,woff2}'],
          // 导航请求使用 NetworkFirst 策略
          // 这样每次都会尝试从服务器获取最新的 index.html（带配置注入）
          navigateFallback: null, // 禁用默认的导航回退
          runtimeCaching: [
            {
              // 导航请求使用 NetworkFirst 策略
              urlPattern: ({ request }) => request.mode === 'navigate',
              handler: 'NetworkFirst',
              options: {
                cacheName: 'pages-cache',
                networkTimeoutSeconds: 3,
                expiration: {
                  maxEntries: 10,
                  maxAgeSeconds: 60 * 60 // 1小时
                }
              }
            },
            {
              // 缓存 Google Fonts
              urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
              handler: 'CacheFirst',
              options: {
                cacheName: 'google-fonts-cache',
                expiration: {
                  maxEntries: 10,
                  maxAgeSeconds: 60 * 60 * 24 * 365 // 1年
                }
              }
            },
            {
              // 缓存 Google Fonts 资源
              urlPattern: /^https:\/\/fonts\.gstatic\.com\/.*/i,
              handler: 'CacheFirst',
              options: {
                cacheName: 'gstatic-fonts-cache',
                expiration: {
                  maxEntries: 10,
                  maxAgeSeconds: 60 * 60 * 24 * 365 // 1年
                }
              }
            }
          ]
        }
      })
    ]
  };
});
