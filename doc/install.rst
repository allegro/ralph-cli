==============
How to install
==============

Linux
---------------

The simplest way to install ralph ralph-cli is(you need curl):

  $ curl https://raw.githubusercontent.com/allegro/ralph-cli/beast/install.sh | bash -


MacOS
---------------

Put on your console below command to install.::

  $ curl https://raw.githubusercontent.com/allegro/ralph-cli/beast/install.sh | bash -


Configuration
=============

Now you need configuration file. Create file 

~/.ralph_cli/config

Windows binary already contains the file inside directory.

Config file should contain such data  :ref:`config_file`


.. _config_file:

Config file - example
---------------------
::

  username="jan.kowalski"
  api_key="478457f9f32323201ebde8ef79cd9d3a028ced56747"
  url="https://ralph-url.com/"
  version="1"
  
  
Obtaining api_key
---------------------

You can find you api_key by clicking on your username on the bottom of the ralph page and selecting API Key link from the menu on the left.
