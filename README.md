Docker Flow
===========

Since the first time I laid my hand on Docker, I started writing my own scripts that I've been running as my continuous deployment flow. I ended up with Shell scripts, Ansible playbooks, Chef cookbooks, Jenkins Pipelines, and so on. Each of those had a similar (not to say the same) objective inside a different context. I realized that was a huge waste of time and decided to create a single executable that I'll be able to run no matter the tool I'm using to execute the continuous deployment pipeline. The result is a birth of the [Docker Flow](https://github.com/vfarcic/docker-flow) project.

Features
========

The goal of the project is to add features and processes that are currently missing inside the Docker ecosystem. The project is in its infancy and solves only the problems of *blue-green deployments* and *relative scaling*. Many additional features will be added very soon.

I'll restrain myself from explaining *blue-green deployment* since I already wrote quite a few articles on this subject. If you are not already familiar with it, please read the [Blue-Green Deployment](http://technologyconversations.com/2016/02/08/blue-green-deployment/) article. For a more hands-on example with Jenkins Pipeline, please read the [Blue-Green Deployment To Docker Swarm with Jenkins Workflow Plugin](http://technologyconversations.com/2015/12/08/blue-green-deployment-to-docker-swarm-with-jenkins-workflow-plugin/) post. Finally, for another practical example in a much broader context (and without Jenkins), please consult the [Scaling To Infinity with Docker Swarm, Docker Compose and Consul](http://technologyconversations.com/2015/07/02/scaling-to-infinity-with-docker-swarm-docker-compose-and-consul-part-14-a-taste-of-what-is-to-come/) series.

The second feature I wanted to present is relative scaling. Docker Compose makes it very easy to scale services to a fixed number. We can specify how many instances of a container we want to run and watch it do the magic. When combined with Docker Swarm, the result is an easy containers management inside a cluster. Depending on how many are already running, Docker Compose will increase (or decrease) the number of running containers so that the desired number of instances is running. However, Docker Compose always expect a fixed number as parameter. I found this very limiting when dealing with production deployments. In many cases, I do not want to know how many instances are already running but, simply, send a signal to increase (or decrease) the capacity by some factor. For example, we might have an increase in traffic and want to increase the capacity by three instances. Similarly, if the demand for some service decreases, I want the number of running instances to decrease as well thus freeing resources for other services and processes. This necessity is even more obvious when we move towards autonomous and automated [Self-Healing Systems](http://technologyconversations.com/2016/01/26/self-healing-systems/) where human interactions are reduced to a minimum.

Running Docker Flow
===================

**Docker Flow** requirements are [Docker Engine](https://www.docker.com/products/docker-engine), [Docker Compose](https://www.docker.com/products/docker-compose), and [Consul](https://www.consul.io/). The idea behind the project is not to substitute any Docker functionality but to provide additional features. It assumes that containers are defined in a docker-compose.yml file (path can be changed) and that Consul is used as service registry (soon to be extended to etcd and Zookeeper as well).

The following examples will setup and environment with Docker Engine and Consul, and will assume that you already the [Docker Toolbox](https://www.docker.com/products/docker-toolbox) installed. Even though the examples will be run through the Docker Toolbox Terminal, feel free to apply them to your existing setup. You can run them on any of your Docker servers or, even better, inside a Docker Swarm cluster.

Let's start setting up the environment we'll use throughout this article. Please launch the *Docker Toolbox Terminal*, clone the project repository, and download the latest *Docker Flow* release.

```bash
git clone https://github.com/vfarcic/docker-flow

cd docker-flow

# Please download the latest release from [Docker Flow Releases](https://github.com/vfarcic/docker-flow/releases/latest) and rename it to docker-flow

wget https://github.com/vfarcic/docker-flow/releases/latest/docker-flow_darwin_amd64 \
    -o docker-flow

# Explain docker-compose.yml

docker-machine create --driver virtualbox docker-flow

eval "$(docker-machine env docker-flow)"

docker run -d \
    -p "8500:8500" \
    -h "consul" \
    --name "consul" \
    progrium/consul -server -bootstrap

export CONSUL_IP=$(docker-machine ip docker-flow)

curl $CONSUL_IP:8500/v1/status/leader
```

Deployment With Downtime
------------------------

```bash
BOOKS_MS_VERSION=:1.0 docker-compose up -d app db

export BOOKS_MS_VERSION=:latest

docker-compose pull app

docker-compose up -d app db

docker-compose down
```

Blue-Green Deployment
---------------------

```bash
./docker-flow \
    --consul-address=http://$CONSUL_IP:8500 \
    --target=app \
    --side-target=db \
    --blue-green

docker ps -a

# books-ms-db is running
# booksms_app-blue_1 is running

export CONSUL_ADDRESS=http://$CONSUL_IP:8500

./docker-flow

docker ps -a

# books-ms-db is running
# booksms_app-blue_1 is stopped
# booksms_app-green_1 is running
```

Relative Scaling
----------------

```bash
./docker-flow --scale=2

docker ps -a

./docker-flow --scale=+4

docker ps -a

# books-ms-db is running
# booksms_app-green_[1-6] are running
# booksms_app-blue_[1-2] is stopped

./docker-flow --scale=-3

docker ps -a

# books-ms-db is running
# booksms_app-green_[1-6] are stopped
# booksms_app-blue_[1-3] are running
```

TODO
====

* Explain why previous color is stopped and not removed
* Mention Swarm
* Explain YAML
* Explain Environment variables
  * Multiple values in FLOW_SIDE_TARGETS should be separated by comma (,)
* Explain command line arguments
  * side-target can be specified multiple times
* Explain the order between YAML, env, and arguments
* Create a release
* Write an article
* Switch from docker-compose stop/rm combination to down