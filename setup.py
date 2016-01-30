# -*- encoding: utf-8 -*-

import os
import sys
from setuptools import setup, find_packages
from distutils.core import setup
import platform
if platform.system() == 'Windows':
    import py2exe

assert sys.version_info >= (2, 7), "Python 2.7+ required."

current_dir = os.path.abspath(os.path.dirname(__file__))
with open(os.path.join(current_dir, 'README.rst')) as readme_file:
    with open(os.path.join(current_dir, 'CHANGES.rst')) as changes_file:
        long_description = readme_file.read() + '\n' + changes_file.read()

sys.path.insert(0, current_dir + os.sep + 'src')
from ralph_cli import VERSION
release = ".".join(str(num) for num in VERSION)

setup(
    name='ralph_cli',
    version=release,
    author='Grupa Allegro Sp. z o.o. and Contributors',
    author_email='pylabs@allegro.pl',
    description="Official Ralph API client",
    long_description=long_description,
    url='http://github.com/allegro/ralph_cli',
    keywords='',
    platforms=['any'],
    license='Apache Software License v2.0',
    packages=find_packages('src'),
    include_package_data=True,
    package_dir={'': 'src'},
    zip_safe=False,  # because templates are loaded from file path
    console=['src/ralph_cli/ralph_cli.py'],
    install_requires=[
        'requests==0.14.2',
        'slumber',
        'docopt',
        'colorconsole',
    ],
    entry_points={
        'console_scripts': [
        'ralph-cli = ralph_cli.ralph_cli:main',
        ],
    },
    options={
        'py2exe': {
            'bundle_files': 1,
            'optimize': 2,
            'packages': 'beast',
            'compressed': True,
            "dll_excludes": ["mfc90.dll"],
            'excludes': [
                'doctest',
                'pdb',
                'unittest',
                'difflib',
                'inspect',
                'pyreadline',
                'optparse',
                'pickle',
                'email',
            ]
        },
    },
    zipfile=None,
    classifiers=[
        'Development Status :: 4 - Beta',
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
