import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The Go API runs on localhost:8080. We proxy /api so the browser talks to the
// Vite origin only: this keeps the HttpOnly refresh_token cookie same-origin
// and avoids CORS setup on the backend.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
