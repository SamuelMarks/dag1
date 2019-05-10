.. _install:

Install
=======

From Source
^^^^^^^^^^^

Clone the `repository <https://github.com/SamuelMarks/dag1>`__ in the appropriate GOPATH subdirectory:

::

    $ mkdir -p $GOPATH/src/github.com/SamuelMarks/
    $ cd $GOPATH/src/github.com/SamuelMarks
    [...]/SamuelMarks$ git clone https://github.com/SamuelMarks/dag1.git


The easiest way to build binaries is to do so in a hermetic Docker container.
Use this simple command:

::

	[...]/dag1$ make dist

This will launch the build in a Docker container and write all the artifacts in
the build/ folder.

::

    [...]/dag1$ tree build
    build/
    ├── dist
    │   ├── dag1_0.1.0_darwin_386.zip
    │   ├── dag1_0.1.0_darwin_amd64.zip
    │   ├── dag1_0.1.0_freebsd_386.zip
    │   ├── dag1_0.1.0_freebsd_arm.zip
    │   ├── dag1_0.1.0_linux_386.zip
    │   ├── dag1_0.1.0_linux_amd64.zip
    │   ├── dag1_0.1.0_linux_arm.zip
    │   ├── dag1_0.1.0_SHA256SUMS
    │   ├── dag1_0.1.0_windows_386.zip
    │   └── dag1_0.1.0_windows_amd64.zip
    └── pkg
        ├── darwin_386
        │   └── dag1
        ├── darwin_amd64
        │   └── dag1
        ├── freebsd_386
        │   └── dag1
        ├── freebsd_arm
        │   └── dag1
        ├── linux_386
        │   └── dag1
        ├── linux_amd64
        │   └── dag1
        ├── linux_arm
        │   └── dag1
        ├── windows_386
        │   └── dag1.exe
        └── windows_amd64
            └── dag1.exe

Go Devs
^^^^^^^

DAG1 is written in `Golang <https://golang.org/>`__. Hence, the first step is
to install **Go version 1.9 or above** which is both the programming language
and a CLI tool for managing Go code. Go is very opinionated  and will require
you to `define a workspace <https://golang.org/doc/code.html#Workspaces>`__
where all your go code will reside.

Dependencies
^^^^^^^^^^^^

DAG1 uses `Glide <http://github.com/Masterminds/glide>`__ to manage
dependencies. For Ubuntu users:

::

    [...]/dag1$ curl https://glide.sh/get | sh
    [...]/dag1$ glide install

This will download all dependencies and put them in the **vendor** folder.

Testing
^^^^^^^

DAG1 has extensive unit-testing. Use the Go tool to run tests:

::

    [...]/dag1$ make test

If everything goes well, it should output something along these lines:

::

    ?       github.com/SamuelMarks/dag1/src/dag1     [no test files]
    ok      github.com/SamuelMarks/dag1/src/common     0.015s
    ok      github.com/SamuelMarks/dag1/src/crypto     0.122s
    ok      github.com/SamuelMarks/dag1/src/poset  10.270s
    ?       github.com/SamuelMarks/dag1/src/mobile     [no test files]
    ok      github.com/SamuelMarks/dag1/src/net        0.012s
    ok      github.com/SamuelMarks/dag1/src/node       19.171s
    ok      github.com/SamuelMarks/dag1/src/peers      0.038s
    ?       github.com/SamuelMarks/dag1/src/proxy      [no test files]
    ok      github.com/SamuelMarks/dag1/src/proxy/dummy        0.013s
    ok      github.com/SamuelMarks/dag1/src/proxy/inmem        0.037s
    ok      github.com/SamuelMarks/dag1/src/proxy/socket       0.009s
    ?       github.com/SamuelMarks/dag1/src/proxy/socket/app   [no test files]
    ?       github.com/SamuelMarks/dag1/src/proxy/socket/dag1        [no test files]
    ?       github.com/SamuelMarks/dag1/src/service    [no test files]
    ?       github.com/SamuelMarks/dag1/src/version    [no test files]
    ?       github.com/SamuelMarks/dag1/cmd/dag1     [no test files]
    ?       github.com/SamuelMarks/dag1/cmd/dag1/commands    [no test files]
    ?       github.com/SamuelMarks/dag1/cmd/dummy      [no test files]
    ?       github.com/SamuelMarks/dag1/cmd/dummy/commands     [no test files]
