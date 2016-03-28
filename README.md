# clici

> Command Line Interface for Continuous Integration

[![Build Status](https://semaphoreci.com/api/v1/milanaleksic/clici/branches/master/badge.svg)](https://semaphoreci.com/milanaleksic/clici)

Command line for Jenkins pipeline overview, so you don't have to use heavy weight "pipeline overview / wall plugins"

NOTE: This application was previously called *jenkins_ping* and was renamed to **clici** on 28th March 2016 (because clici sounds better).

![Current state](current_state.png "Current state")

## How to run

Program takes almost no command line parameters at all and uses a configuration `TOML` file.

In case you are starting application for the first time, execute `clici -make-default-config`
which will generate a TOML file (it uses mock source instead of Jenkins server so you can experiment a bit).

## How to develop

This is a `golang` 1.6 project

`go get github.com/milanaleksic/clici` should be enough to get the code and build.

In case you want to mimic my workflow you should use the `Makefile`:

    # get all 3rd party tools
    make prepare
    # build & test
    make test
