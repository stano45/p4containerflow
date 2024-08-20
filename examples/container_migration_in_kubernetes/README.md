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

### 1. Install CNI Plugins on each node

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

### 2. Initialize the Kubernetes cluster using kubeadm (optional):
```
sudo kubeadm init --pod-network-cidr=10.85.0.0/16 --cri-socket=unix:///var/run/crio/crio.sock
```

### 3. Untaint the master node to allow pods to be scheduled (optional, assuming a single node cluster):
```
kubectl taint nodes --all node-role.kubernetes.io/master-
kubectl taint nodes --all node-role.kubernetes.io/control-plane-
```


### 4. Deploy daemonset
```
kubectl apply -f manifests/kube-router-daemonset.yaml
```

### 5. Deploy an HTTP server

```
kubectl apply -f manifests/http-server-deployment.yaml
kubectl apply -f manifests/http-server-service.yaml

# Check the status of the deployment
kubectl get deployments

# Check the assigned IP address
kubectl get service http-server
```

### 6. Apply the RBAC configuration to allow the checkpoint plugin to create a checkpoint (optional if your config already allows this):
First, replace `<your_machine_name>` with the name of your machine. Then, run:
```
kubectl apply -f manifests/checkpoint-rbac.yaml
```

### 7. Setup a local container registry (optional, you can use any other registry)

```
cd local-registry/
./generate-password.sh <user>
./generate-certificates.sh <hostname>
./trust-certificates.sh
./run.sh

buildah login <hostname>:5000
```

### 8. Install the kubectl checkpoint plugin

```
sudo cp kubectl-plugin/kubectl-checkpoint /usr/local/bin/
```

### 9. Enable checkpoint/restore with established TCP connections
```
sudo mkdir -p /etc/criu/
echo "tcp-established" | sudo tee -a /etc/criu/runc.conf
```

### 10. Create container checkpoint

```
kubectl checkpoint <pod> <container>
```

### 11. Build a checkpoint OCI image and push to registry

```
build-image/build-image.sh -a <annotations-file> -c <checkpoint-path> -i <hostname>:5000/<image>:<tag>

buildah push <hostname>:5000/<image>:<tag>
```

### 12. Restore container from checkpoint image

Replace the container `image` filed in `http-server-deployment.yaml` with the
checkpoint OCI image `<hostname>:5000/<image>:<tag>` and apply the new deployment.

```
kubectl apply -f manifests/http-server-deployment.yaml
```

