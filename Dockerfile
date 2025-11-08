FROM golang:1.24-alpine AS base
WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=bind,source=go.mod,target=go.mod \
  go mod download -x

FROM base AS builder
RUN --mount=type=cache,target=/go/pkg/mod/ \
 --mount=type=bind,target=. \
  go build -o /bin/server .

FROM scratch AS runner
COPY --from=builder /bin/server /bin/
ENTRYPOINT [ "/bin/server" ]