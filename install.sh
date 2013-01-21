#!/bin/sh

mkdir ~/.beast 2>/dev/null

rm -rf ~/.beast/virtual 2>/dev/null
rm -rf ~/.beast/beast 2>/dev/null

curl https://raw.github.com/pypa/virtualenv/master/virtualenv.py >/tmp/virtualenv.py
python /tmp/virtualenv.py --no-site-packages ~/.beast/virtual/ 2>/dev/null
source ~/.beast/virtual/bin/activate

cd ~/.beast

git clone git@github.com:vi4m/beast.git
cd beast 
pip install -e . 
ln -s ~/.beast/beast/src/beast/beast ~/beast
chmod a+x ~/.beast/beast/src/beast/beast 
