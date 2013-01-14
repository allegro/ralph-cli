#!/usr/bin/env python
"""Beast - ralph REST client.

Usage:
 beast export <resource> [--filter=filter_expression] [--output_fields=output_fields] [--csv]
 beast update <resource> <filter_expression> <output_fields>
 beast -h | --help
 beast --version

Options:
  -h --help     Show this screen.
  --version     Show version.

"""

import pygments
import pygments.lexers
import pygments.formatters;

import slumber
import sys
import requests
import pprint
from docopt import docopt
import yaml


if __name__ == '__main__':
    arguments = docopt(__doc__, version='Naval Fate 2.0')


pp = pprint.PrettyPrinter(indent=4)

resource = arguments.get('<resource>')
filter_expression = arguments.get('--filter')
output_fields = arguments.get('--output_fields')
csv_export = arguments.get('--csv')

s = slumber.API(
    #'https://ralph.dc2/api/v0.9/',
    'http://localhost:8000/api/v0.9/',
    session=requests.session(verify=False)
)

username = 'marcin.kliks'
api_key = '478457f935e901ebde8ef79cd9d3a028ced56747'

data = getattr(s, resource).get(username=username, api_key=api_key)
result_data = []
first = True

def multiget(row, key):
    actual = row
    nested = key.split('.')
    for n in nested:
        actual = actual[n]
    return actual


def high(s):
    return pygments.highlight(
        s,
        pygments.lexers.YamlLexer(), 
        pygments.formatters.Terminal256Formatter()
    )


for row in data['objects']:
    if not first:
        first = False

    e = eval("%s" % (filter_expression or True))
    all_fields = row.keys()
    of = output_fields.split(',') if output_fields else all_fields 
    if e:
        if csv_export:
            print ','.join([str(multiget(row, key)) for key in of])
        else:
            result_data.append(row)

if result_data:
    print high(yaml.dump(result_data))