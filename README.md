# Bazooka CI end-to-end tests

This package contains end-to-end tests for the bazooka ci project.


## Prerequisites
### SCM servers
The tests in this project require starting SCM servers (git for the time being).

You need to build the git server docker image before being able to run the tests by running:

```
make scm
```

### Environment variables
The tests in this projet need 2 required environment variables and one optional:

* `BZK_E2E_TEMP`: **required** variable, needs to be set to a directory in the host machine which will be used as a temporary bazooka home for the tests
* `BZK_E2E_HOST`: **required** variable, needs to be set to the host machine's name or ip adress. The set value needs to be accessible from docker containers
* `BZK_E2E_DOCKER_SOCK`: **optional** variable, can be set to the location of the docker socket. Defaults to  `/var/run/docker.sock`

### Running

Simply run:

```
make test
```
