# 빌드 스테이지
FROM golang:1.14-alpine as builder

# 빌드 패키지 설정
RUN apk add git build-base && \
	go get -u github.com/swaggo/swag/cmd/swag && \
	mkdir /build
# 빌드 환경 설정
WORKDIR /build
COPY ./ /build/

# swagger 생성, Unit Test, 바이너리 빌드
RUN swag init && \
	go test -v && \
	go build

# 실제 이미지 생성
FROM alpine:3.12

# 실행 환경 구성
RUN mkdir /app
WORKDIR /app
COPY --from=builder /build/skhuinfo-api /app/

ENTRYPOINT ["./skhuinfo-api"]
