FROM golang:1.22-alpine AS build

ARG GOPRIVATE_REPO
ARG AZURE_CREDENTIAL

WORKDIR /app
RUN apk update && apk add --no-cache git
RUN export GOPRIVATE=${GOPRIVATE_REPO}

# Create nonroot user
RUN adduser --disabled-password -u 10001 appuser

# NETRC
RUN echo "machine dev.azure.com" >> ~/.netrc
RUN echo "login albatross-getnet" >> ~/.netrc
RUN echo "password ${AZURE_CREDENTIAL}" >> ~/.netrc

ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV CGO_ENABLED=0

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

WORKDIR /app/
RUN go build -ldflags="-w -s" -o /app/api cmd/main.go

##
## Deploy
FROM scratch

WORKDIR /app

COPY --from=build /app/api /app

EXPOSE 8443

# copy userus from build
COPY --from=build /etc/passwd /etc/passwd
USER appuser

ENTRYPOINT ["/app/api"]
