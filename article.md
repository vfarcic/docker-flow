Since the first time I laid my hand on Docker, I started writing my own scripts that I've been running as my continuous deployment flow. I ended up with Shell scripts, Ansible playbooks, Chef cookbooks, Jenkins Pipelines, and so on. Each of those had a similar (not to say the same) objective inside a different context. I realized that was a huge waste of time and decided to create a single executable that I'll be able to run no matter the tool I'm using to execute the continuous deployment pipeline. The result is a birth of the [Docker Flow](https://github.com/vfarcic/docker-flow) project.

Features
========

The goal of the project is to add features and processes that are currently missing inside the Docker ecosystem. The project is in its infancy and solves only the problems of *blue-green deployments* and *relative scaling*. Many additional features will be added very soon.

I'll restrain myself from explaining *blue-green deployment* since I already wrote quite a few articles on this subject. If you are not familiar with it, please read the [Blue-Green Deployment](http://technologyconversations.com/2016/02/08/blue-green-deployment/) article. For a more hands-on example with Jenkins Pipeline, please read the [Blue-Green Deployment To Docker Swarm with Jenkins Workflow Plugin](http://technologyconversations.com/2015/12/08/blue-green-deployment-to-docker-swarm-with-jenkins-workflow-plugin/) post. Finally, for another practical example in a much broader context (and without Jenkins), please consult the [Scaling To Infinity with Docker Swarm, Docker Compose and Consul](http://technologyconversations.com/2015/07/02/scaling-to-infinity-with-docker-swarm-docker-compose-and-consul-part-14-a-taste-of-what-is-to-come/) series.

