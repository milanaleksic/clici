# Jenkins Ping

[![Build Status](https://semaphoreci.com/api/v1/milanaleksic/jenkins_ping/branches/master/badge.svg)](https://semaphoreci.com/milanaleksic/jenkins_ping)

Command line for Jenkins pipeline overview

![Current state](current_state.png "Current state")

## How to run

I moved entire configuration from CLI switches into a `TOML` file.

In case you are starting application for the first time, execute `jenkins_ping -make-default-config`
which will generate a TOML file (it uses mock source instead of Jenkins server so you can experiment a bit).

## How to develop

This is a `golang` 1.6 project

`go get github.com/milanaleksic/jenkins_ping` should be enough to get the code and build.

In case you want to mimic my workflow you should use the `Makefile`:

    # get all 3rd party tools
    make prepare
    # build & test
    make test
