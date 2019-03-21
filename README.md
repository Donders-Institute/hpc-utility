# hpc-cluster-tools

CLI tools for the High Performance Computing (HPC) cluster.

## Clone and build the code

To build the code, you will need the [GO compiler](https://golang.org).

Following the "GO concept", identify the `$GOPATH` on your system under which various GO source codes, libraries and binaries will be stored (and organized).

Assuming we use `$HOME/projects/go` as the `$GOPATH`, we will then clone the code of this repository as the follows:

```bash
$ export GOPATH=$HOME/projects/go
$ mkdir -p $GOPATH/src/github.com/Donders-Institute
$ cd $GOPATH/src/github.com/Donders-Institute
$ git clone https://github.com/Donders-Institute/hpc-cluster-tools.git
$ cd hpc-cluster-tools
```

We then build the code with the following command:

```bash
$ make
```

After the build, the binaries will be located under various subdirectories in the `$GOPATH`, for example, the executables are in `$GOPATH/bin` and library files are in `$GOPATH/pkg`.

One could also build RPM on a CentOS 7.x using the following command:

```bash
$ make release
```

or make a GitHub release with the RPM as the release asset:

```bash
$ VERSION={RELEASE_NUMBER} make github_release
```

where the `{RELEASE_NUMBER}` is the new release number to be created on this repository's [release page](https://github.com/Donders-Institute/hpc-cluster-tools/releases). It cannot be an existing release number.

## Contribute to the code

### Folder structure of the repository

The repository holds a GO project.  The folders are organized according to the [GO standard](https://github.com/golang-standards/project-layout). The rationale of this style of organization is explained [in this post](https://medium.com/golang-learn/go-project-layout-e5213cdcfaa2).

### The `Makefile`

The [Makefile](Makefile) is created to simplify the build process.  In it, we make use of the [go dep](https://golang.github.io/dep/) to manage the library dependancies.

