#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""Beast - ralph REST client.

Usage:
 beast export <resource> [--filter=filter_expression] [--fields=fields] [--csv] [--yaml] [--trim] [--limit=limit]
 beast update <resource> <id> <fields> <fields_values>
 beast export
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
import csv

from collections import defaultdict


def get_session(settings):
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
    return s


def put_resource(settings, resource, id, data):
    username = settings.get('username')
    api_key = settings.get('api_key')
    s = get_session(settings)
    try:
        data = getattr(s, resource)(id).put(
            data=data,
            username=username,
            api_key=api_key
        )
    except slumber.exceptions.HttpClientError as e:
        print 'Error: ', e.content
    return data


def get_resource(settings, resource, limit=None):
    username = settings.get('username')
    api_key = settings.get('api_key')
    s = get_session(settings)
    if limit:
        data = getattr(s, resource).get(limit=limit, username=username, api_key=api_key)
    else:
        data = getattr(s, resource).get(limit=0, username=username, api_key=api_key)
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
    put_resource(settings, resource, id, data)


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
        data = get_resource(settings, resource, limit=1)
        if data.get('objects'):
            first_item = data['objects'][0]
            list_of_keys = [x for x in first_item.keys()]
            print '\n'.join(sorted(list_of_keys))


def console_repr(field):
    if type(field) == type(u''):
        if '/api/v0.9/' in field:
            field = unicode(field.replace('/api/v0.9/',''))
        else:
            field = unicode(field[:10])
    elif type(field) == type(False):
        field = 'x' if field else 'o'
    elif type(field) == type({}):
        field = field['id']
    elif type(field) == type(None):
        field = ''
    elif type(field) == type([]):
        field = ','.join([console_repr(subfield) for subfield in field])
    else:
        field = unicode(field)
    return field



def remove_links(row):
    for field in row.keys():
        row[field] = console_repr(row[field])


def smallest_list_of(widths, of, max_width):
    of_truncated = []
    current_width = 0
    for key in of:
        width = widths.get(key)
        if current_width + width > max_width:
            return of_truncated
        else:
            of_truncated.append(key)
        current_width += width
    return of_truncated


def export(arguments, settings):
    trim_columns = arguments.get('--trim')
    pp = pprint.PrettyPrinter(indent=4)
    resource = arguments.get('<resource>')
    limit = arguments.get('--limit')
    filter_expression = arguments.get('--filter')
    output_fields = arguments.get('--fields')
    csv_export = arguments.get('--csv')
    if not resource:
        print "Resource not specified. Type beast inspect [resource] to inspect available fields."
        return inspect(arguments, settings)
    data = get_resource(settings, resource, limit)
    result_data = []
    first = True
    if limit:
        print "Limited rows requested: %s" % limit
    for row in data['objects']:
        if not first:
            first = False
        try:
            e = eval("%s" % (filter_expression or True))
        except Exception as e:
            print("Filter epression invalid: %s" % e)
            sys.exit(5)

        all_fields = row.keys()
        of = output_fields.split(',') if output_fields else all_fields
        if e:
            if csv_export:
                result_data.append(','.join([unicode(multiget(row, key)) for key in of]))
            else:
                remove_links(row)
                result_data.append(row)
    if csv_export:
        for row in of:
            sys.stdout.write(row)
            sys.stdout.write(",")
        print
        for i in result_data:
            print(i)
    else:
        widths = defaultdict(int)
        for row in of:
            if not trim_columns:
                widths[row] = len(row) + 4
            else:
                widths[row] = 5

        for row in result_data:
            for key in row.keys():
                widths[key] = max(widths[key], len(row[key])+4)
                all_fields = row.keys()
                of = output_fields.split(',') if output_fields else all_fields

        max_width = 120

        if not output_fields:
            of = smallest_list_of(widths, of, max_width)

        print "-" * max_width
        sys.stdout.write("|")
        for key in of:
            fill = widths.get(key)
            align = widths.get(key)
            sys.stdout.write(
                ' {: <{fill}.{align}}|'.format(
                    key, fill=fill-4, align=align-4))
        sys.stdout.write('\n')

        for row in result_data:
            print "-" * max_width
            sys.stdout.write("|")
            for key in of:
                #key = row.keys()[key_row]
                fill = widths.get(key)
                align = widths.get(key)
                sys.stdout.write(
                    ' {: <{fill}.{align}}|'.format(
                        row[key].encode('utf-8','ignore'), fill=fill-4, align=align-4))
            sys.stdout.write('\n')


    #if result_data:
    #    print high(yaml.safe_dump(result_data))


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

