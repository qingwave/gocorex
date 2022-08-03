docker volume create --name etcd-data
export DATA_DIR="etcd-data"

REGISTRY=quay.io/coreos/etcd
VERSION=v3.5.4

# NODE1 is your local ip
# ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}' | head -n 1
export NODE1={}

docker rm -f etcd
docker run \
  -d \
  -p 2379:2379 \
  -p 2380:2380 \
  --volume=$DATA_DIR:/etcd-data \
  --name etcd $REGISTRY:$VERSION \
  /usr/local/bin/etcd \
  --data-dir=/etcd-data --name node1 \
  --initial-advertise-peer-urls http://$NODE1:2380 --listen-peer-urls http://0.0.0.0:2380 \
  --advertise-client-urls http://$NODE1:2379 --listen-client-urls http://0.0.0.0:2379 \
  --initial-cluster node1=http://$NODE1:2380
