# -*- coding: utf-8 -*-
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

__version__ = "5.2.0"
__description__ = "Command Line Interface for Apache CloudStack"
__maintainer__ = "The Apache CloudStack Team"
__maintaineremail__ = "dev@cloudstack.apache.org"
__project__ = "The Apache CloudStack Team"
__projectemail__ = "dev@cloudstack.apache.org"
__projecturl__ = "http://cloudstack.apache.org"

try:
    import os
    import sys

    from ConfigParser import ConfigParser
    from os.path import expanduser
except ImportError, e:
    print "ImportError", e

param_type = ['boolean', 'date', 'float', 'integer', 'short', 'list',
              'long', 'object', 'map', 'string', 'tzdate', 'uuid']

iterable_type = ['set', 'list', 'object']

config_dir = expanduser('~/.cloudmonkey')
config_file = expanduser(config_dir + '/config')

# cloudmonkey config fields
mandatory_sections = ['core', 'ui']
default_profile_name = 'local'
config_fields = {'core': {}, 'ui': {}}

# core
config_fields['core']['asyncblock'] = 'true'
config_fields['core']['paramcompletion'] = 'false'
config_fields['core']['cache_file'] = expanduser(config_dir + '/cache')
config_fields['core']['history_file'] = expanduser(config_dir + '/history')
config_fields['core']['log_file'] = expanduser(config_dir + '/log')
config_fields['core']['profile'] = default_profile_name

# ui
config_fields['ui']['color'] = 'true'
config_fields['ui']['prompt'] = '> '
config_fields['ui']['display'] = 'default'

# default profile
default_profile = {}
default_profile['url'] = 'http://localhost:8080/client/api'
default_profile['timeout'] = '3600'
default_profile['expires'] = '600'
default_profile['username'] = 'admin'
default_profile['password'] = 'password'
default_profile['apikey'] = ''
default_profile['secretkey'] = ''

def write_config(get_attr, config_file, first_time=False):
    global config_fields
    config = ConfigParser()
    if os.path.exists(config_file) and not first_time:
        config = ConfigParser()
        try:
            with open(config_file, 'r') as cfg:
                config.readfp(cfg)
        except IOError, e:
            print "Error: config_file not found", e
    for section in config_fields.keys():
        config.add_section(section)
        for key in config_fields[section].keys():
            if first_time:
                config.set(section, key, config_fields[section][key])
            else:
                config.set(section, key, get_attr(key))
    with open(config_file, 'w') as cfg:
        config.write(cfg)
    return config


def read_config(get_attr, set_attr, config_file):
    global config_fields, config_dir
    if not os.path.exists(config_dir):
        os.makedirs(config_dir)

    config_options = reduce(lambda x, y: x + y, map(lambda x:
                            config_fields[x].keys(), config_fields.keys()))

    if os.path.exists(config_file):
        config = ConfigParser()
        try:
            with open(config_file, 'r') as cfg:
                config.readfp(cfg)
        except IOError, e:
            print "Error: config_file not found", e
    else:
        config = write_config(get_attr, config_file, True)
        print "Welcome! Using `set` configure the necessary settings:"
        print " ".join(sorted(config_options))
        print "Config file:", config_file
        print "After setting up, run the `sync` command to sync apis\n"

    missing_keys = []
    for section in config_fields.keys():
        for key in config_fields[section].keys():
            try:
                set_attr(key, config.get(section, key))
            except Exception:
                set_attr(key, config_fields[section][key])
                missing_keys.append(key)

    if len(missing_keys) > 0:
        print "Found missing config keys and replace with default value:"
        print "`%s` in %s" % (', '.join(missing_keys), config_file)
        write_config(get_attr, config_file, False)

    return config_options
