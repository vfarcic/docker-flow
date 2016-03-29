TODO: Change the example to three VMs (swarm-master/proxy, swarm-node-1, swarm-node-2)

```bash
go test --cover
```

Setup
=====

```bash
docker-machine create \
    -d virtualbox \
    docker-flow

eval "$(docker-machine env docker-flow)"

docker run -d \
    -p "8500:8500" \
    -h "consul" \
    --name "consul" \
    progrium/consul -server -bootstrap
```

Provisioning
============

```bash
# The first time

export CONSUL_IP=$(docker-machine ip docker-flow)

export FLOW_CONSUL_ADDRESS=http://$CONSUL_IP:8500

docker ps -a --filter name=docker-flow-proxy

./docker-flow \
    --proxy-host $DOCKER_HOST \
    --proxy-docker-host $DOCKER_HOST \
    --proxy-docker-cert-path $DOCKER_CERT_PATH \
    --flow proxy

docker ps -a --filter name=docker-flow-proxy

# When proxy is stopped

docker stop docker-flow-proxy

docker ps -a --filter name=docker-flow-proxy

export FLOW_PROXY_HOST=$(docker-machine ip docker-flow)

export FLOW_PROXY_DOCKER_HOST=$DOCKER_HOST

export FLOW_PROXY_DOCKER_CERT_PATH=$DOCKER_CERT_PATH

./docker-flow --flow proxy

docker ps -a --filter name=docker-flow-proxy

# When proxy is removed

docker rm -f docker-flow-proxy

docker ps -a --filter name=docker-flow-proxy

./docker-flow --flow proxy

docker ps -a --filter name=docker-flow-proxy
```

Reconfiguring Proxy After Deployment
====================================

```bash
./docker-flow --flow deploy --flow proxy

./docker-flow --scale +1 --flow scale

./docker-flow --flow deploy

# Run integration tests

./docker-flow --flow proxy
```
