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
    import base64
    import hashlib
    import hmac
    import itertools
    import json
    import requests
    import ssl
    import sys
    import time
    import urllib
    import urllib2

    from datetime import datetime, timedelta
    from requests_toolbelt import SSLAdapter
    from urllib2 import HTTPError, URLError
except ImportError, e:
    print "Import error in %s : %s" % (__name__, e)
    import sys
    sys.exit()


# Disable HTTPS verification warnings.
from requests.packages import urllib3
urllib3.disable_warnings()


def logger_debug(logger, message):
    if logger is not None:
        logger.debug(message)


def writeError(msg):
    sys.stderr.write(msg)
    sys.stderr.write("\n")
    sys.stderr.flush()


def login(url, username, password, domain="/", verifysslcert=False):
    """
    Login and obtain a session to be used for subsequent API calls
    Wrong username/password leads to HTTP error code 531
    """
    args = {}

    args["command"] = 'login'
    args["username"] = username
    args["password"] = password
    args["domain"] = domain
    args["response"] = "json"

    sessionkey = ''
    session = requests.Session()
    session.mount('https://', SSLAdapter(ssl.PROTOCOL_TLSv1))

    try:
        resp = session.post(url, params=args, verify=verifysslcert)
    except requests.exceptions.ConnectionError, e:
        writeError("Connection refused by server: %s" % e)
        return None, None

    if resp.status_code == 200:
        sessionkey = resp.json()['loginresponse']['sessionkey']
    elif resp.status_code == 405:
        writeError("Method not allowed, unauthorized access on URL: %s" % url)
        session = None
        sessionkey = None
    elif resp.status_code == 531:
        writeError("Error authenticating at %s using username: %s"
                   ", password: %s, domain: %s" % (url, username, password,
                                                   domain))
        session = None
        sessionkey = None
    else:
        resp.raise_for_status()

    return session, sessionkey


def logout(url, session):
    if not session:
        return
    session.get(url, params={'command': 'logout'})


def make_request_with_password(command, args, logger, url, credentials,
                               verifysslcert=False):
    error = None
    args = args.copy()
    username = credentials['username']
    password = credentials['password']
    domain = credentials['domain']

    if not (username and password):
        error = "Username and password cannot be empty"
        result = None
        return result, error

    tries = 0
    retry = True

    while tries < 2 and retry:
        sessionkey = credentials.get('sessionkey')
        session = credentials.get('session')
        tries += 1

        # obtain a valid session if not supplied
        if not (session and sessionkey):
            session, sessionkey = login(url, username, password, domain,
                                        verifysslcert)
            if not (session and sessionkey):
                return None, 'Authentication failed'
            credentials['session'] = session
            credentials['sessionkey'] = sessionkey

        args['sessionkey'] = sessionkey

        # make the api call
        resp = session.get(url, params=args, verify=verifysslcert)
        result = resp.text
        logger_debug(logger, "Response received: %s" % resp.text)

        if resp.status_code == 200:  # success
            retry = False
            break
        if resp.status_code == 401:  # sessionkey is wrong
            credentials['session'] = None
            credentials['sessionkey'] = None
            continue

        if resp.status_code != 200 and resp.status_code != 401:
            error_message = resp.headers.get('X-Description')
            if not error_message:
                error_message = resp.text
            error = "{0}: {1}".format(resp.status_code, error_message)
            result = None
            retry = False

    return result, error


