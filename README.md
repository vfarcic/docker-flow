```bash
go run src/*.go --help

docker-machine create --driver virtualbox docker-flow

docker-machine env docker-flow

eval "$(docker-machine env docker-flow)"

go run src/*.go --host=$DOCKER_HOST --project=books-ms --target=app --side-targets=db --scale=1

go build src/technologyconversations.com/docker-flow/docker-flow.go
```

https://github.com/vfarcic/books-ms/blob/master/Jenkinsfile
https://github.com/vfarcic/ms-lifecycle/blob/master/ansible/roles/jenkins/files/scripts/workflow-util.groovy