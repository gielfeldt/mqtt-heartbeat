FROM --platform=${BUILDPLATFORM} golang:1.14.3-alpine AS base
WORKDIR /src
ENV CGO_ENABLED=0

FROM base AS dev
RUN apk update && \
    apk add git

RUN go get -v github.com/mitranim/gow

CMD [ "gow", "run", "/src" ]

FROM base AS build
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/heartbeat .

FROM scratch AS prod
COPY --from=build /out/heartbeat /

CMD [ "/heartbeat" ]