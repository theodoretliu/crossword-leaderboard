// Same-origin in every environment: Caddy (production) and the Vite dev server
// both proxy /api/* to the Go server with the /api prefix stripped.
export const API_URL = import.meta.env.VITE_API_URL ?? "/api";
