FROM golang:1.24.2-alpine3.21 as builder

WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o website-backend cmd/server/main.go

FROM scratch as final
COPY --from=builder /app/website-backend .
EXPOSE 8080

CMD ["./website-backend"]