import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Dev: `vite` on :5173 proxies the auth routes to the BFF (bun server.ts on :3000).
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: { proxy: { "/login": "http://localhost:3000", "/callback": "http://localhost:3000", "/logout": "http://localhost:3000", "/api": "http://localhost:3000" } },
  build: { outDir: "dist" },
});
