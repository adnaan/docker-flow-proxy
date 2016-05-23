## Isolating containers for clean sharing between networks for flexible CI/CD.

```bash
$ git clone https://github.com/adnaan/docker-flow-proxy.git

$ cd docker-flow

$ vagrant plugin install vagrant-cachier

# Takes a while to finish.
$ vagrant up swarm-master swarm-node-1 swarm-node-2 proxy

$ vagrant ssh proxy

$ cd /vagrant/replier

# This will create master, integration and custom overlays with
# respective containers. It will also create orphan overlay Networks
# with "nonet" containers which are initially connected to the
# bridge network. This takes a while to finish.
$ ./multienv.py

# To destroy everything created above
$ ./destroy_multienv.sh
```


### Using a "frontend" network and links with overlays:

We want to share service1-master and service5-integration to the "custom" network.
What we don't want is that service1-master and service5-integration containers to
callback the corresponding containers from the custom network. To do that we create a
"frontend" overlay which acts a mediator between the various revisions and still
provides the required

```bash
$ docker network create -d overlay --subnet=16.0.0.0/24 frontend
```

The "custom" network has service2-custom, service3-custom and service4-custom containers
with service2.myntra.com, service3.myntra.com and service4.myntra.com DNS's respectively.
Same naming convention goes for "master" and "integration" networks. Both "master" and
"integration" networks have service<revision>(1-5) containers.


Now let's connect service1-master and service5-integration to frontend.

```bash
vagrant@proxy:/vagrant/replier$ docker network connect frontend service1-master
vagrant@proxy:/vagrant/replier$ docker network connect frontend service5-integration
# link the custom network container to each master and integration.

vagrant@proxy:/vagrant/replier$ docker network connect --link=service1-master:service1.myntra.com \
--link service5-integration:service5.myntra.com frontend service2-custom
vagrant@proxy:/vagrant/replier$ docker exec -it service1-master sh
/ # wget -qO- service5.myntra.com:5555
I am service5:master listening on port 5555
/ # wget -qO- service2.myntra.com:2222
I am service2:master listening on port 2222

vagrant@proxy:/vagrant/replier$ docker exec -it service2-custom sh
/ # wget -qO- service1.myntra.com:1111
I am service1:master listening on port 1111
/ # wget -qO- service3.myntra.com:3333
I am service3:custom listening on port 3333
/ # wget -qO- service4.myntra.com:4444
I am service4:custom listening on port 4444
```
Since the developer explicity requires master and integration containers, we can safely assume a one-many mapping L.H.S is the custom network container and R.H.S is the master/integration containers.

In this approach the overhead is creating an extra overlay
for each cluster. The drawbacks are:
1. Creating an extra overlay

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
