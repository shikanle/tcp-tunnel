FROM iron/go:dev
LABEL maintainer="Kanle Shi <shikanle@gmail.com>"
ENV PROJECT_DIR /go/src/github.com/shikanle/tcp-tunnel
ADD . ${PROJECT_DIR}
WORKDIR ${PROJECT_DIR}
RUN go build -o /go/bin/tcp-tunnel .

FROM iron/go
LABEL maintainer="Kanle Shi <shikanle@gmail.com>"
ENV MODE server
ENV PUBLISH_PORT 80
ENV TUNNEL_PORT 7000
ENV SERVER_URI localhost:7000
ENV LOCAL_URI localhost:80
ENV POOL_SIZE 32
COPY --from=0 /go/bin/tcp-tunnel /app/
WORKDIR /app
ENTRYPOINT ./tcp-tunnel -mode ${MODE} -publish ${PUBLISH_PORT} -tunnel ${TUNNEL_PORT} -server {SERVER_URI} -local {LOCAL_URI} -pool ${POOL_SIZE}
