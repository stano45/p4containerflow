
sudo sysctl -w net.ipv4.ip_forward=1
sudo kubeadm reset -f --cri-socket=unix:///var/run/crio/crio.sock
sudo rm -f /var/log/bmv2-cni.log

sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --cri-socket=unix:///var/run/crio/crio.sock

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

kubectl taint nodes --all node-role.kubernetes.io/master-
kubectl taint nodes --all node-role.kubernetes.io/control-plane-

kubectl apply -f manifests/kube-router-daemonset.yaml

sleep 5
kubectl apply -f manifests/http-server-deployment.yaml
kubectl apply -f manifests/http-client-deployment.yaml

kubectl get pods --all-namespaces