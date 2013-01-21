#!/usr/bin/env python

"""Beast - ralph REST client.

Usage:
 beast export <resource> [--filter=filter_expression] [--fields=fields] [--csv]
 beast update <resource> <id> <fields> <fields_values>
 beast inspect [--resource=resource]
 beast -h | --help
 beast --version

Options:
  -h --help     Show this screen.
  --version     Show version.

"""

import os
import pprint
import sys

from docopt import docopt
import pygments
import pygments.formatters
import pygments.lexers
import requests
import slumber
import yaml
import csv


def put_resource(settings, resource, id, data):
    username = settings.get('username')
    api_key = settings.get('api_key')
    url = settings.get('url')
    if not username or not api_key or not url:
        print("username, api_key and url in ~/.beast/config are required.")
        sys.exit(2)
    s = requests.session(verify=False)
    s = slumber.API(
        '%(url)s/api/v0.9/' % dict(url=url),
        session=s,
    )
    try:
        data = getattr(s, resource)(id).put(
        data=data,username=username, api_key=api_key)
    except slumber.exceptions.HttpClientError as e:
        print 'Error: ', e.content
    return data


def get_resource(settings, resource):
    username = settings.get('username')
    api_key = settings.get('api_key')
    url = settings.get('url')
    if not username or not api_key or not url:
        print("username, api_key and url in ~/.beast/config are required.")
        sys.exit(0)
    s = requests.session(verify=False)
    s = slumber.API(
        '%(url)s/api/v0.9/' % dict(url=url),
        session=s,
    )
    data = getattr(s, resource).get(username=username, api_key=api_key)
    return data


def do_main(arguments):
    settings = dict()
    f = os.path.expanduser("~/.beast/config")
    try:
        execfile(f, settings)
    except IOError:
        print("Config file ~/.beast/config doesn't exist.")
        sys.exit(4)

    if arguments.get('inspect'):
        inspect(arguments, settings)
    elif arguments.get('export'):
        export(arguments, settings)
    elif arguments.get('update'):
        update(arguments, settings)


def update(arguments, settings):
    resource = arguments.get('<resource>')
    id = arguments.get('<id>')
    fields = arguments.get('<fields>').split(',')
    fields_values = arguments.get('<fields_values>')

    for row in csv.reader([fields_values],
        quotechar='"', quoting=csv.QUOTE_MINIMAL):
        pass

    data = dict(zip(fields, row))
    s = put_resource(settings, resource, id, data)


def inspect(arguments, settings):
    resource = arguments.get('--resource')
    if not resource:
        print "Available resources:"
        print "-" * 50
        data = get_resource(settings, '')
        list_of_resources = [x for x in data]
        print '\n'.join(sorted(list_of_resources))
    else:
        print "Available fields for resource: %s" % resource
        print "-" * 50
        data = get_resource(settings, resource)
        if data.get('objects'):
            first_item = data['objects'][0]
            list_of_keys = [x for x in first_item.keys()]
            print '\n'.join(sorted(list_of_keys))


def export(arguments, settings):
    pp = pprint.PrettyPrinter(indent=4)
    resource = arguments.get('<resource>')
    filter_expression = arguments.get('--filter')
    output_fields = arguments.get('--fields')
    csv_export = arguments.get('--csv')
    data = get_resource(settings, resource)
    result_data = []
    first = True

    for row in data['objects']:
        if not first:
            first = False

        e = eval("%s" % (filter_expression or True))
        all_fields = row.keys()
        of = output_fields.split(',') if output_fields else all_fields
        if e:
            if csv_export:
                print ','.join([unicode(multiget(row, key)) for key in of])
            else:
                result_data.append(row)

    if result_data:
        print high(yaml.safe_dump(result_data))


def multiget(row, key):
    if not key:
        print("Empty field error. Please check if all fields are given")
        sys.exit(1)
    actual = row
    nested = key.split('.')
    for n in nested:
        try:
            actual = actual[n]
        except KeyError:
            print "Unknown field: %s" % key
            sys.exit(3)
    return actual


def high(s):
    if sys.stdout.isatty() or sys.stdin.isatty():
        return s
    else:
        return pygments.highlight(
            s,
            pygments.lexers.YamlLexer(),
            pygments.formatters.Terminal256Formatter()
        )


def main():
    arguments = docopt(__doc__, version='0.1')
    do_main(arguments)

if __name__ == '__main__':
    main()

