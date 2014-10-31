#!/usr/bin/python
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

try:
    import argcomplete
    import argparse
    import atexit
    import cmd
    import copy
    import json
    import logging
    import os
    import shlex
    import sys
    import time
    import types

    from cachemaker import loadcache, savecache, monkeycache, splitverbsubject
    from config import __version__, __description__, __projecturl__
    from config import read_config, write_config, config_file, default_profile
    from optparse import OptionParser
    from prettytable import PrettyTable
    from printer import monkeyprint
    from requester import monkeyrequest
    from requester import login
    from requester import logout
    from urlparse import urlparse
except ImportError, e:
    print("Import error in %s : %s" % (__name__, e))
    import sys
    sys.exit()

try:
    from precache import apicache
except ImportError:
    apicache = {'count': 0, 'verbs': [], 'asyncapis': []}

try:
    import readline
except ImportError, e:
    print("Module readline not found, autocompletions will fail", e)
else:
    import rlcompleter
    if 'libedit' in readline.__doc__:
        readline.parse_and_bind("bind ^I rl_complete")
    else:
        readline.parse_and_bind("tab: complete")

log_fmt = '%(asctime)s - %(filename)s:%(lineno)s - [%(levelname)s] %(message)s'
logger = logging.getLogger(__name__)


class CloudMonkeyShell(cmd.Cmd, object):
    intro = ("☁ Apache CloudStack 🐵 cloudmonkey " + __version__ +
             ". Type help or ? to list commands.\n")
    ruler = "="
    config_options = []
    profile_names = []
    verbs = []
    error_on_last_command = False
    param_cache = {}
    prompt = "🐵 > "
    protocol = "http"
    host = "localhost"
    port = "8080"
    path = "/client/api"

    def __init__(self, pname, cfile):
        self.program_name = pname
        self.config_file = cfile
        self.config_options = read_config(self.get_attr, self.set_attr,
                                          self.config_file)
        self.loadcache()
        self.init_credential_store()
        logging.basicConfig(filename=self.log_file,
                            level=logging.DEBUG, format=log_fmt)
        logger.debug("Loaded config fields:\n%s" % map(lambda x: "%s=%s" %
                                                       (x, getattr(self, x)),
                                                       self.config_options))
        cmd.Cmd.__init__(self)

        try:
            if os.path.exists(self.history_file):
                readline.read_history_file(self.history_file)
        except IOError, e:
            logger.debug(u"Error: Unable to read history. " + unicode(e))
        atexit.register(readline.write_history_file, self.history_file)

    def init_credential_store(self):
        self.credentials = {'apikey': self.apikey, 'secretkey': self.secretkey,
                            'username': self.username,
                            'password': self.password}
        parsed_url = urlparse(self.url)
        self.protocol = "http" if not parsed_url.scheme else parsed_url.scheme
        self.host = parsed_url.netloc
        self.port = "8080" if not parsed_url.port else parsed_url.port
        self.path = parsed_url.path
        self.set_prompt()

    def get_prompt(self):
        return self.prompt.split(") ")[-1]

    def set_prompt(self):
        self.prompt = "(%s) %s" % (self.profile, self.get_prompt())

    def get_attr(self, field):
        return getattr(self, field)

    def set_attr(self, field, value):
        return setattr(self, field, value)

    def emptyline(self):
        pass

    def cmdloop(self, intro=None):
        print(self.intro)
        print "Using management server profile:", self.profile, "\n"
        while True:
            try:
                super(CloudMonkeyShell, self).cmdloop(intro="")
            except KeyboardInterrupt:
                print("^C")

    def loadcache(self):
        if os.path.exists(self.cache_file):
            self.apicache = loadcache(self.cache_file)
        else:
            self.apicache = apicache
        if 'verbs' in self.apicache:
            self.verbs = self.apicache['verbs']

        for verb in self.verbs:
            def add_grammar(verb):
                def grammar_closure(self, args):
                    if not args:
                        return
                    args = args.decode("utf-8")
                    if self.pipe_runner(u"{0} {1}".format(verb, args)):
                        return
                    if ' --help' in args or ' -h' in args:
                        self.do_help("%s %s" % (verb, args))
                        return
                    try:
                        args_partition = args.partition(" ")
                        cmd = self.apicache[verb][args_partition[0]]['name']
                        args = args_partition[2]
                    except KeyError, e:
                        self.monkeyprint("Error: invalid %s api arg " % verb,
                                         str(e))
                        return
                    self.default(u"{0} {1}".format(cmd, args))
                return grammar_closure

            grammar_handler = add_grammar(verb)
            grammar_handler.__doc__ = "%ss resources" % verb.capitalize()
            grammar_handler.__name__ = "do_" + str(verb)
            setattr(self.__class__, grammar_handler.__name__, grammar_handler)

    def monkeyprint(self, *args):
        output = u""
        try:
            for arg in args:
                if isinstance(type(arg), types.NoneType) or not arg:
                    continue
                output += arg
        except Exception, e:
            print(str(e))

        output = output.encode("utf-8")
        if self.color == 'true':
            monkeyprint(output)
        else:
            if output.startswith("Error"):
                sys.stderr.write(output + "\n")
            else:
                print output

    def print_result(self, result, result_filter=[]):
        if not result or len(result) == 0:
            return

        def printer_helper(printer, toprow):
            if printer:
                self.monkeyprint(printer.get_string())
            return PrettyTable(toprow)

        def print_result_json(result, result_filter=None):
            tfilter = {}  # temp var to hold a dict of the filters
            tresult = copy.deepcopy(result)  # dupe the result to filter
            if result_filter:
                for res in result_filter:
                    tfilter[res] = 1
                for okey, oval in result.iteritems():
                    if isinstance(oval, dict):
                        for tkey in oval:
                            if tkey not in tfilter:
                                try:
                                    del(tresult[okey][oval][tkey])
                                except:
                                    pass
                    elif isinstance(oval, list):
                        for x in range(len(oval)):
                            if isinstance(oval[x], dict):
                                for tkey in oval[x]:
                                    if tkey not in tfilter:
                                        try:
                                            del(tresult[okey][x][tkey])
                                        except:
                                            pass
                            else:
                                try:
                                    del(tresult[okey][x])
                                except:
                                    pass
            self.monkeyprint(json.dumps(tresult,
                                        sort_keys=True,
                                        indent=2,
                                        ensure_ascii=False,
                                        separators=(',', ': ')))

        def print_result_tabular(result, result_filter=None):
            toprow = None
            printer = None
            for node in result:
                if toprow != node.keys():
                    if result_filter and len(result_filter) != 0:
                        commonkeys = filter(lambda x: x in node.keys(),
                                            result_filter)
                        if commonkeys != toprow:
                            toprow = commonkeys
                            printer = printer_helper(printer, toprow)
                    else:
                        toprow = node.keys()
                        printer = printer_helper(printer, toprow)
                row = map(lambda x: node[x], toprow)
                if printer and row:
                    printer.add_row(row)
            if printer:
                self.monkeyprint(printer.get_string())

        def print_result_as_dict(result, result_filter=[]):
            if self.display == "json":
                print_result_json(result, result_filter)
                return

            for key in sorted(result.keys(), key=lambda x:
                              x not in ['id', 'count', 'name'] and x):
                if not (isinstance(result[key], list) or
                        isinstance(result[key], dict)):
                    if result_filter and key not in result_filter:
                        continue
                    value = unicode(result[key])
                    if result_filter and len(result_filter) == 1:
                        self.monkeyprint(value)
                    else:
                        self.monkeyprint(key, " = ", value)
                else:
                    if result_filter and key not in result_filter:
                        self.print_result(result[key], result_filter)
                        continue
                    self.monkeyprint(key + ":")
                    self.print_result(result[key], result_filter)

        def print_result_as_list(result, filter=[]):
            for node in result:
                if isinstance(node, dict) and self.display == 'table':
                    print_result_tabular(result, filter)
                    break
                self.print_result(node, filter)
                if result and (filter is None or len(filter) != 1):
                    self.monkeyprint(self.ruler * 80)

        if isinstance(result, dict):
            print_result_as_dict(result, result_filter)
        elif isinstance(result, list):
            print_result_as_list(result, result_filter)
        elif isinstance(result, str):
            print result
        elif result:
            self.monkeyprint(result)

    def make_request(self, command, args={}, isasync=False):
        self.error_on_last_command = False
        response, error = monkeyrequest(command, args, isasync,
                                        self.asyncblock, logger,
                                        self.url, self.credentials,
                                        self.timeout, self.expires)
        if error:
            self.monkeyprint(u"Error {0}".format(error))
            self.error_on_last_command = True
        return response

    def update_param_cache(self, api, result={}):
        if not api:
            return
        logger.debug("Updating param cache for %s API" % api)
        responsekey = filter(lambda x: 'response' in x, result.keys())[0]
        result = result[responsekey]
        options = []
        uuids = []
        for key in result.keys():
            if isinstance(result[key], list):
                for element in result[key]:
                    if 'id' in element.keys():
                        uuid = unicode(element['id'])
                        name = ""
                        keyspace = ["name", "displayname",
                                    "username", "description"]
                        for name_key in keyspace:
                            if name_key in element.keys():
                                name = element[name_key]
                                break
                        options.append((uuid, name,))
                        uuids.append(uuid)
        self.param_cache[api] = {}
        self.param_cache[api]["ts"] = int(time.time())
        self.param_cache[api]["options"] = sorted(options)
        return sorted(uuids)

    def default(self, args):
        if self.pipe_runner(args):
            return

        try:
            args = args.strip()
            args.decode("utf-8")
        except UnicodeError, ignore:
            args = args.encode("utf-8")

        apiname = args.partition(' ')[0]
        verb, subject = splitverbsubject(apiname)

        lexp = shlex.shlex(args)
        lexp.whitespace = " "
        lexp.whitespace_split = True
        lexp.posix = True
        args = []
        while True:
            next_val = lexp.next()
            if not next_val:
                break
            next_val = next_val.decode("utf-8")
            args.append(next_val.replace(u'\x00', u''))

        args_dict = dict(map(lambda x: [x.partition("=")[0],
                                        x.partition("=")[2]],
                             args[1:])[x] for x in range(len(args) - 1))
        field_filter = None
        if 'filter' in args_dict:
            field_filter = filter(lambda x: x is not '',
                                  map(lambda x: x.strip(),
                                      args_dict.pop('filter').split(',')))
            field_filter = list(set(field_filter))

        missing = []
        if verb in self.apicache and subject in self.apicache[verb]:
            missing = filter(lambda x: x not in [key.split('[')[0]
                                                 for key in args_dict],
                             self.apicache[verb][subject]['requiredparams'])

        if len(missing) > 0:
            self.monkeyprint("Missing arguments: ", ' '.join(missing))
            return

        isasync = False
        if 'asyncapis' in self.apicache:
            if apiname.decode("utf-8") in self.apicache["asyncapis"]:
                isasync = True

        result = self.make_request(apiname, args_dict, isasync)

        if not result:
            return

        try:
            responsekeys = filter(lambda x: 'response' in x, result.keys())
            for responsekey in responsekeys:
                self.print_result(result[responsekey], field_filter)
            print
            if apiname.startswith("list") and "id" not in args_dict:
                self.update_param_cache(apiname, result)
        except Exception as e:
            self.monkeyprint("Error on parsing and printing ", e)

    def completedefault(self, text, line, begidx, endidx):
        partitions = line.partition(" ")
        verb = partitions[0].strip()
        rline = partitions[2].lstrip().partition(" ")
        subject = rline[0]
        separator = rline[1]
        params = rline[2].lstrip()

        if verb not in self.verbs:
            return []

        autocompletions = []
        search_string = ""

        if separator != " ":   # Complete verb subjects
            autocompletions = self.apicache[verb].keys()
            search_string = subject
        else:                  # Complete subject params
            autocompletions = map(lambda x: x + "=",
                                  map(lambda x: x['name'],
                                      self.apicache[verb][subject]['params']))
            search_string = text
            if self.paramcompletion == 'true':
                param = line.split(" ")[-1]
                idx = param.find("=")
                value = param[idx + 1:]
                param = param[:idx]
                if param == "filter":
                    response_params = self.apicache[verb][subject]["response"]
                    used = filter(lambda x: x.strip() != "", value.split(",")[:-1])
                    unused = map(lambda x: x['name'] + ",", filter(lambda x:
                                 "name" in x and x["name"] not in used,
                                 response_params))
                    last_value = value.split(",")[-1]
                    if last_value:
                        unused = filter(lambda x: x.startswith(last_value),
                                        unused)
                    suffix = ",".join(used)
                    if suffix:
                        suffix += ","
                    return filter(lambda x: x.startswith(value),
                                  map(lambda x: suffix + x, unused))
                elif len(value) < 36 and idx != -1:
                    api = None
                    logger.debug("[Paramcompl] For %s %s %s=" % (verb, subject,
                                                                 param))
                    if "id" in param:
                        logger.debug("[Paramcompl] Using 'list' heuristics")
                        if param == "id" or param == "ids":
                            entity = subject
                        else:
                            entity = param.replace("id", "")
                        apis = []
                        for resource in self.apicache["list"]:
                            if resource.startswith(entity):
                                api = self.apicache["list"][resource]['name']
                                if (entity + "s") == resource.lower():
                                    break
                                apis.append(api)
                                api = None
                        if len(apis) > 0 and not api:
                            logger.debug("[Paramcompl] APIs: %s" % apis)
                            api = min(apis, key=len)
                        logger.debug("[Paramcompl] Possible API: %s" % api)
                    if not api:
                        logger.debug("[Paramcompl] Using relative approx")
                        params = self.apicache[verb][subject]['params']
                        arg = filter(lambda x: x['name'] == param, params)[0]
                        if "type" in arg and arg["type"] == "boolean":
                            return filter(lambda x: x.startswith(value),
                                          ["true", "false"])
                        related = arg['related']
                        apis = filter(lambda x: 'list' in x, related)
                        logger.debug("[Paramcompl] Related APIs: %s" % apis)
                        if len(apis) > 0:
                            api = apis[0]
                        else:
                            return
                    uuids = []
                    cache_burst_ts = int(time.time()) - 900
                    logger.debug("Trying paramcompletion using API: %s" % api)
                    if api in self.param_cache.keys() and \
                        len(self.param_cache[api]["options"]) > 0 and \
                            self.param_cache[api]["ts"] > cache_burst_ts:
                        for option in self.param_cache[api]["options"]:
                            uuid = option[0]
                            if uuid.startswith(value):
                                uuids.append(uuid)
                    else:
                        api_args = {'listall': 'true', 'templatefilter': 'all'}
                        response = self.make_request(api, args=api_args)
                        if not response:
                            return
                        uuids = self.update_param_cache(api, response)
                    if len(uuids) > 1:
                        print
                        for option in self.param_cache[api]["options"]:
                            uuid = option[0]
                            name = option[1]
                            if uuid.startswith(value):
                                print uuid, name
                    autocompletions = uuids
                    search_string = value

        if subject != "" and line.split(" ")[-1].find('=') == -1:
            autocompletions.append("filter=")
        return [s for s in autocompletions if s.startswith(search_string)]

    def do_sync(self, args):
        """
        Asks cloudmonkey to discovery and sync apis available on user specified
        CloudStack host server which has the API discovery plugin, on failure
        it rollbacks last datastore or api precached datastore.
        """
        response = self.make_request("listApis")
        if not response:
            monkeyprint("Failed to sync apis, please check your config?")
            monkeyprint("Note: `sync` requires api discovery service enabled" +
                        " on the CloudStack management server")
            return
        self.apicache = monkeycache(response)
        savecache(self.apicache, self.cache_file)
        monkeyprint("%s APIs discovered and cached" % self.apicache["count"])
        self.loadcache()

    def do_api(self, args):
        """
        Make raw api calls. Syntax: api <apiName> <args>=<values>.

        Example:
        api listAccount listall=true
        """
        if len(args) > 0:
            return self.default(args)
        else:
            self.monkeyprint("Please use a valid syntax")

    def do_set(self, args):
        """
        Set config for cloudmonkey. For example, options can be:
        url, auth, log_file, history_file
        You may also edit your ~/.cloudmonkey_config instead of using set.

        Example:
        set url http://localhost:8080/client/api
        set prompt 🐵 cloudmonkey>
        set log_file /var/log/cloudmonkey.log
        """
        args = args.strip().partition(" ")
        key, value = (args[0].strip(), args[2].strip())
        if not key:
            return
        if not value:
            print "Blank value of %s is not allowed" % key
            return

        self.prompt = self.get_prompt()
        setattr(self, key, value)
        if key in ['host', 'port', 'path', 'protocol']:
            key = 'url'
            self.url = "%s://%s:%s%s" % (self.protocol, self.host,
                                         self.port, self.path)
            print "This option has been deprecated, please set 'url' instead"
            print "This server url will be used:", self.url
        write_config(self.get_attr, self.config_file)
        read_config(self.get_attr, self.set_attr, self.config_file)
        self.init_credential_store()
        if key.strip() == 'profile':
            print "\nLoaded server profile '%s' with options:" % value
            for option in default_profile.keys():
                value = self.get_attr(option)
                if option in ["password", "apikey", "secretkey"] and value:
                    value = value[:2] + "XXX" + value[4:6] + "YYY...(hidden)"
                print "    %s = %s" % (option, value)
            print

    def complete_set(self, text, line, begidx, endidx):
        mline = line.partition(" ")[2].lstrip().partition(" ")
        option = mline[0].strip()
        separator = mline[1]
        value = mline[2].lstrip()
        if separator == "":
            return [s for s in self.config_options if s.startswith(option)]
        elif option == "profile":
            return [s for s in self.profile_names if s.startswith(value)]
        elif option == "display":
            return [s for s in ["default", "table", "json"]
                    if s.startswith(value)]
        elif option == "asyncblock" or option == "color":
            return [s for s in ["true", "false"] if s.startswith(value)]

        return []

    def do_login(self, args):
        """
        Login using stored credentials. Starts a session to be reused for
        subsequent api calls
        """
        try:
            session, sessionkey = login(self.url, self.username, self.password)
            self.credentials['session'] = session
            self.credentials['sessionkey'] = sessionkey
        except Exception, e:
            self.monkeyprint("Error: Login failed to the server: ", unicode(e))

    def do_logout(self, args):
        """
        Logout of session started with login with username and password
        """
        try:
            logout(self.url, self.credentials.get('session'))
        except Exception, e:
            pass
        self.credentials['session'] = None
        self.credentials['sessionkey'] = None

    def pipe_runner(self, args):
        if args.find(u" |") > -1:
            pname = self.program_name
            if '.py' in pname:
                pname = "python " + pname
            self.do_shell(u"{0} {1}".format(pname, args))
            return True
        return False

    def do_shell(self, args):
        """
        Execute shell commands using shell <command> or !<command>

        Example:
        !ls
        shell ls
        !for((i=0; i<10; i++)); do cloudmonkey create user account=admin \
            email=test@test.tt firstname=user$i lastname=user$i \
            password=password username=user$i; done
        """
        os.system(args.encode("utf-8"))

    def do_help(self, args):
        """
        Show help docs for various topics

        Example:
        help list
        help list users
        ?list
        ?list users
        """
        fields = args.partition(" ")
        if fields[2] == "":
            cmd.Cmd.do_help(self, args)
        else:
            verb = fields[0]
            subject = fields[2].partition(" ")[0]
            if subject in self.apicache[verb]:
                api = self.apicache[verb][subject]
                helpdoc = "(%s) %s" % (api['name'], api['description'])
                if api['isasync']:
                    helpdoc += "\nThis API is asynchronous."
                required = api['requiredparams']
                if len(required) > 0:
                    helpdoc += "\nRequired params are %s" % ' '.join(required)
                helpdoc += "\nParameters\n" + "=" * 10
                for param in api['params']:
                    helpdoc += "\n%s = (%s) %s" % (
                               param['name'], param['type'],
                               param['description'])
                self.monkeyprint(helpdoc)
            else:
                self.monkeyprint("Error: no such api (%s) on %s" %
                                 (subject, verb))

    def complete_help(self, text, line, begidx, endidx):
        fields = line.partition(" ")
        subfields = fields[2].partition(" ")

        if subfields[1] != " ":
            return cmd.Cmd.complete_help(self, text, line, begidx, endidx)
        else:
            line = fields[2]
            text = subfields[2]
            return self.completedefault(text, line, begidx, endidx)

    def do_EOF(self, args):
        """
        Quit on Ctrl+d or EOF
        """
        self.do_logout(None)
        sys.exit()

    def do_exit(self, args):
        """
        Quit CloudMonkey CLI
        """
        return self.do_quit(args)

    def do_quit(self, args):
        """
        Quit CloudMonkey CLI
        """
        self.monkeyprint("Bye!")
        return self.do_EOF(args)


