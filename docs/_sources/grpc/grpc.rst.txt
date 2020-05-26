
Diggind in to the gRPC example code
==========================================

Installing gRPC
--------------------------------------------------------

see the `quickstart <https://grpc.io/docs/quickstart/go/>`_ for more detail

First you need to install the gRPC module for go

.. code-block:: shell 
        :caption: Install Go Module for gRPC

        $ export GO111MODULE=on # this forces go to use Go modules instead of GOPATH
        $ go get google.golang.org/grpc@v1.28.1


The **GO111MODULE** situation is a bit of mess, this article explains it well
https://dev.to/maelvls/why-is-go111module-everywhere-and-everything-about-go-modules-24k

Then install the `protoc <https://github.com/protocolbuffers/protobuf>`_ compiler
If you're on Arch Linux there is a package for it 

.. code-block:: shell
        :caption: Install protobuf from AUR

        $ sudo pacman -Syu protobuf

Go Import System Sidebar
________________________________________________________

**OK...** Coming from a python & kotlin background go's import system is
a mess to deal with and has already caused me too much greif.

when I try to run the example code given in the `quickstart`_ I get this

.. code-block:: shell
        :caption: I really wish GOROOT was a Guardians of the Galaxy reference

        greeter_server/main.go:30:2: package examples/helloworld/helloworld is not in GOROOT (/usr/lib/go/src/examples/helloworld/helloworld)

Clearly I've been spoiled by pythons import system

------

:Everything needs to be in the $GOPATH:
        I thought this was my problem but it didn't fix the issue

------

:**SUCCESS**:
        In Go,  every program must be part of a package. I did initialize a package, but I
        did so one directory above where I should have. After running

.. code-block:: shell
        :caption: It's always something simple isn't it, That's 1/2 hour I won't get back

        $ go mod init

So for future, **not in the GOROOT** means I probably didn't initalize the package

        



