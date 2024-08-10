# Container migration in Kubernetes

In this example, we have HTTP server running in a Kubernetes Pod where a BMv2
switch is used for traffic load balancing to dynamically reroutes packets to
the correct IP address after container migration.

This example assumes that the Kubernenes cluster has been configured with
recent version of CRI-O that supports container checkpointing, and Kubelet
Checkpoint API has been enabled. To learn more about the container
checkpointing feature in Kubernetes, please refer to the following pages:

 - https://kubernetes.io/blog/2022/12/05/forensic-container-checkpointing-alpha/
 - https://kubernetes.io/docs/reference/node/kubelet-checkpoint-api/

## Running the example

1. Install CNI Plugins on each node

The CNI configuration file is expected to be present as `/etc/cni/net.d/10-kuberouter.conf`
```
sudo mkdir -p /etc/cni/net.d/
sudo cp cni/10-kuberouter.conf /etc/cni/net.d/
```

Install `bridge` CNI plugin and `host-local` IP address management plugin:

```
git clone https://github.com/containernetworking/plugins
cd plugins
git checkout v1.1.1

./build_linux.sh

sudo mkdir -p /opt/cni/bin
sudo cp bin/* /opt/cni/bin/
```

2. Deploy daemonset
```
kubectl apply -f manifests/kube-router-daemonset.yaml
```

3. Setup a local container registry

```
cd local-registry/
./generate-password.sh <user>
./generate-certificates.sh <hostname>
./trust-certificates.sh
./run.sh

buildah login <hostname>:5000
```

3. Deploy an HTTP server

```
kubectl apply -f manifests/http-server-deployment.yaml
kubectl apply -f manifests/http-server-service.yaml

# Check the status of the deployment
kubectl get deployments

# Check the assigned IP address
kubectl get service http-server
```

4. Install kubectl checkpoint plugin

```
sudo cp kubectl-plugin/kubectl-checkpoint /usr/local/bin/
```

5. Enable checkpoint/restore with established TCP connections
```
mkdir -p /etc/criu/
echo "tcp-established" >> /etc/criu/runc.conf
```

6. Create container checkpoint

```
kubectl checkpoint <pod> <container>
```

7. Build a checkpoint OCI image and push to registry

```
build-image/build-image.sh -a <annotations-file> -c <checkpoint-path> -i <hostname>:5000/<image>:<tag>

buildah push <hostname>:5000/<image>:<tag>
```

7. Restore container from checkpoint image

Replace the container `image` filed in `http-server-deployment.yaml` with the
checkpoint OCI image `<hostname>:5000/<image>:<tag>` and apply the new deployment.

```
kubectl apply -f manifests/http-server-deployment.yaml
```

