sudo apt install -y build-essential

sudo apt install -y libprotobuf-dev libprotobuf-c-dev protobuf-c-compiler protobuf-compiler python3-protobuf

sudo apt install -y  libbsd-dev pkg-config libbsd-dev iproute2 libnftables-dev libcap-dev libnl-3-dev libnl-3-200 libnet1-dev libnet1 libnl-3-dev libnet-dev libaio-dev libgnutls28-dev python3-future libdrm-dev

sudo apt install -y asciidoc xmlto

# CRIU
printf "\nInstalling CRIU...\n"
git config --global advice.detachedHead false
git clone --depth 1 --branch master https://github.com/checkpoint-restore/criu.git
cd criu
make criu -j$(nproc)
cp ./criu/criu /sbin
criu --version || exit 1
cd ..
# rm -rf criu
