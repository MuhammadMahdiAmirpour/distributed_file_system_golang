FROM docker.arvancloud.ir/golang:1.23.1 AS build

RUN mkdir /app/distributed-file-system -p
RUN chmod +rwx /app/distributed-file-system
WORKDIR /app/distributed-file-system

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/fs driver/main.go

FROM docker.arvancloud.ir/alpine:3.18

RUN mkdir /app/distributed-file-system -p
RUN chmod +rwx /app/distributed-file-system

WORKDIR /app/distributed-file-system

COPY --from=build /app/distributed-file-system/bin/fs /app/distributed-file-system/fs

RUN chmod +x /app/distributed-file-system/fs

EXPOSE 3000 7000 5000

CMD ["/app/distributed-file-system/fs"]
