import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { "@": fileURLToPath(new URL(".", import.meta.url)) },
  },
  build: {
    rolldownOptions: {
      output: {
        codeSplitting: {
          // React's runtime is the largest and least frequently changing code, so give it its
          // own long-cacheable chunk. Deliberately not a catch-all node_modules group: that would
          // pull recharts back out of the lazily loaded user page chunk.
          groups: [{ name: "react", test: /node_modules[\\/](react|react-dom|scheduler)[\\/]/ }],
        },
      },
    },
  },
  server: {
    port: 20000,
    strictPort: true,
    // Mirrors the production Caddyfile: /api/* goes to the Go server with the prefix stripped.
    // Key must be "/api/" (not "/api"), or it would also swallow the /api.tsx module.
    proxy: {
      "/api/": {
        target: process.env.API_PROXY_TARGET ?? "http://localhost:20002",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});
