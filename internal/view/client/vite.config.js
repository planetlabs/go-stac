import {defineConfig} from 'vite';

export default defineConfig({
  root: 'src',
  plugins: [spaFallbackWithDot()],
  build: {
    outDir: '../dist',
    emptyOutDir: true,
  },
  esbuild: {
    legalComments: 'none',
  },
});

/**
 * Vite doesn't handle fallback html with dot (.), see https://github.com/vitejs/vite/issues/2415
 * @returns {import('vite').Plugin}
 */
function spaFallbackWithDot() {
  return {
    name: 'spa-fallback-with-dot',
    configureServer(server) {
      return () => {
        server.middlewares.use(function customSpaFallback(req, res, next) {
          if (req.url.includes('.') && !req.url.endsWith('.html')) {
            req.url = '/index.html';
          }
          next();
        });
      };
    },
  };
}
