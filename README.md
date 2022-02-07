# kubectl-switch

A wrapper over kubectl to ensure using a version that matches the server version.

# Installation

## From releases

Visit the [latest release page](https://github.com/tjamet/kubectl-switch/releases/latest) of this repo and download the binary that matches your system.

From a command-line, run

```bash
curl -L -o /usr/local/bin/k https://github.com/tjamet/kubectl-switch/releases/download/v1.1/kubectl-switch-$(uname -s)-$(uname -m)
chmod +x /usr/local/bin/k
```

## From sources

```bash
go get -u github.com/tjamet/kubectl-switch
go build -o /usr/local/bin/k github.com/tjamet/kubectl-switch
```

# Usage

```bash
k version
k get ns
```
