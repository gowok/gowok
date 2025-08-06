FROM golang:1.24.5-alpine AS build-stage

WORKDIR /app
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/cli ./cli/cli.go

# --- Using Multi-Stage Builds ---
FROM alpine:3.13

WORKDIR /app

COPY --from=build-stage /app/bin/cli .
COPY entrypoint.sh .

RUN chmod +x entrypoint.sh

# --- Using Environment Variables in the Final Stage ---
ARG DEFAULT_PORT=8080
ARG DEFAULT_CONFIG_FILE="config.yaml"

# Use ENV to set a runtime environment variable in the final image.
ENV PORT=${DEFAULT_PORT}
ENV CONFIG=${DEFAULT_CONFIG_FILE}

EXPOSE ${PORT}
CMD ["./entrypoint.sh"]