FROM --platform=$BUILDPLATFORM golang:1.27 AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -tags timetzdata -o /out/quizon ./cmd

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/quizon /quizon

EXPOSE 8000

ENTRYPOINT ["/quizon"]
