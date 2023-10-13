FROM --platform=$BUILDPLATFORM golang AS build

ARG GIT_DESC=undefined

WORKDIR /go/src/github.com/Snawoot/drng
COPY . .
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -a -tags netgo -trimpath -asmflags -trimpath -ldflags '-s -w -extldflags "-static" -X main.version='"$GIT_DESC" ./cmd/drng

FROM scratch
COPY --from=build /go/src/github.com/Snawoot/drng/drng /
USER 9999:9999
ENTRYPOINT ["/drng"]
