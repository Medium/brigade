FROM golang:1.10-alpine3.8 AS builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

RUN apk --no-cache add ca-certificates git

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/github.com/Medium/brigade
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /brigade .

FROM alpine:3.8
RUN apk --no-cache add ca-certificates
COPY --from=builder /brigade ./
ENTRYPOINT ["./brigade"]
