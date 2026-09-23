FROM --platform=linux/amd64 alpine:3.20
ENV DEBIAN_FRONTEND noninteractive
RUN apk add --no-cache --no-check-certificate ca-certificates curl tzdata mediainfo && update-ca-certificates
ENV TZ=Asia/Ho_Chi_Minh
WORKDIR /app
COPY ffmpeg-release-amd64-static.tar.xz /app
RUN mkdir -p /usr/share && tar -xf ffmpeg-release-amd64-static.tar.xz && mv ffmpeg-7.0.2-amd64-static /usr/share/ffmpeg && ln -s /usr/share/ffmpeg/ffmpeg /usr/local/bin/ffmpeg && ln -s /usr/share/ffmpeg/ffprobe /usr/local/bin/ffprobe && rm /app/ffmpeg-release-amd64-static.tar.xz
COPY .env.prod .env
COPY bin/server /app/server
EXPOSE 8000
CMD ["/app/server"]