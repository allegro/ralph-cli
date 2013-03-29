#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""Beast is convenient Ralph API commandline client.

Documentation on https://github.com/allegro/ralph_beast/

Usage:
 beast show
 beast show <resource> [--schema] [--filter=field_filter] [--pyhton-filter=python_filter] [--fields=fields] [--csv] [--trim] [--limit=limit]
 beast update <resource> <id> <fields> <fields_values>
 beast -h | --help
 beast --version

Options:
  -h --help     Show this screen.
  --version     Show version.

"""

import csv
import os
import pprint
import pygments
import pygments.formatters
import pygments.lexers
import requests
import slumber
import sys
import urlparse

from collections import defaultdict
from docopt import docopt

class Api(object):
    def get_session(self, settings):
        url = settings.get('url')
        username = settings.get('username')
        api_key = settings.get('api_key')
        if not username or not api_key or not url:
            print("username, api_key and url in ~/.beast/config are required.")
            sys.exit(2)
        session = requests.session(verify=False)
        session = slumber.API(
            '%(url)s/api/v0.9/' % dict(url=url),
            session=session,
        )
        return session

    def put_resource(self, settings, resource, id, data):
        session = self.get_session(settings)
        username = settings.get('username')
        api_key = settings.get('api_key')
        try:
            data = getattr(session, resource)(id).put(
                data=data,
                username=username,
                api_key=api_key
            )
        except slumber.exceptions.HttpClientError as error:
            print 'Error: ', error.content
        return data

    def get_resource(self, settings, resource, limit=None, filters=None,
        python_filter=None):
        session = self.get_session(settings)
        resource = '' if not resource else resource
        limit = 0 if not limit else limit
        attrs = [
            ('username', settings.get('username')),
            ('api_key', settings.get('api_key')),
            ('limit', limit),
        ]
        if filters:
            url_dict = urlparse.parse_qsl(filters)
            attrs.extend(url_dict)
        attrs_dict = dict(attrs)
        return getattr(session, resource).get(**dict(attrs_dict))


class Content(object):
    def inspect(self, arguments, settings, resource=None):
        # TODO rzniecie scheme
        if not resource:
            print "-" * 50
            data = Api().get_resource(settings, '')
            list_of_resources = [x for x in data]
            print '\n'.join(sorted(list_of_resources))
        else:
            print "Available fields for resource: %s" % resource
            print "-" * 50
            data = Api().get_resource(settings, resource)
            if data.get('objects'):
                first_item = data['objects'][0]
                list_of_keys = [key for key in first_item.keys()]
                print '\n'.join(sorted(list_of_keys))

    def get_api_objects(self, data, python_filter, output_fields, csv_export):
        result_data = []

        if not 'objects' in data:
            return data, []
        for row in data['objects']:
            try:
                code = eval("%s" % (python_filter or True))
            except Exception as error:
                print("Filter epression invalid: %s" % error)
                sys.exit(5)
            of = output_fields.split(',') if output_fields else row.keys()
            if code:
                if csv_export:
                    result_data.append(
                        ','.join([unicode(self.multiget(row, key)) for key in of])
                    )
                else:
                    self.remove_links(row)
                    result_data.append(row)
        return result_data, of

    def csv_format(self, header, content):
        for row in header:
            sys.stdout.write(row.encode('utf-8'))
            sys.stdout.write(",")
        print
        for item in content:
            print item.encode('utf-8')

    def multiget(self, row, key):
        if not key:
            print("Empty field error. Please check if all fields are given")
            sys.exit(1)
        actual = row
        for nested in key.split('.'):
            try:
                actual = actual[nested]
            except KeyError:
                print "Unknown field: %s" % key
                sys.exit(3)
        return actual

    def highlight(self, string):
        if sys.stdout.isatty() or sys.stdin.isatty():
            return string
        else:
            return pygments.highlight(
                string,
                pygments.lexers.YamlLexer(),
                pygments.formatters.Terminal256Formatter()
            )

    def console_repr(self, field):
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
            field = ','.join([self.console_repr(subfield) for subfield in field])
        else:
            field = unicode(field)
        return field

    def remove_links(self, row):
        for field in row.keys():
            row[field] = self.console_repr(row[field])

    def smallest_list_of(self, widths, output_fields, max_width):
        of_truncated = []
        current_width = 0
        for key in output_fields:
            width = widths.get(key)
            if current_width + width > max_width:
                return of_truncated
            else:
                of_truncated.append(key)
            current_width += width
        return of_truncated

    def show_header(self, of, trim_columns, result_data, output_fields,
        max_width, widths):
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
        if not output_fields:
            of = self.smallest_list_of(widths, of, max_width)
        print "-" * max_width
        sys.stdout.write("|")
        for key in of:
            fill = widths.get(key)
            align = widths.get(key)
            sys.stdout.write(
                ' {: <{fill}.{align}}|'.format(
                    key,
                    fill=fill-4,
                    align=align-4
                )
            )
        sys.stdout.write('\n')

    def show_content(self, of, result_data, trim_columns, output_fields,
                                                            max_width, widths):
        self.show_header(
            of,
            trim_columns,
            result_data,
            output_fields,
            max_width,
            widths,
        )
        for row in result_data:
            print "-" * max_width
            sys.stdout.write("|")
            for key in of:
                fill = widths.get(key)
                align = widths.get(key)
                sys.stdout.write(
                    ' {: <{fill}.{align}}|'.format(
                        row[key].encode('utf-8'),
                        fill=fill-4,
                        align=align-4
                    )
                )
            sys.stdout.write('\n')


def show(arguments, settings):
    pp = pprint.PrettyPrinter(indent=4)
    max_width = 120
    widths = defaultdict(int)
    resource = arguments.get('<resource>', '')
    limit = arguments.get('--limit')
    schema = arguments.get('--schema')
    filters = arguments.get('--filter')
    python_filter = arguments.get('--python-filter')
    output_fields = arguments.get('--fields')
    csv_export = arguments.get('--csv')
    trim_columns = arguments.get('--trim')
    data = Api().get_resource(
        settings,
        resource,
        limit,
        filters,
        python_filter,
    )
    result_data, of = Content().get_api_objects(
        data,
        python_filter,
        output_fields,
        csv_export,
    )
    if not resource:
        print "Ralph API schema"
        return Content().inspect(arguments, settings)
    if schema:
        # Rżnąć schema
        print "Ralph API schema for %s" % resource
        return Content().inspect(arguments, settings, resource=resource)
    if limit:
        print "Limited rows requested: %s" % limit
    # Return format (default: table)
    if csv_export:
        return Content().csv_format(header=of, content=result_data)
    else:
        return Content().show_content(
            of,
            result_data,
            trim_columns,
            output_fields,
            max_width,
            widths,
        )


def update(arguments, settings):
    resource = arguments.get('<resource>')
    id = arguments.get('<id>')
    fields = arguments.get('<fields>').split(',')
    fields_values = arguments.get('<fields_values>')
    for row in csv.reader([fields_values],
        quotechar='"', quoting=csv.QUOTE_MINIMAL):
        pass
    data = dict(zip(fields, row))
    Api.put_resource(settings, resource, id, data)


def do_main(arguments):
    settings = dict()
    config_file = os.path.expanduser("~/.beast/config")
    try:
        execfile(config_file, settings)
    except IOError:
        print("Config file ~/.beast/config doesn't exist.")
        sys.exit(4)
    if arguments.get('show'):
        show(arguments, settings)
    elif arguments.get('update'):
        update(arguments, settings)


def main():
    arguments = docopt(__doc__, version='0.1')
    do_main(arguments)


if __name__ == '__main__':
    main()
