#!/bin/bash

declare -a server=("03" "04" "05" "06")
declare -a serverIdx=("3" "4" "5" "6")
#declare -a server=("01")
#declare -a serverIdx=("1")
declare -a ip_suffix=(0 1 2 3)
declare -a nic=(0)

BASE_IP_SUFFIX=10


echo "Step1: Create Docker Containers"

for i in "${server[@]}"; do
    echo "create 4 docker containers in GPU$i"
    for j in ${ip_suffix[@]}; do
        for k in ${nic[@]}; do
            docker_ip_suffix=$((${BASE_IP_SUFFIX}+$j))
            core_begin=$(($j*20))
            core_end=$((${core_begin}+19))
            ssh -o StrictHostKeyChecking=no -t gpu$i "sudo docker stop midas-2node-$j"
            ssh -o StrictHostKeyChecking=no -t gpu$i "sudo docker rm midas-2node-$j"
            ssh -o StrictHostKeyChecking=no -t gpu$i "sudo /data/glusterfs/***/create-docker.sh rdma$k ns$k midas-2node-$j \"$(($j*2)),$(($j*2+1))\" ${docker_ip_suffix} ${core_begin}-${core_end}"
        done
    done
done



<<COMMENT

ssh -o StrictHostKeyChecking=no -t gpu0$i "sudo /data/glusterfs/home/***/create-docker/create-docker.sh rdma2 ns2 4 105 40-49"

echo "Step2: Setup SSH No Password Access"

for i1 in "${serverIdx[@]}"; do
    for j1 in ${ip_suffix[@]}; do
        for k1 in ${nic[@]}; do
            sshpass -p *** ssh-copy-id root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1}))
            for i2 in "${serverIdx[@]}"; do
                for j2 in ${ip_suffix[@]}; do
                    for k2 in ${nic[@]}; do
                        echo "set ssh from 10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) to 10.${k1}.${i2}.$((${BASE_IP_SUFFIX}+${j2}))"
                        sshpass -p *** ssh -o StrictHostKeyChecking=no root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) -t "sshpass -p *** ssh-copy-id root@10.${k1}.${i2}.$((${BASE_IP_SUFFIX}+${j2}))";
                    done
                done
            done
        done
    done
done


COMMENT