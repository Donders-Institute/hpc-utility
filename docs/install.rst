.. install:

Build
=====

For building the source code, the [Go compiler](https://golang.org/) is required.

The ``Makefile`` is provided to simplify the dependancy resolution and code compilation.  To build the binaries, simply do:

.. code:: bash

    $ make

The executable binary (i.e. ``hpcutil``) will be built into your ``$GOPATH/bin`` directory.  The ``$GOPATH`` is by default ``$HOME/go`` if it is not set explicitly in your environment. 

Use the ``release`` target to build a RPM for the CentOS 7.x environment.  For example,

.. code:: bash

    $ make release

Use the ``github_release`` target with a proper ``VERSION`` variable to make a release on the GitHub repository.  For example,

.. code:: bash

    $ VERSION=1.0.0 make github_release

will creates a release tag ``1.0.0`` in the GitHub repository followed by build the release RPM for the CentOS 7.x environment. By the end of the process, you will be asked whether to upload the RPM as one of the release assets.