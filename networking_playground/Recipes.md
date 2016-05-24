## Isolating containers for clean sharing between networks for flexible CI/CD.

Setup:

```bash
$ git clone https://github.com/adnaan/docker-flow-proxy.git

$ cd docker-flow

$ vagrant plugin install vagrant-cachier

# Takes a while to finish.
$ vagrant up swarm-master swarm-node-1 swarm-node-2 proxy

$ vagrant ssh proxy

$ cd /vagrant/networking_playground

# This will create master, integration and custom overlays with
# respective containers. It will also create orphan overlay Networks
# with "nonet" containers which are initially connected to the
# bridge network. This takes a while to finish.
$ ./multienv.py

$ docker ps -qa --filter "name=service*" --format={{.Names}}

$ docker network ls

# To destroy everything created above
$ ./destroy_multienv.sh
```


### Using a "frontend" network and links with overlays:

The git versions master, integration and custom are used as references to create
images, containers and networks. A small Go program "reply.go" is selectively compiled
to emulate an array of different services.

This is done in the create_container.sh script:

```bash
$ env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \
"-X main.port=$1 -X main.name=$2 -X main.branch=$3" -o reply .

$ docker build --rm=true --build-arg PORT=$1 -t $2:$3 .
```

We are redirecting traffic from port 80 to the service's set PORT so that we can do this:

```bash
vagrant@proxy:/vagrant/networking_playground$ docker exec  -it service2-custom wget -qO- service3.myntra.com
I am service3:custom listening on port 3333
```

We want to share service1-master and service5-integration to the "custom" network.
 On the other hand, service1-master and service5-integration containers should not
callback the corresponding containers in the custom network but rather talk to their
original network. To enforce that we create a "frontend" overlay network which acts a mediator
between the various revisions and yet provides the isolation.

```bash
$ docker network create -d overlay --subnet=16.0.0.0/24 frontend
```

The "custom" network has service2-custom, service3-custom and service4-custom containers
with service2.myntra.com, service3.myntra.com and service4.myntra.com DNS's respectively.
Same naming convention goes for "master" and "integration" networks. Both "master" and
"integration" networks have service<revision>(1-5) containers.


Now let's connect service1-master and service5-integration to frontend.

```bash
vagrant@proxy:/vagrant/networking_playground$ docker network connect frontend service1-master
vagrant@proxy:/vagrant/networking_playground$ docker network connect frontend service5-integration
# link the custom network container to each master and integration.

vagrant@proxy:/vagrant/networking_playground$ docker network connect --link=service1-master:service1.myntra.com \
--link service5-integration:service5.myntra.com frontend service2-custom
vagrant@proxy:/vagrant/networking_playground$ docker exec -it service1-master wget -qO- service5.myntra.com
I am service5:master listening on port 5555

vagrant@proxy:/vagrant/networking_playground$ docker exec -it service1-master wget -qO- service2.myntra.com
I am service2:master listening on port 2222

vagrant@proxy:/vagrant/networking_playground$ docker exec -it service2-custom wget -qO- service1.myntra.com
I am service1:master listening on port 1111
vagrant@proxy:/vagrant/networking_playground$ docker exec -it service2-custom wget -qO- service3.myntra.com
I am service3:custom listening on port 3333
vagrant@proxy:/vagrant/networking_playground$ docker exec -it service2-custom wget -qO- service4.myntra.com
I am service4:custom listening on port 4444
vagrant@proxy:/vagrant/networking_playground$ docker exec -it service2-custom wget -qO- service5.myntra.com
I am service5:integration listening on port 5555

```
Since the developer explicitly requires master and integration containers in the cluster,
 we can safely assume a one-many mapping L.H.S is the custom network container
 and R.H.S is the master/integration containers.

In this approach the overhead is creating an extra overlay
for each cluster. The drawbacks are:

1. Creating an extra overlay for each cluster.

### Using only links without overlays

To build a cluster we need to build a dependency map. Let's assume

```bash
service1-master-nonet-> service2-master-nonet,service3-master-nonet
service2-master-nonet-> service2-master-nonet,service4-master-nonet
service3-master-nonet-> service1-master-nonet,service3-master-nonet,service5-master-nonet
service4-master-nonet ...
service5-master-nonet ...

```

First we need to create the master-nonet cluster

```bash
$ docker network connect --link=service2-master-nonet:service2.myntra.com \
  --link=service3-master-nonet:service3.myntra.com \
  master-nonet service1-master-nonet
```

We do this for all services individually. Since containers are only connected through
linked aliases there is no danger of dirty shared calls between networks.

```bash
$ docker network connect --link=service2-integration-nonet:service2.myntra.com \
  --link=service3-integration-nonet:service3.myntra.com \
  integration-nonet service1-integration-nonet
```

Now to use service1-master-nonet and service5-master-nonet in the custom overlay,
all we need to do is connect them to it.

```bash
$ docker network connect --alias=service1.myntra.com custom service1-master-nonet
$ docker network connect --alias=service5.myntra.com custom service5-master-nonet
```

In this approach the overhead is local to the container. The drawbacks are:

1. Explicit dependency map for each service.
2. If a new service is added to the dependency graph. All dependent containers need to disconnect and connect with the new links map.
