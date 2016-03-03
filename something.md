```bash
export GOPATH=~/go/

go run src/*.go --help

docker-machine create --driver virtualbox docker-flow

docker-machine env docker-flow

eval "$(docker-machine env docker-flow)"

go build

go test --cover -v

go test -coverprofile=coverage.out dockerflow

go tool cover --html=coverage.out

docker run -d \
    -p "8500:8500" \
    -h "consul" \
    progrium/consul -server -bootstrap

go build && ./docker-flow

go build && FLOW_HOST=$DOCKER_HOST ./docker-flow

# Make sure that --consul-address is correct
go build && FLOW_HOST=$DOCKER_HOST ./docker-flow \
    --host=$DOCKER_HOST \
    --consul-address=http://192.168.99.100:8500 \
    --project=books-ms \
    --target=app \
    --side-targets=db \
    --scale=1 \
    --blue-green
```

https://github.com/vfarcic/books-ms/blob/master/Jenkinsfile
https://github.com/vfarcic/ms-lifecycle/blob/master/ansible/roles/jenkins/files/scripts/workflow-util.groovy

TODO
====

* Write README
  * Explain the flow
  * Roadmap
  * Explain YAML
  * Explain Environment variables
    * Multiple values in FLOW_SIDE_TARGETS should be separated by comma (,)
  * Explain command line arguments
    * side-target can be specified multiple times
  * Explain the order between YAML, env, and arguments
* Create a release
* Write an article
