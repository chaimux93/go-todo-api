FROM golang:1.24.0-alpine AS builder
RUN apk add --no-cache upx
WORKDIR /app
COPY go.* ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o main -a --trimpath --ldflags="-s -w" -installsuffix cgo .
RUN upx --ultra-brute -qq main && upx -t main
FROM scratch
COPY --from=builder /app/main /main
CMD ["/main"]
EXPOSE 8080
