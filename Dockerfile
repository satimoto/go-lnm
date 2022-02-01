FROM golang:1.16-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o build/satimotod cmd/satimotod/main.go

FROM scratch
WORKDIR /app
COPY --from=build /build/satimotod .
CMD [ "./satimotod" ]