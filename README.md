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

|Command argument                     |Description|
|-------------------------------------|-----------|
|-H, --host=                          |Docker daemon socket to connect to. If not specified, DOCKER_HOST environment variable will be used instead.|
|    --cert-path=                     |Docker certification path. If not specified, DOCKER_CERT_PATH environment variable will be used instead.|
|-f, --compose-path=docker-compose.yml|Path to the Docker Compose configuration file. (default: docker-compose.yml)|
|-b, --blue-green                     |Perform blue-green deployment. (**bool**)|
|-t, --target=                        |Docker Compose target. (default: app)|
|-T, --side-target=                   |Side or auxiliary Docker Compose targets. Multiple values are allowed. (default: [db]) (**multi**)|
|-S, --pull-side-targets              |Pull side or auxiliary targets. (**bool**)|
|-p, --project=                       |Docker Compose project. If not specified, the current directory will be used instead.|
|-c, --consul-address=                |The address of the Consul server.|
|-s, --scale=                         |Number of instances to deploy. If the value starts with the plus sign (+), the number of instances will be increased by the given number. If the value begins with the minus sign (-), the number of instances will be decreased by the given number.|
|-F, --flow=                          |The actions that should be performed as the flow. Multiple values are allowed.<br>**deploy**: Deploys a new release<br>**scale**: Scales currently running release<br>**stop-old**: Stops the old release<br>**proxy**: Reconfigures the proxy<br>(default: [deploy]) (**multi**)|
|    --proxy-host=                    |The host of the proxy. Visitors should request services from this domain. Docker Flow uses it to request reconfiguration when a new service is deployed or an existing one is scaled. This argument is required only if the proxy flow step is used.|
|    --proxy-docker-host=             |Docker daemon socket of the proxy host. This argument is required only if the proxy flow step is used.|
|    --proxy-docker-cert-path=        |Docker certification path for the proxy host.|
|    --proxy-reconf-port=             |The port used by the proxy to reconfigure its configuration|
|    --service-path=                  |Path that should be configured in the proxy (e.g. /api/v1/my-service). This argument is required only if the proxy flow step is used. (**multi**)|
|-h, --help                           |Show this help message|

### Mappings from command line arguments to YML and environment variables

|Command argument                     |YML                   |Environment variable       |
|-------------------------------------|----------------------|---------------------------|
|-H, --host=                          |host                  |FLOW_HOST or DOCKER_HOST   |
|    --cert-path=                     |cert_path             |FLOW_CERT_PATH             |
|-f, --compose-path=docker-compose.yml|compose_path          |FLOW_COMPOSE_PATH          |
|-b, --blue-green                     |blue_green            |FLOW_BLUE_GREEN            |
|-t, --target=                        |target                |FLOW_TARGET                |
|-T, --side-target=                   |side_targets          |FLOW_SIDE_TARGETS          |
|-S, --pull-side-targets              |pull_side_targets     |FLOW_PULL_SIDE_TARGETS     |
|-p, --project=                       |project               |FLOW_PROJECT               |
|-c, --consul-address=                |consul_address        |FLOW_CONSUL_ADDRESS        |
|-s, --scale=                         |scale                 |SCALE                      |
|-F, --flow=                          |flow                  |FLOW                       |
|    --proxy-host=                    |proxy_host            |FLOW_PROXY_HOST            |
|    --proxy-docker-host=             |proxy_docker_host|FLOW_PROXY_DOCKER_HOST          |
|    --proxy-docker-cert-path=        |proxy_docker_cert_path|FLOW_PROXY_DOCKER_CERT_PATH|
|    --proxy-reconf-port=             |proxy_reconf_port     |FLOW_PROXY_RECONF_PORT     |
|    --service-path=                  |service_path          |FLOW_SERVICE_PATH          |

Arguments can be strings, boolean, or multiple values. Command line arguments of boolean type do not have any value (i.e. *--blue-green*). Environment variables and YML arguments of boolean type should use *true* as value (i.e. *FLOW_BLUE_GREEN=true* and *blue_green: true*). When allowed, multiple values can be specified by repeating the command line argument (i.e. *--flow=deploy --flow=stop-old*). When specified through environment variables, multiple values should be separated by comma (i.e. *FLOW=deploy,stop-old*). YML accepts multiple values through the standard format.

```yml
flow:
  - deploy
  - stop-old
```

Examples
--------

Examples can be found in the [Docker Flow: Blue-Green Deployment and Relative Scaling](http://technologyconversations.com/2016/03/07/docker-flow-blue-green-deployment-and-relative-scaling/) article.
