FROM golang:1.21.3-alpine3.18 as builder

ARG GITCOMMIT=docker
ARG GITDATE=docker
ARG GITVERSION=docker

RUN apk add make jq git gcc musl-dev linux-headers

COPY ./starknet-proxyd /app

WORKDIR /app

RUN make starknet-proxyd

FROM alpine:3.18

RUN apk add bind-tools jq curl bash git redis

COPY ./starknet-proxyd/entrypoint.sh /bin/entrypoint.sh

RUN apk update && \
    apk add ca-certificates && \
    chmod +x /bin/entrypoint.sh

EXPOSE 8080

VOLUME /etc/starknet-proxyd

COPY --from=builder /app/bin/starknet-proxyd /bin/starknet-proxyd

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD ["/bin/starknet-proxyd", "/etc/starknet-proxyd/starknet-proxyd.toml"]
