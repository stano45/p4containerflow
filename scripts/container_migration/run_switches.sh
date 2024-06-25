sudo podman network create s1-net --subnet 10.0.0.0/24

sudo podman run -d \
    --name s1 \
    --privileged \
    --replace \
    --network s1-net \
    --ip 10.0.0.0 \
    -v /home/p4/p4containerflow/scripts/container_migration/s1.sh:/s1.sh \
    -v /home/p4/p4containerflow/load_balancer/build/load_balance.json:/load_balance.json \
    --entrypoint /s1.sh \
    p4c