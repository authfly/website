# Build from workspace root (parent of @_FlightPHP and @_Go):
#   docker build -f @_FlightPHP/@AuthFlySite/Dockerfile -t authfly/website:ru .
# Or: docker compose -f @_FlightPHP/@AuthFlySite/docker-compose.yml up --build

FROM oven/bun:1.2-alpine AS css
WORKDIR /src
COPY @_FlightPHP/@AuthFlySite/package.json @_FlightPHP/@AuthFlySite/bun.lock* ./
RUN bun install
COPY @_FlightPHP/@AuthFlySite/fastygo.config.mjs ./
COPY @_FlightPHP/@AuthFlySite/scripts ./scripts
COPY @_FlightPHP/@AuthFlySite/web/static/css ./web/static/css
COPY @_FlightPHP/@AuthFlySite/internal/views ./internal/views
COPY @_FlightPHP/@AuthFlySite/internal/ui ./internal/ui
COPY @_FlightPHP/@AuthFlySite/internal/kit/ui ./internal/kit/ui
RUN bun run build:css

FROM golang:1.25-bookworm AS build
WORKDIR /src
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
COPY @_FlightPHP/@AuthFlySite/ ./
COPY @_Go/@Framework/ /framework/
COPY --from=css /src/web/static/css/app.css ./web/static/css/app.css
RUN sed -i 's|replace github.com/fastygo/framework => .*|replace github.com/fastygo/framework => /framework|' go.mod \
  && go mod tidy \
  && templ generate ./... \
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
