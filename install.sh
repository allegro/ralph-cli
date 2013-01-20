#!/bin/sh

#curl https://raw.github.com/pypa/virtualenv/master/virtualenv.py >/tmp/virtualenv.py
#python /tmp/virtualenv.py --no-site-packages ~/.beast/virtual/
source ~/.beast/virtual/bin/activate

~/.beast/virtual/bin/pip install pygments  slumber docopt pyyaml requests
~/.beast/virtual/bin/python beast.py
