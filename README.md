# foreman-builder

A mac tool to help build foreman dev env.
(Currently WIP)

This tool will help setup a container based foreman dev environment.

## Requirements
[Go](https://go.dev/doc/install)
[Orbstack](https://orbstack.dev/)

## Getting started
First install foreman-builder
```bash
go install github.com/aidenfine/foreman-builder@latest
```

**Create a container**
To create a dev container we can run the `create` command. This will bring you through the process and setup of the container which will be installed. Once the container is finished you will be able to ssh into the dev environment. Any ports that are opened on the container will be able to be visited via `localhost`
```bash
foreman-builder create
```

**List containers**
Will display a list of containers that `foreman-builder` has created.
```bash
foreman-builder list
```



