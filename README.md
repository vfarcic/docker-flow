```bash
go run src/*.go --help

docker-machine create --driver virtualbox docker-flow

docker-machine env docker-flow

eval "$(docker-machine env docker-flow)"

docker run -d \
    -p "8500:8500" \
    -h "consul" \
    progrium/consul -server -bootstrap

go run src/*.go

FLOW_HOST=$DOCKER_HOST go run *.go

go run \
    --host=$DOCKER_HOST \
    --consul-address=http://192.168.99.100:8500 \
    --project=books-ms \
    --target=app \
    --side-targets=db \
    --scale=1 \
    --blue-green

go build
```

https://github.com/vfarcic/books-ms/blob/master/Jenkinsfile
https://github.com/vfarcic/ms-lifecycle/blob/master/ansible/roles/jenkins/files/scripts/workflow-util.groovy

TODO
====

* Add tests
* Add to Travis/CircleCI
* Write README
* Explain the flow
* Roadmap
* Explain YAML
* Explain Environment variables
* Explain command line arguments
* Explain the order between YAML, env, and arguments
