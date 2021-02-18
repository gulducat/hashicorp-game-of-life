# TODO: scratch https://github.com/hashicorp/checkpoint/pull/26/commits/87b8622c5ac53fa6bbc6ddc5a7b77af5429f9749
FROM golang:1.15-alpine as builder
WORKDIR /gol
COPY go.* ./
RUN go mod download
COPY *.go ./
# RUN go build .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-w -s"

# FROM alpine:latest
# RUN apk add curl
FROM alpine:latest
COPY --from=builder /gol/hashicorp-game-of-life /
ENV NOMAD_JOB_NAME=0-0
ENTRYPOINT ["/hashicorp-game-of-life"]
CMD ["run"]

# for debuggin
# RUN apk add curl
# ENTRYPOINT []
# CMD ["/hashicorp-game-of-life", "run"]
