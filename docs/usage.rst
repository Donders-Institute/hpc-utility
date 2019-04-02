.. usage:

Usage
=====

The single CLI tool provided by this package is called ``hpcutil``. The usage of it is similar to the CLI of ``git`` where a group of (sub-)sub-commands are provided in a hierarchical manner.

The inline help for a subcommand and associated flags (options) of it is always available via the ``-h`` option.  It also supports tab-completion in the Bash shell which means that one could resolve the available subcommands or flags by pressing the TAB key twice.

Currently, the CLI provides two main subcommands on the first level: ``cluster`` and ``webhook``.

The ``cluster`` subcommand
--------------------------

The ``cluster`` subcommand can be used to retrieve job or resource (node, software license) information of the HPC cluster.  To get the in-terminal help of the ``cluster`` subcommand, one uses the following command in the terminal:

.. code:: bash

    $ hpcutil cluster -h

where ``-h`` is optional.

It shows another level of subcommands that are available, for instance:

.. code:: bash

    Available Commands:
        config      Print Torque and Moab server configurations.
        job         Retrieve information about a cluster job.
        matlablic   Print a summary of the Matlab license usage.
        nodes       Retrieve information about cluster nodes.
        qstat       Print job list in the memory of the Torque server.

One can then take one from those available command to move onto another level of the sub-commands.  For example, if one wants to get nodes resource information, one does

.. code:: bash

    $ hpcutil cluster nodes

and via the supported subcommands shown below, you will be able to get resource information such as the disk/memory usage on a list of cluster nodes, or a summary table of the running VNC sessions on the cluster. 

.. code:: bash

    Available Commands:
        diskfree    Print total and free disk space of the cluster nodes.
        memfree     Print total and free memory on the cluster nodes.
        vnc         Print list of VNC servers on the cluster or a specific node.

Example: list MATLAB licenses allocated by DCCN users
*****************************************************

.. code:: bash

    $ hpcutil cluster matlablic

Example: list VNC sessions
**************************

To get all VNC sessions running on the cluster's access nodes, one can do:

.. code:: bash

    $ hpcutil cluster nodes vnc

To get VNC sessions on a given host (e.g. ``mentat001.dccn.nl``), one does:

.. code:: bash

    $ hpcutil cluster nodes vnc mentat001.dccn.nl

To get VNC sessions owned by a given user (e.g. ``honlee``), one does:

.. code:: bash

    $ hpcutil cluster nodes vnc -u honlee

One could combine the last two examples to find VNC sessions owned by a user on a specific host.  For example, the following command will find VNC sessions owned by user ``honlee`` on host ``mentat001.dccn.nl``.

.. code:: bash

    $ hpcutil cluster nodes vnc -u honlee mentat001.dccn.nl


Example: show all cluster jobs
******************************

.. code:: bash

    $ hpcutil cluster qstat

Example: check memory utilization of a running job
**************************************************

Assuming a running job with ID ``1234567``, the owner of the job can perform the following command to check the memory usage in real time:

.. code:: bash

    $ hpcutil cluster job meminfo 1234567

Example: get job's trace log
****************************

Assuming a job with ID ``1234567``, the job trace log can be obtained from the Torque server via the following command:

.. code:: bash

    $ hpcutil cluster job trace 1234567

The ``webhook`` subcommand
--------------------------

The ``webhook`` subcommand is used to manage the webhook facility of the HPC cluster.  To get the in-terminal help of the ``webhook`` subcommand, one uses the following command in the terminal:

.. code:: bash

    $ hpcutil webhook -h

where ``-h`` is optional.