FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# modernc.org/sqlite is pure Go, so CGO_ENABLED=0 still works and gives a
# fully static binary that runs on a minimal distroless base.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/book-scrape .
# an empty directory to become /data, so it can be owned by the nonroot user
RUN mkdir /out/data

FROM gcr.io/distroless/static-debian12:nonroot
# 65532 is the nonroot user in this image. A new named volume mounted at /data
# copies this ownership, so sqlite can create its file there.
COPY --from=build --chown=65532:65532 /out/data /data
WORKDIR /data
COPY --from=build /out/book-scrape /usr/local/bin/book-scrape
VOLUME ["/data"]
ENV DB_PATH=/data/db.db
ENTRYPOINT ["/usr/local/bin/book-scrape"]
