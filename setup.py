# -*- encoding: utf-8 -*-

import os
import sys
from setuptools import setup, find_packages

assert sys.version_info >= (2, 7), "Python 2.7+ required."

current_dir = os.path.abspath(os.path.dirname(__file__))
with open(os.path.join(current_dir, 'README.markdown')) as readme_file:
    with open(os.path.join(current_dir, 'CHANGES.markdown')) as changes_file:
        long_description = readme_file.read() + '\n' + changes_file.read()

sys.path.insert(0, current_dir + os.sep + 'src')
from beast import VERSION
release = ".".join(str(num) for num in VERSION)

setup (
    name = 'beast',
    version = release,
    author = 'Grupa Allegro Sp. z o.o. and Contributors',
    author_email = 'it-beast-dev@allegro.pl',
    description = "Beast, ralph api client API",
    long_description = long_description,
    url = 'http://beast.allegrogroup.com/',
    keywords = '',
    platforms = ['any'],
    license = 'Apache Software License v2.0',
    packages = find_packages('src'),
    include_package_data = True,
    package_dir = {'':'src'},
    zip_safe = False, # because templates are loaded from file path
    install_requires = [
	'requests==0.14.2',
        'slumber',
	'pygments','docopt','pyyaml'
    ],
    entry_points={
        'console_scripts': [
		'beast = beast.beast:main',
        ],
    },
    classifiers = [
        'Development Status :: 4 - Beta',
        'Framework :: Django',
        'Intended Audience :: System Administrators',
        'License :: OSI Approved :: Apache Software License',
        'Natural Language :: English',
        'Operating System :: POSIX',
        'Operating System :: MacOS :: MacOS X',
        'Operating System :: Microsoft :: Windows :: Windows NT/2000',
        'Programming Language :: Python',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 2 :: Only',
        'Topic :: Internet :: WWW/HTTP',
        ]
    )