The second feature I wanted to present is relative scaling. Docker Compose makes it very easy to scale services to a fixed number. We can specify how many instances of a container we want to run and watch it do the magic. When combined with Docker Swarm, the result is an easy containers management inside a cluster. Depending on how many are already running, Docker Compose will increase (or decrease) the number of running containers so that the desired number of instances is running. However, Docker Compose always expect a fixed number as parameter. I found this very limiting when dealing with production deployments. In many cases, I do not want to know how many instances are already running but, simply, send a signal to increase (or decrease) the capacity by some factor. For example, we might have an increase in traffic and want to increase the capacity by three instances. Similarly, if the demand for some service decreases, I want the number of running instances to decrease as well thus freeing resources for other services and processes. This necessity is even more obvious when we move towards autonomous and automated [Self-Healing Systems](http://technologyconversations.com/2016/01/26/self-healing-systems/) where human interactions are reduced to a minimum.

Running Docker Flow
===================

**Docker Flow** requirements are [Docker Engine](https://www.docker.com/products/docker-engine), [Docker Compose](https://www.docker.com/products/docker-compose), and [Consul](https://www.consul.io/). The idea behind the project is not to substitute any Docker functionality but to provide additional features. It assumes that containers are defined in a docker-compose.yml file (path can be changed) and that Consul is used as service registry (soon to be extended to etcd and Zookeeper as well).

The following examples will setup and environment with Docker Engine and Consul, and will assume that you already the [Docker Toolbox](https://www.docker.com/products/docker-toolbox) installed. Even though the examples will be run through the Docker Toolbox Terminal, feel free to apply them to your existing setup. You can run them on any of your Docker servers or, even better, inside a Docker Swarm cluster.

Let's start setting up the environment we'll use throughout this article. Please launch the *Docker Toolbox Terminal* and clone the project repository

```bash
git clone https://github.com/vfarcic/docker-flow

cd docker-flow
```

The next step is to download the latest *Docker Flow* release. Please open the [Docker Flow Releases](https://github.com/vfarcic/docker-flow/releases/latest) page, download the release that maches your OS, rename it to docker-flow, and make sure that it has execute permissions (i.e. `chmod +x docker-flow`). The rest of the article will assume that the binary is in the *docker-flow* directory created when we cloned the repository. If you choose to use it, please place it to one of the directories included in your *PATH*.

Before we proceed, let's take a look at the *docker-compose.yml* file we'll use.

```yml
version: '2'

services:
  app:
    image: vfarcic/books-ms${BOOKS_MS_VERSION}
    ports:
      - 8080
    environment:
      - SERVICE_NAME=books-ms
      - DB_HOST=books-ms-db

  db:
    container_name: books-ms-db
    image: mongo
    environment:
      - SERVICE_NAME=books-ms-db
```

As you can see, it does not contain anything special. It defines two targets. The *app* target defines the main container of the server. The *db* is a "side" target required by *app*. Since the version is *2*, Docker Compose will utilize one of its new features and create a network around those targets allowing containers to communicate with each others (handy if deployed to a cluster). Finally, the *app* image uses the *BOOKS_MS_VERSION* environment variable that will allow us to simulated multiple releases. I assume that you already used Docker Compose and that there is no reason to go into more details.

We'll use *docker-machine* to create a VM that will simulate our production environment.

```bash
docker-machine create \
    --driver virtualbox docker-flow

eval "$(docker-machine env docker-flow)"
```

*Docker Flow* needs to store the state of the containers it is deploying, so I made a choice to use *Consul* in this first iteration. The other registries (*etcd*, *Zookeeper*, and whichever other someone requests) will follow soon. Let's bring up an instance.

```bash
docker run -d \
    -p "8500:8500" \
    -h "consul" \
    --name "consul" \
    progrium/consul -server -bootstrap

export CONSUL_IP=$(docker-machine ip docker-flow)
```

Now we are ready to start deploying the *books-ms* service.

Deployment With Downtime
------------------------

Before we dive into *Docker Flow* features, let's see how does the deployment work with Docker Compose.

We'll start by deploying the release *1.0*.

```bash
BOOKS_MS_VERSION=:1.0 docker-compose up -d app db

docker-compose ps
```

The output of the `docker-compose ps` command is as follows.

```
      Name                Command          State            Ports
--------------------------------------------------------------------------
books-ms-db        /entrypoint.sh mongod   Up      27017/tcp
dockerflow_app_1   /run.sh                 Up      0.0.0.0:32771->8080/tcp
```

We haven't done anything special (for now). The two containers (*mongo* and *books-ms* version 1.0) are running.

Let's take a look what happens if we deploy a new release.

```
export BOOKS_MS_VERSION=:latest

docker-compose up -d app db
```

The output of the `docker-compose up` command is as follows.

```
Recreating dockerflow_app_1
books-ms-db is up-to-date
```

The problem lies in the first line that states that the *dockerflow_app_1* is being recreated. There are quite a few problem that might arise from it, two most important ones is downtime and inability to test the new release before making it available in production. Not only that your service will be inavailable during a (hopefully) short period of time, but it will be untested as well. In this particular case, I do not want to say that the service is not tested at all, but that the deployment has not been tested. Even if you run a round of integration testing in a staging environment, there is no guarantee that the same tests would pass in production. For that reason, a preferable way to handle this scenario is to apply *blue-green deployment* process. We should run the new release in parallel with the old one, run a set of tests that confirm that the deployment was done correctly and that the service is integrated with the rest of the system, and, if everything went as expected, switch the proxy from the old to the new release. With this process, we avoid downtime and, at the same time, guarantee that the new release is indeed working as expected before it becomes available to our users.

Before we see how we can accomplish *blue-green deployment* with *Docker Flow*, let's destroy the containers we just run and start over.

```bash
docker-compose down
```

Blue-Green Deployment
---------------------

*Docker Flow* is a single binary sitting on top of Docker Compose and utilizing service discovery to decide which actions should be performed. It allows a combination of three kinds of inputs; command line arguments, environment variables, and *docker-flow.yml* definition. We'll start with command line arguments.

```bash
./docker-flow \
    --consul-address=http://$CONSUL_IP:8500 \
    --target=app \
    --side-target=db \
    --blue-green
```

Let's see which containers are running.

```bash
docker ps \
    --format "table{{.Image}}\t{{.Status}}\t{{.Names}}"
```

The output of the `docker ps` command is as follows.

```
IMAGE                     STATUS              NAMES
vfarcic/books-ms:latest   Up About a minute   dockerflow_app-blue_1
mongo                     Up About a minute   books-ms-db
progrium/consul           Up 32 minutes       consul
```

The major difference when compared with examples without Docker Flow is the name of the deployed target. This time, the container is named *dockerflow_app-blue_1*. Since this was the first deployment, only the *blue* release is running.

Let's see what happens when we deploy the second release. This time we'll use a combination of environment variables and *docker-flow.yml* file. The later is as follows.

```yml
target: app
side_targets:
  - db
blue_green: true
```

As you can see, the arguments in the *docker-flow.yml* file are (almost) the same as those we used through the command line. The only difference is that keys are using dash (*-*) instead of underscore so that they follow YML conventions. The second difference is in the way lists (in this case *side_targets*) are defined in YML.

Environment variables also follow the same namings but, just like YML, formatted to follow appropriate naming conventions. The are all in capital letters. Another diffference is the *FLOW* prefix. It is added so that environment variables used with *Docker Flow* do no override other variables that might be in use by your system.

It is up to you to choose whether you prefer command line arguments, YML specification, environment variables, or a combination of those three.

Let's repeat the deployment. This time we'll specify the Consul address as an environment variable and use the *docker-flow.yml* for the rest of arguments.

```bash
export FLOW_CONSUL_ADDRESS=http://$CONSUL_IP:8500

./docker-flow --flow=deploy

docker ps \
    --format "table{{.Image}}\t{{.Status}}\t{{.Names}}"
```

The output of the `docker ps` command is as follows.

```
IMAGE                     STATUS              NAMES
vfarcic/books-ms:latest   Up 24 seconds       dockerflow_app-green_1
vfarcic/books-ms:latest   Up 4 minutes        dockerflow_app-blue_1
mongo                     Up About an hour    books-ms-db
progrium/consul           Up 2 hours          consul
```

As you can see, this time, both releases are running in parallel. The *green* release has joined the *blue* release we run earlier. At this moment, you should run your *integration tests* (I tend to call them *post-deployment tests*) and, if everything seems to be working correctly, change your proxy to point to the new release (*green*). The choice how to do that is yours (I tend to us *Consul Template* for to reconfigure my *nginx* or *HAProxy*). The plan is to incorporate proxy reconfiguration and reloading into *Docker Flow*. Until then, you should do that yourself.

Once the deployment is tested and the proxy is reconfigured, your users will be redirected to the new release, you can stop the old one. *Docker Flow* can help you with that as well.

```bash
./docker-flow --flow=stop-old

docker ps -a \
    --format "table{{.Image}}\t{{.Status}}\t{{.Names}}"
```

The output of the `docker ps` command is as follows.

```
IMAGE                     STATUS                        NAMES
vfarcic/books-ms:latest   Up 5 minutes                  dockerflow_app-green_1
vfarcic/books-ms:latest   Exited (137) 38 seconds ago   dockerflow_app-blue_1
mongo                     Up About an hour              books-ms-db
progrium/consul           Up 2 hours                    consul
```

As you can see, it stopped the old release (blue). Containers were intentionally only stopped and not removed so that you can easily rollback in case you discover that something went wrong after the proxy has been reconfigured.

Let's take a look at the second *Docker Flow* feature.

Relative Scaling
----------------

Just like with Docker Compose, *Docker Flow* allows you to scale the a service to a fixed number of instances (only this time using *blue-green* process).

Let's scale our service to two instances. Since, this time, we are not deploying a new release, we can skip pulling it from *Docker Hub*.

```bash
./docker-flow \
    --flow=deploy \
    --scale=2

docker ps \
    --format "table{{.Image}}\t{{.Status}}\t{{.Names}}"
```

The output of the `docker ps` command is as follows.

```
IMAGE                     STATUS              NAMES
vfarcic/books-ms:latest   Up 4 seconds        dockerflow_app-blue_2
vfarcic/books-ms:latest   Up 4 seconds        dockerflow_app-blue_1
vfarcic/books-ms:latest   Up About a minute   dockerflow_app-green_1
mongo                     Up 4 hours          books-ms-db
progrium/consul           Up 5 hours          consul
```


As expected, two instances of the new release (*blue*) were deployed. This behaviour is the same as what Docker Compose offers (except the addition of blue-green deployment). What *Docker Flow* allows us is to prepend the *scale* value with plus (*+*) or minus (*-*) signs. Let's see it in action before discussing the benefits.

```bash
./docker-flow \
    --flow=deploy \
    --scale=+2

docker ps \
    --format "table{{.Image}}\t{{.Status}}\t{{.Names}}"
```

The output of the `docker ps` command is as follows.

```
IMAGE                     STATUS              NAMES
vfarcic/books-ms:latest   Up 6 seconds        dockerflow_app-green_4
vfarcic/books-ms:latest   Up 7 seconds        dockerflow_app-green_3
vfarcic/books-ms:latest   Up 7 seconds        dockerflow_app-green_2
vfarcic/books-ms:latest   Up 22 minutes       dockerflow_app-blue_2
vfarcic/books-ms:latest   Up 22 minutes       dockerflow_app-blue_1
vfarcic/books-ms:latest   Up 24 minutes       dockerflow_app-green_1
mongo                     Up 5 hours          books-ms-db
progrium/consul           Up 5 hours          consul
```

Since there were two instances of the previous release, the new release was deployed and the number of instances was increase by two (four in total). While this is useful when we want to deploy a new release and know in advance that the number of instances should be scaled, a much more commonly used option will be to run *Docker Flow* with the *--flow=scale* argument. It will follow the same rules of scaling (and de-scaling) but without deploying a new release. Before we try it out, let's stop the old release.

```
./docker-flow --flow=stop-old
```

Let's try `--flow scale` by descaling the number of instances by one.

```bash
./docker-flow --scale=-1 --flow=scale

docker ps \
    --format "table{{.Image}}\t{{.Status}}\t{{.Names}}"
```

The output of the `docker ps` command is as follows.

```
IMAGE                     STATUS              NAMES
vfarcic/books-ms:latest   Up 19 minutes       dockerflow_app-green_3
vfarcic/books-ms:latest   Up 19 minutes       dockerflow_app-green_2
vfarcic/books-ms:latest   Up 43 minutes       dockerflow_app-green_1
mongo                     Up 5 hours          books-ms-db
progrium/consul           Up 5 hours          consul
```

The number of running instances reduced by one (from four to three). The scale and descale by using relative values has many usages. You might, for example, schedule some of your services to scale every Monday morning because you know, from experience, that's the time when it receives increased traffic. Following the same scenario, you might want to descale every Monday afternoon because at that time traffic tends to go back to normal and you'd like to free resources for some other services, or, even, stop machines on AWS and save one usage. When scaling/descaling is automated, using absolute number is very risky. You might have a script that scales from four to six instances during peak hours. After some time, normal hours might require eight instances and scaling on peak hours to six would have an opposite effect by actually descaling. The need for relative scaling and descaling is even more obvious in case of [Self-Healing Systems](http://technologyconversations.com/2016/01/26/self-healing-systems/).

The Roadmap
===========

Even though the examples from this article used a single server, the main use case for those features is inside a Docker Swarm cluster. For information regarding all arguments that can be used, please refer to the [Docker Flow README](https://github.com/vfarcic/docker-flow).

This is the very beginning of the *Docker Flow* project. The main idea behind it is to provide an easy way to execute processes (like blue-green deployment) and provide easy integration with other tools (like Consul). I have a huge list of features I'm planning to add, however, before I announce them, I would like to get some feedback. Do you think this project would be useful? Which features would you like to see next? With which tools would you like it to be integrated with?