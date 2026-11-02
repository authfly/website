# Build the SPA, then run the BFF (Bun) as a non-root user. No secrets in the image: pass them at run time.
FROM oven/bun:1 AS build
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bunx vite build

FROM oven/bun:1-slim
WORKDIR /app
ENV NODE_ENV=production PORT=3000
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/dist ./dist
COPY package.json server.ts ./
USER bun
EXPOSE 3000
CMD ["bun", "server.ts"]
