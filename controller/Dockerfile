FROM python:3.11-slim

ENV PYTHONUNBUFFERED=1

ENV DEPS libreadline-dev \
    valgrind \
    libtool-bin \
    libprotobuf-dev \
    protobuf-compiler \
    protobuf-compiler-grpc \
    libgrpc-dev \
    libgrpc++-dev \
    git \
    make \
    automake \
    build-essential \
    g++ \
    libtool \ 
    pkg-config \
    ca-certificates \
    libboost-dev \
    libboost-system-dev \
    libboost-thread-dev

WORKDIR /app

COPY controller.py /app/
COPY node_manager.py /app/
COPY switch_controller.py /app/
COPY p4runtime_lib /app/p4runtime_lib
COPY requirements.txt /app/

RUN pip install --no-cache-dir -r requirements.txt && \
    apt-get update && \
    apt-get install -y --no-install-recommends $DEPS && \
    git clone https://github.com/p4lang/PI.git && \
    cd PI && \
    git checkout 05cb92564af77ae4826565cbde84e3fd4960c6bd && \
    git submodule update --init --recursive && \
    ./autogen.sh && \
    ./configure --with-proto --without-internal-rpc --without-cli --without-bmv2 && \
    make -j2 && \
    make install-strip && \
    ldconfig && \
    apt-get purge -y $DEPS && \
    apt-get autoremove --purge -y && \
    apt-get clean && \
    rm -rf /app/PI /var/lib/apt/lists/* /tmp/* /var/tmp/*


ENTRYPOINT ["python3", "controller.py"]
