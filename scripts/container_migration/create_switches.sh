sudo podman run -d \
    --name p4c \
    --privileged \
    --replace \
    --pod testpod \
    -v $(pwd)/s1.sh:/s1.sh \
    -v /home/p4/p4containerflow/load_balancer/build/load_balance.json:/load_balance.json \
    --entrypoint /s1.sh \
    p4c