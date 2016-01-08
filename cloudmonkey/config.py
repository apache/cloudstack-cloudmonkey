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

__version__ = "5.3.3"
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

# cloudmonkey display types
display_types = ["json", "xml", "csv", "table", "default"]

config_dir = expanduser('~/.cloudmonkey')
config_file = expanduser(config_dir + '/config')

# cloudmonkey config fields
mandatory_sections = ['core', 'ui']
default_profile_name = 'local'
config_fields = {'core': {}, 'ui': {}}

# core
config_fields['core']['asyncblock'] = 'true'
config_fields['core']['paramcompletion'] = 'true'
config_fields['core']['cache_file'] = expanduser(config_dir + '/cache')
config_fields['core']['history_file'] = expanduser(config_dir + '/history')
config_fields['core']['log_file'] = expanduser(config_dir + '/log')
config_fields['core']['profile'] = default_profile_name

# ui
config_fields['ui']['color'] = 'true'
config_fields['ui']['prompt'] = '🐵 > '
config_fields['ui']['display'] = 'default'

# default profile
default_profile = {}
default_profile['url'] = 'http://localhost:8080/client/api'
default_profile['timeout'] = '3600'
default_profile['expires'] = '600'
default_profile['username'] = 'admin'
default_profile['password'] = 'password'
default_profile['domain'] = '/'
default_profile['apikey'] = ''
default_profile['secretkey'] = ''
default_profile['verifysslcert'] = 'true'
default_profile['signatureversion'] = '3'


def write_config(get_attr, config_file):
    global config_fields, mandatory_sections
    global default_profile, default_profile_name
    config = ConfigParser()
    if os.path.exists(config_file):
        try:
            with open(config_file, 'r') as cfg:
                config.readfp(cfg)
        except IOError, e:
            print "Error: config_file not found", e

    profile = None
    try:
        profile = get_attr('profile')
    except AttributeError, e:
        pass
    if profile is None or profile == '':
        profile = default_profile_name
    if profile in mandatory_sections:
        print "Server profile name cannot be '%s'" % profile
        sys.exit(1)

    has_profile_changed = False
    profile_in_use = default_profile_name
    try:
        profile_in_use = config.get('core', 'profile')
    except Exception:
        pass
    if profile_in_use != profile:
        has_profile_changed = True

    for section in (mandatory_sections + [profile]):
        if not config.has_section(section):
            try:
                config.add_section(section)
                if section not in mandatory_sections:
                    for key in default_profile.keys():
                        config.set(section, key, default_profile[key])
                else:
                    for key in config_fields[section].keys():
                        config.set(section, key, config_fields[section][key])
            except ValueError, e:
                print "Server profile name cannot be", profile
                sys.exit(1)
        if section in mandatory_sections:
            section_keys = config_fields[section].keys()
        else:
            section_keys = default_profile.keys()
        for key in section_keys:
            try:
                if not (has_profile_changed and section == profile):
                    config.set(section, key, get_attr(key))
            except Exception:
                pass
    with open(config_file, 'w') as cfg:
        config.write(cfg)
    return config


def read_config(get_attr, set_attr, config_file):
    global config_fields, config_dir, mandatory_sections
    global default_profile, default_profile_name
    if not os.path.exists(config_dir):
        os.makedirs(config_dir)

    config_options = reduce(lambda x, y: x + y, map(lambda x:
                            config_fields[x].keys(), config_fields.keys()))
    config_options += default_profile.keys()

    config = ConfigParser()
    if os.path.exists(config_file):
        try:
            with open(config_file, 'r') as cfg:
                config.readfp(cfg)
        except IOError, e:
            print "Error: config_file not found", e
    else:
        config = write_config(get_attr, config_file)
        print "Welcome! Use the `set` command to configure options"
        print "Config file:", config_file
        print "After setting up, run the `sync` command to sync apis\n"

    missing_keys = []
    if config.has_option('core', 'profile'):
        profile = config.get('core', 'profile')
    else:
        global default_profile_name
        profile = default_profile_name

    if profile is None or profile == '' or profile in mandatory_sections:
        print "Server profile cannot be", profile
        sys.exit(1)

    set_attr("profile_names", filter(lambda x: x != "core" and x != "ui",
                                     config.sections()))

    if not config.has_section(profile):
        print ("Selected profile (%s) does not exist," +
               " using default values") % profile
        try:
            config.add_section(profile)
        except ValueError, e:
            print "Server profile name cannot be", profile
            sys.exit(1)
        for key in default_profile.keys():
            config.set(profile, key, default_profile[key])

    for section in (mandatory_sections + [profile]):
        if section in mandatory_sections:
            section_keys = config_fields[section].keys()
        else:
            section_keys = default_profile.keys()
        for key in section_keys:
            try:
                set_attr(key, config.get(section, key))
            except Exception, e:
                if section in mandatory_sections:
                    set_attr(key, config_fields[section][key])
                else:
                    set_attr(key, default_profile[key])
                missing_keys.append("%s = %s" % (key, get_attr(key)))
            # Cosmetic fix for prompt
            if key == 'prompt':
                set_attr(key, get_attr('prompt').strip() + " ")

    if len(missing_keys) > 0:
        print "Missing configuration was set using default values for keys:"
        print "`%s` in %s" % (', '.join(missing_keys), config_file)
        write_config(get_attr, config_file)

    return config_options
