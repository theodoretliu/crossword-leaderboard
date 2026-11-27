import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react({ jsxImportSource: "@emotion/react" })],
  resolve: {
    alias: { "@": fileURLToPath(new URL(".", import.meta.url)) },
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
