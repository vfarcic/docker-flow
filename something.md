```bash
export GOPATH=~/go/

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
```

https://github.com/vfarcic/books-ms/blob/master/Jenkinsfile
https://github.com/vfarcic/ms-lifecycle/blob/master/ansible/roles/jenkins/files/scripts/workflow-util.groovy
