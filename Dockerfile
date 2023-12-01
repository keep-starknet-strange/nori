FROM golang:1.21.3-alpine3.18 as builder

ARG GITCOMMIT=docker
ARG GITDATE=docker
ARG GITVERSION=docker

RUN apk add make jq git gcc musl-dev linux-headers

COPY ./nori /app

WORKDIR /app

RUN make nori

FROM alpine:3.18

RUN apk add bind-tools jq curl bash git redis

COPY ./nori/entrypoint.sh /bin/entrypoint.sh

RUN apk update && \
    apk add ca-certificates && \
    chmod +x /bin/entrypoint.sh

EXPOSE 8080

VOLUME /etc/nori

COPY --from=builder /app/bin/nori /bin/nori

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD ["/bin/nori", "/etc/nori/nori.toml"]