def make_request(command, args, logger, url, credentials, expires,
                 verifysslcert=False, signatureversion=3):
    result = None
    error = None

    if not url.startswith('http'):
        error = "Server URL should start with 'http' or 'https', " + \
                "please check and fix the url"
        return None, error

    if not args:
        args = {}

    args = args.copy()
    args["command"] = command
    args["response"] = "json"
    signatureversion = int(signatureversion)
    if signatureversion >= 3:
        args["signatureversion"] = signatureversion
        if not expires:
            expires = 600
        expirationtime = datetime.utcnow() + timedelta(seconds=int(expires))
        args["expires"] = expirationtime.strftime('%Y-%m-%dT%H:%M:%S+0000')

    for key in args.keys():
        value = args[key]
        if isinstance(value, unicode):
            value = value.encode("utf-8")
        args[key] = value
        if not key:
            args.pop(key)
        else:
            if key in ['publickey', 'privatekey', 'certificate']:
                args[key] = urllib.quote_plus(str(value))

    # try to use the apikey/secretkey method by default
    # followed by trying to check if we're using integration port
    # finally use the username/password method
    if not credentials['apikey'] and not ("8096" in url):
        try:
            return make_request_with_password(command, args, logger, url,
                                              credentials, verifysslcert)
        except (requests.exceptions.ConnectionError, Exception), e:
            return None, e

    def sign_request(params, secret_key):
        request = zip(params.keys(), params.values())
        request.sort(key=lambda x: x[0].lower())
        hash_str = "&".join(
            ["=".join(
                [r[0].lower(),
                 urllib.quote_plus(str(r[1])).lower()
                 .replace("+", "%20").replace("%3A", ":")]
            ) for r in request]
        )
        return base64.encodestring(hmac.new(secret_key, hash_str,
                                   hashlib.sha1).digest()).strip()

    args['apiKey'] = credentials['apikey']
    args["signature"] = sign_request(args, credentials['secretkey'])

    session = requests.Session()
    session.mount('https://', SSLAdapter(ssl.PROTOCOL_TLSv1))

    try:
        response = session.get(url, params=args, verify=verifysslcert)
        logger_debug(logger, "Request sent: %s" % response.url)
        result = response.text

        if response.status_code == 200:  # success
            error = None
        elif response.status_code == 401:      # auth issue
            error = "401 Authentication error"
        elif response.status_code != 200 and response.status_code != 401:
            error = "{0}: {1}".format(response.status_code,
                                      response.headers.get('X-Description'))
    except requests.exceptions.ConnectionError, e:
        return None, "Connection refused by server: %s" % e
    except Exception, pokemon:
        error = pokemon.message

    logger_debug(logger, "Response received: %s" % result)
    if error is not None:
        logger_debug(logger, "Error: %s" % (error))
        return result, error

    return result, error


def monkeyrequest(command, args, isasync, asyncblock, logger, url,
                  credentials, timeout, expires, verifysslcert=False,
                  signatureversion=3):
    response = None
    error = None
    logger_debug(logger, "======== START Request ========")
    logger_debug(logger, "Requesting command=%s, args=%s" % (command, args))
    response, error = make_request(command, args, logger, url,
                                   credentials, expires, verifysslcert,
                                   signatureversion)

    logger_debug(logger, "======== END Request ========\n")

    if error is not None and not response:
        return response, error

    def process_json(response):
        try:
            response = json.loads(response, "utf-8")
        except ValueError, e:
            logger_debug(logger, "Error processing json: %s" % e)
            writeError("Error processing json: %s" % e)
            response = None
            error = e
        return response

    response = process_json(response)
    if not response or not isinstance(response, dict):
        return response, error

    isasync = isasync and (asyncblock == "true" or asyncblock == "True")
    responsekey = filter(lambda x: 'response' in x, response.keys())[0]

    if isasync and 'jobid' in response[responsekey]:
        jobid = response[responsekey]['jobid']
        command = "queryAsyncJobResult"
        request = {'jobid': jobid}
        if not timeout:
            timeout = 3600
        timeout = int(timeout)
        cursor = itertools.cycle([u'|', u'/', u'-', u'\\'])
        while timeout > 0:
            interval = 2
            while interval > 0:
                sys.stdout.write(u"%s\r" % cursor.next())
                sys.stdout.flush()
                time.sleep(0.1)
                interval -= 0.1
            timeout = timeout - 2
            logger_debug(logger, "Job %s to timeout in %ds" % (jobid, timeout))

            response, error = make_request(command, request, logger, url,
                                           credentials, expires, verifysslcert)
            if error and not response:
                return response, error

            response = process_json(response)
            responsekeys = filter(lambda x: 'response' in x, response.keys())

            if len(responsekeys) < 1:
                continue

            result = response[responsekeys[0]]
            if "errorcode" in result or "errortext" in result:
                return response, error

            jobstatus = result['jobstatus']
            if jobstatus == 2:
                jobresult = result["jobresult"]
                error = "\rAsync job %s failed\nError %s, %s" % (
                        jobid, jobresult["errorcode"], jobresult["errortext"])
                return response, error
            elif jobstatus == 1:
                sys.stdout.write(u"\r \n")
                sys.stdout.flush()
                return response, error
            elif jobstatus == 0:
                pass  # Job in progress
            else:
                logger_debug(logger, "We should not arrive here!")
                sys.stdout.flush()

        error = "Error: Async query timeout occurred for jobid %s" % jobid

    return response, error
