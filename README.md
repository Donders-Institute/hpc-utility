# hpc-cluster-tools
High Performance Computing (HPC) cluster tools.

## build the code

To build the code, you will need the go compiler.

You also need to decide on the `$GOPATH` variable in which the dependancies, binaries and libraries are going to stored during and after the build. Assuming we use `$HOME/go` as the `$GOPATH`, we build the code as the following command:

```bash
$ GOPATH=$HOME/go make
```

After the build, the binaries will be located under various subdirectories in the `$GOPATH`.
