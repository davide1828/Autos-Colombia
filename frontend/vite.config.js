import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

// En desarrollo, las llamadas /api se redirigen al backend Go.
export default defineConfig({
  plugins: [react()],
  server: { proxy: { "/api": "http://localhost:8080" } },
  test: { environment: "node", include: ["src/**/*.test.js"] },
});
