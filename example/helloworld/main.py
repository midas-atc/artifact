import os
import shutil

WORKDIR = os.environ.get('MIDAS_WORKDIR')
USERDIR = os.environ.get('MIDAS_USERDIR')

os.system('tree -L 2 {}'.format(USERDIR))

shutil.copytree(WORKDIR, "{}/helloworld".format(USERDIR))
