#!/bin/bash

set -x 
set -e

mkdir ~/.beast 2>/dev/null

rm -rf ~/.beast/virtual 2>/dev/null
rm -rf ~/.beast/beast 2>/dev/null
rm -f ~/beast 2>/dev/null


curl https://raw.githubusercontent.com/pypa/virtualenv/master/virtualenv.py >/tmp/virtualenv.py
python2.7 /tmp/virtualenv.py --no-site-packages ~/.beast/virtual/ 2>/dev/null
source ~/.beast/virtual/bin/activate

cd ~/.beast

git clone https://github.com/allegro/ralph_beast.git
cd ralph_beast
pip install -e .
ln -s ~/.beast/virtual/bin/beast ~/beast
chmod a+x ~/beast
