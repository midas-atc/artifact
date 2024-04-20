#!/bin/bash
echo 'update'
echo -e '\n'

rsync -aP -e  'ssh -p 22' --exclude '.git' --exclude '.gitignore' --exclude '.DS_Store' /Users/***/OneDrive\ -\ HKUST\ Connect/rsync/docker10-13/ gpu06:/home/***/glusterfs/manage/docker10-13/

rsync -aP -e  'ssh -p 30041' --exclude '.git' --exclude '.gitignore' --exclude '.DS_Store' /Users/***/OneDrive\ -\ HKUST\ Connect/rsync/docker10-13/all-exec.sh ubuntu@cpu03:~/***/script/