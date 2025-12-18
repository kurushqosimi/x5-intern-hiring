# ---------- builder ----------
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN GOOS=linux go build -o /server ./cmd/app

# ---------- runtime ----------
FROM scratch

COPY --from=builder /server /bin/server

#COPY configs/config.yaml /bin/config/config.yaml

EXPOSE 8080

ENTRYPOINT ["/bin/server"]
