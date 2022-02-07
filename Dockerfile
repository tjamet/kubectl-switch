FROM golang:1.17 as build

RUN mkdir /src
WORKDIR /src

COPY go.* /src/
RUN go mod download


COPY pkg /src/pkg
COPY cmd /src/cmd

RUN go test ./...

# Compile binary with no dependencies
RUN CGO_ENABLED=0  go build -o /usr/bin/kubectl-switch ./cmd/kubectl-switch

FROM gcr.io/distroless/static:nonroot
COPY --from=build /usr/bin/kubectl-switch /usr/bin/kubectl-switch
ENTRYPOINT [ "/usr/bin/kubectl-switch" ]
# the plain docker image is actually /home/nonroot/.kube
LABEL io.whalebrew.config.volumes '["~/.kube:/.kube"]'
LABEL io.whalebrew.config.volumes_from_args '["--kubeconfig"]'
