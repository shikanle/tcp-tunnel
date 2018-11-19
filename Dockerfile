FROM iron/go:dev
LABEL maintainer="Kanle Shi <shikanle@gmail.com>"
ENV PROJECT_DIR /go/src/github.com/shikanle/tcp-tunnel
ADD . ${PROJECT_DIR}
WORKDIR ${PROJECT_DIR}
RUN go build -o /go/bin/tcp-tunnel .

FROM iron/go
LABEL maintainer="Kanle Shi <shikanle@gmail.com>"
ENV PUBLISH_PORT 80
ENV TUNNEL_PORT 7000
COPY --from=0 /go/bin/tcp-tunnel /app/
WORKDIR /app
ENTRYPOINT ./tcp-tunnel -mode server -publish ${PUBLISH_PORT} -tunnel ${TUNNEL_PORT}
