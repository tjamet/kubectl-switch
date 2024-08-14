# kubectl-switch

⚠️ This project is deprecated and will be turned into a read-only mode at the begining of October 2024.
Other similar project like [kuberlr] now exist and are packaged with Kubernetes distributions.

In particular [kuberlr] offers a more flexible interface to allow listing and managing installed kubectl versions.

```plaintext
Usage:
  kuberlr [command]

Available Commands:
  bins        Print information about the kubectl binaries found
  completion  Generate the autocompletion script for the specified shell
  get         Download the kubectl version specified
  help        Help about any command
  version     Print version information

Flags:
  -h, --help              help for kuberlr
  -v, --verbosity Level   log level [0-5]. 0 (Only Error and Warning) to 5 (Maximum detail).

Use "kuberlr [command] --help" for more information about a command.
```

kubectl-switch: A wrapper over kubectl to ensure using a version that matches the server version.

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
[kuberlr]: https://github.com/flavio/kuberlr
