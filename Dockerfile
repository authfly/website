# Build from repo root:
#   docker build -t authfly/website:ru .
# Or compose from e:/ with adjusted context — prefer building inside this repo now
# that Framework is a public Go module.

FROM oven/bun:1.2-alpine AS css
WORKDIR /src
COPY package.json bun.lock* ./
RUN bun install
COPY fastygo.config.mjs ./
COPY scripts ./scripts
COPY web/static/css ./web/static/css
COPY internal/views ./internal/views
COPY internal/ui ./internal/ui
COPY internal/kit/ui ./internal/kit/ui
RUN bun run build:css

FROM golang:1.25-bookworm AS build
WORKDIR /src
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=css /src/web/static/css/app.css ./web/static/css/app.css
RUN templ generate ./... \
  && CGO_ENABLED=0 GOOS=linux go build -o /out/authflyweb ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/authflyweb /app/authflyweb
COPY --from=build /src/web/static /app/web/static
COPY --from=build /src/internal/adapters/fixture/data /app/internal/adapters/fixture/data
COPY --from=build /src/internal/adapters/fixture/data-en /app/internal/adapters/fixture/data-en
ENV APP_BIND=0.0.0.0:8080 \
    APP_STATIC_DIR=/app/web/static \
    APP_DEFAULT_LOCALE=ru \
    FIXTURE_DIR=/app/internal/adapters/fixture/data
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/authflyweb"]
