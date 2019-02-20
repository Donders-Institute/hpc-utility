# hpc-cluster-tools

CLI tools for the High Performance Computing (HPC) cluster.

## Clone and build the code

To build the code, you will need the [GO compiler](https://golang.org).

Following the "GO concept", identify the `$GOPATH` on your system under which various GO source codes, libraries and binaries will be stored (and organized).

Assuming we use `$HOME/projects/go` as the `$GOPATH`, we will then clone the code of this repository as the follows:

```bash
$ export GOPATH=$HOME/projects/go
$ mkdir -p $GOPATH/src/github.com
$ cd $GOPATH/src/github.com
$ git clone https://github.com/Donders-Institute/hpc-cluster-tools.git
$ cd hpc-cluster-tools
```

We then build the code with the following command:

```bash
$ GOPATH=$HOME/projects/go make
```

After the build, the binaries will be located under various subdirectories in the `$GOPATH`, for example, the executables are in `$GOPATH/bin` and library files are in `$GOPATH/pkg`.

## Contribute to the code

### Folder structure of the repository

The repository holds a GO project.  The folders are organized accordingly the [GO standard](https://github.com/golang-standards/project-layout).

### The `Makefile`

The [Makefile](Makefile) is created to simplify the build process.  In it, we make use of the [go dep](https://golang.github.io/dep/) to manage the library dependancies.

