```bash
go test --cover
```

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

export CONSUL_IP=$(docker-machine ip docker-flow)

export FLOW_CONSUL_ADDRESS=http://$CONSUL_IP:8500

# The first time

./docker-flow \
    --proxy-host $DOCKER_HOST \
    --proxy-cert-path $DOCKER_CERT_PATH \
    --flow proxy

docker ps -a --filter name=docker-flow-proxy

# When proxy is stopped

docker stop docker-flow-proxy

docker ps -a --filter name=docker-flow-proxy

./docker-flow \
    --proxy-host $DOCKER_HOST \
    --proxy-cert-path $DOCKER_CERT_PATH \
    --flow proxy

docker ps -a --filter name=docker-flow-proxy

# When proxy is removed

docker rm -f docker-flow-proxy

docker ps -a --filter name=docker-flow-proxy

./docker-flow \
    --proxy-host $DOCKER_HOST \
    --proxy-cert-path $DOCKER_CERT_PATH \
    --flow proxy

docker ps -a --filter name=docker-flow-proxy











./docker-flow \
    --proxy-host $DOCKER_HOST \
    --proxy-cert-path $DOCKER_CERT_PATH \
    --flow deploy \
    --flow proxy \
    --flow stop-old
```