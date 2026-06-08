FROM ghcr.io/brandonkowalski/quasimodo:latest

WORKDIR /build

COPY go.mod go.sum* ./

RUN GOWORK=off go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN GOWORK=off go build -v \
    -tags nodefaultfont \
    -ldflags "-X github.com/aaronearles/nextui-hak-store/version.Version=${VERSION} \
              -X github.com/aaronearles/nextui-hak-store/version.GitCommit=${GIT_COMMIT} \
              -X github.com/aaronearles/nextui-hak-store/version.BuildDate=${BUILD_DATE}" \
    -o hak-store ./app

CMD ["/bin/bash"]
