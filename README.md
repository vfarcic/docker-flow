Docker Flow
===========

*Docker Flow* is a project aimed towards creating an easy to use continuous deployment flow. It uses [Docker Engine](https://www.docker.com/products/docker-engine), [Docker Compose](https://www.docker.com/products/docker-compose), and [Consul](https://www.consul.io/).

For information regarding features and motivations behind this project, please read the [Docker Flow: Blue-Green Deployment and Relative Scaling](http://technologyconversations.com/2016/03/07/docker-flow-blue-green-deployment-and-relative-scaling/) article.

The current list of features is as follows.

* Blue-green deployment
* Relative scaling

The latest release can be found [here](https://github.com/vfarcic/docker-flow/releases/latest).

Arguments
---------

Arguments can be specified through *docker-flow.yml* file, environment variables, and command line arguments. If the same argument is specified in several places, command line overwrites all others and environment variables overwrite *docker-flow.yml*.

### Command line arguments

|Command argument       |Environment variable  |YML              |Description|
|-----------------------|----------------------|-----------------|-----------|
|-F, --flow             |FLOW                  |flow             |The actions that should be performed as the flow. (**multi**)<br>**deploy**: Deploys a new release<br>**scale**: Scales currently running release<br>**stop-old**: Stops the old release<br>**proxy**: Reconfigures the proxy<br>(**default**: [deploy])|
|-H, --host             |FLOW_HOST             |host             |Docker daemon socket to connect to. If not specified, DOCKER_HOST environment variable will be used instead.|
|-f, --compose-path     |FLOW_COMPOSE_PATH     |compose_path     |Path to the Docker Compose configuration file. (**default**: docker-compose.yml)|
|-b, --blue-green       |FLOW_BLUE_GREEN       |blue_green       |Perform blue-green deployment. (**bool**)|
|-t, --target           |FLOW_TARGET           |target           |Docker Compose target. (**required**)|
|-T, --side-target      |FLOW_SIDE_TARGETS     |side_targets     |Side or auxiliary Docker Compose targets. (**multi**)|
|-P, --skip-pull-targets|FLOW_SKIP_PULL_TARGET |skip_pull_target |Skip pulling targets. (**bool**)|
|-S, --pull-side-targets|FLOW_PULL_SIDE_TARGETS|pull_side_targets|Pull side or auxiliary targets. (**bool**)|
|-p, --project          |FLOW_PROJECT          |project          |Docker Compose project. If not specified, the current directory will be used instead.|
|-c, --consul-address   |FLOW_CONSUL_ADDRESS   |consul_address   |The address of the Consul server. (**required**)|
|-s, --scale            |FLOW_SCALE            |scale            |Number of instances to deploy. If the value starts with the plus sign (+), the number of instances will be increased by the given number. If the value begins with the minus sign (-), the number of instances will be decreased by the given number.|
|-r, --proxy-host       |FLOW_PROXY_HOST       |proxy_host       |Docker daemon socket of the proxy host.|

Arguments can be strings, boolean, or multiple values. Command line arguments of boolean type do not have any value (i.e. *--blue-green*). Environment variables and YML arguments of boolean type should use *true* as value (i.e. *FLOW_BLUE_GREEN=true* and *blue_green: true*). When allowed, multiple values can be specified by repeating the command line argument (i.e. *--flow=deploy --flow=stop-old*). When specified through environment variables, multiple values should be separated by comma (i.e. *FLOW=deploy,stop-old*). YML accepts multiple values through the standard format.

```yml
flow:
  - deploy
  - stop-old
```

Examples
--------

Examples can be found in the [Docker Flow: Blue-Green Deployment and Relative Scaling](http://technologyconversations.com/2016/03/07/docker-flow-blue-green-deployment-and-relative-scaling/) article.
