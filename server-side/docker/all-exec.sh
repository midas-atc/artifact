#!/bin/bash

declare -a serverIdx=("3" "4" "5" "6")
declare -a ip_suffix=(0 1 2 3)
declare -a nic=(0)
BASE_IP_SUFFIX=10

for i1 in "${serverIdx[@]}"; do
    for j1 in ${ip_suffix[@]}; do
        for k1 in ${nic[@]}; do
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "apt update"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "apt install -y clustershell munge slurm-wlm mysql-client"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "sudo bash -c 'echo \"Name=gpu File=/dev/nvidia\"$(($j1*2)) > /etc/slurm-llnl/gres.conf'"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "sudo bash -c 'echo \"Name=gpu File=/dev/nvidia\"$(($j1*2+1)) > /etc/slurm-llnl/gres.conf'"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "mkdir -p /var/run/munge"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "chown -R munge:munge /var/run/munge"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "cp /mnt/data/munge.key /etc/munge/munge.key"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "cat /mnt/data/midas-etc-users.txt >> /etc/passwd"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "cat /mnt/data/midas-hosts.txt >> /etc/hosts"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "cp /mnt/data/slurm.conf /etc/slurm-llnl/"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "sudo /usr/sbin/munged --force"
            sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "sudo /usr/sbin/slurmd"
        done
    done
done

<<COMMENT

sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "useradd -m ubuntu -s /bin/bash -u 1000 -G sudo"
sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "echo -e \"***\n***\" | sudo passwd ubuntu"
sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "echo \"ubuntu ALL=(ALL) NOPASSWD:ALL\" > /etc/sudoers.d/ubuntu-user"

sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "useradd -m ubuntu -s /bin/bash -u 1000 -G sudo"
sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "echo -e \"***\n***\" | sudo passwd ubuntu"
sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "echo \"ubuntu ALL=(ALL) NOPASSWD:ALL\" > /etc/sudoers.d/ubuntu-user"
sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "sudo bash -c 'echo \"Name=gpu File=/dev/nvidia\"$(($j1*2)) > /etc/slurm-llnl/gres.conf'"
sshpass -p *** ssh root@10.${k1}.${i1}.$((${BASE_IP_SUFFIX}+${j1})) "sudo bash -c 'echo \"Name=gpu File=/dev/nvidia\"$(($j1*2+1)) >> /etc/slurm-llnl/gres.conf'"

useradd -m ubuntu -s /bin/bash -u 1000 -G sudo
echo -e \"***\n***\" | sudo passwd ubuntu
echo \"ubuntu ALL=(ALL) NOPASSWD:ALL\" > /etc/sudoers.d/ubuntu-user
"sudo bash -c 'echo \"Name=gpu File=/dev/nvidia\"$(($j1-1)) > /etc/slurm-llnl/gres.conf'"
COMMENT