def main():
    displayTypes = ["json", "table", "default"]
    parser = argparse.ArgumentParser(usage="cloudmonkey [options] [commands]",
                                     description=__description__,
                                     epilog="Try cloudmonkey [help|?]")

    parser.add_argument("-v", "--version", action="version",
                        default=argparse.SUPPRESS,
                        version="cloudmonkey %s" % __version__,
                        help="show CloudMonkey's version and exit")

    parser.add_argument("-c", "--config-file",
                        dest="configFile", default=config_file,
                        help="config file for cloudmonkey", metavar="FILE")

    parser.add_argument("-d", "--display-type",
                        dest="displayType", default=None,
                        help="output display type, json, table or default",
                        choices=tuple(displayTypes))

    parser.add_argument("commands", nargs=argparse.REMAINDER,
                        help="API commands")

    argcomplete.autocomplete(parser)
    args = parser.parse_args()

    shell = CloudMonkeyShell(sys.argv[0], args.configFile)

    if args.displayType and args.displayType in displayTypes:
        shell.set_attr("display", args.displayType)

    if len(args.commands) > 0:
        shell.set_attr("color", "false")
        shell.onecmd(" ".join(map(lambda x: x.replace("\\ ", " ")
                                             .replace(" ", "\\ "),
                                  args.commands)))
        if shell.error_on_last_command:
            sys.exit(1)
    else:
        shell.cmdloop()

if __name__ == "__main__":
    main()
