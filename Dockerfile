# Self-host the docs site: docker build -t authfly-website . && docker run -p 3011:3011 authfly-website
# Generated templ code and assets/app.css are committed, so the image needs only the Go toolchain.
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o /out/site ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/site /site
ENV PORT=3011
EXPOSE 3011
USER nonroot:nonroot
ENTRYPOINT ["/site"]
