#!/bin/bash

#set -eu -o pipefail

if [ $UID -ne 0 ]; then
	echo "please execute this script as root"
	exit 1
fi

function create_docker_network() {
	DEV=$1
	NAME=$2

	docker network ls | grep $NAME 2>&1 >/dev/null
	if [ $? -eq 0 ]; then
		echo $NAME network has been created
		return 0
	fi

	echo "This may take several seconds, please wait..."

	SUBNET=`ip a show $DEV | grep 'inet ' | awk '{print $2}'`
	GW=`ip r | grep $DEV | grep via | awk '{print $3}'`

	echo $DEV, $NAME, $SUBNET, $GW

	echo "docker network create -d sriov --subnet=$SUBNET --gateway=$GW -o netdevice=$DEV $NAME"
    docker network create -d sriov --subnet=$SUBNET --gateway=$GW -o netdevice=$DEV $NAME

	num_vfs=`cat /sys/class/net/$DEV/device/sriov_numvfs`
	for ((i=0;i<$num_vfs;i++)); do
		# set speed to 40Gbps
		sudo ip link set $DEV vf $i trust on
		sudo ip link set $DEV vf $i max_tx_rate 100000 min_tx_rate 100000
	done
}

IMAGE=lt-3090:latest
#DIR_MAPPING=" --volume /home/***/glusterfs/testbed/share_dir:/share_dir --volume /home/***/glusterfs/mlt/testbed/code:/code --volume /data/glusterfs/public:/data "
DIR_MAPPING=" --volume /mnt/home:/mnt/home --volume /mnt/data:/mnt/data"
SSH_PORT=22

function create_container() {
	DEV=$1
	NET_NAME=$2
	CONTAINER_NAME=$3
	GPU_ID=$4
    DOCKER_IP_ID=$5
    CPU_SET=$6
	GW=`ip r | grep $DEV | grep via | awk '{print $3}'`
	IP=`echo $GW | awk 'BEGIN{FS="."}{print $1 "." $2 "." $3 "."}'`$DOCKER_IP_ID

    HN=`echo $IP | awk 'BEGIN{FS="."}{print $1 "-" $2 "-" $3 "-" $4}'`

	echo "conatiner: $CONTAINER_NAME, IP: $IP"

	echo "docker_rdma_sriov run -it -d --cpus=10 --cpuset-cpus=$CPU_SET --name=$CONTAINER_NAME -h --hostname=$HN $DIR_MAPPING --net=$NET_NAME --ip=$IP --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=$GPU_ID --ulimit memlock=-1 --shm-size=65536m --cap-add=IPC_LOCK --cap-add SYS_NICE --device=/dev/infiniband $IMAGE /usr/sbin/sshd -D -p $SSH_PORT"
    docker_rdma_sriov run -it -d --cpus=40 --cpuset-cpus=$CPU_SET --name=$CONTAINER_NAME --hostname=$HN $DIR_MAPPING --net=$NET_NAME --ip=$IP --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=$GPU_ID --ulimit memlock=-1 --shm-size=65536m --cap-add=IPC_LOCK --cap-add SYS_NICE --device=/dev/infiniband $IMAGE /usr/sbin/sshd -D -p $SSH_PORT

	pid=$(sudo docker inspect -f '{{.State.Pid}}' $CONTAINER_NAME)
	nsenter -t $pid -n ip route add 10.0.0.0/8 dev 0 via $GW
	nsenter -t $pid -n ip link set 0 mtu 1500

	#mkdir -p /var/run/netns
	#ln -s /proc/$pid/ns/net /var/run/netns/$pid
	#ip netns exec $pid ip route add 10.0.0.0/8 dev eth0 via $GW

	#docker exec -it $CONTAINER_NAME ip route add 10.0.0.0/8 dev eth0 via $GW

	docker network connect bridge $CONTAINER_NAME

	# inspect
	docker exec -it $CONTAINER_NAME ip addr
	docker exec -it $CONTAINER_NAME ip route
}


create_docker_network rdma0 ns0

# create_container rdma0 ns0 lt-3090-test-0-1 0 52 0-9
# create_container rdma2 ns2 lt-2-1 1 105 10-19

echo "create_container $1 $2 $3 $4 $5 $6"
create_container $1 $2 $3 $4 $5 $6
# docker network create -d sriov --subnet=10.0.1.0/24 -o netdevice=rdma0 rdma0