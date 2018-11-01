# kubectl-switch

A wrapper over kubectl to ensure using a version that matches the server version.

# Installation

```bash
go get -u github.com/tjamet/kubectl-switch
go build -o /usr/local/bin/k github.com/tjamet/kubectl-switch
```

# Usage

```bash
k version
k get ns
```
