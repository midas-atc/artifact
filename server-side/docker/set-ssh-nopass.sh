#!/bin/bash

declare -a serverIdx=("3" "4" "5" "6")
declare -a ip_suffix=(0 1 2 3)
declare -a nic=(0)
BASE_IP_SUFFIX=10

for i1 in "${serverIdx[@]}"; do
    for j1 in ${ip_suffix[@]}; do
        for k1 in ${nic[@]}; do
            sshpass -p *** ssh-copy-id root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1}))
        done
    done
done