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
    import json
    import requests
    import sys
    import time
    import urllib
    import urllib2
    from datetime import datetime, timedelta
    from urllib2 import HTTPError, URLError

except ImportError, e:
    print "Import error in %s : %s" % (__name__, e)
    import sys
    sys.exit()


def logger_debug(logger, message):
    if logger is not None:
        logger.debug(message)


def login(url, username, password):
    """
    Login and obtain a session to be used for subsequent API calls
    Wrong username/password leads to HTTP error code 531
    """
    args = {}

    args["command"] = 'login'
    args["username"] = username
    args["password"] = password
    args["domain"] = "/"
    args["response"] = "json"

    sessionkey = ''
    session = requests.Session()

    try:
        resp = session.post(url, params=args)
    except requests.exceptions.ConnectionError, e:
        print "Connection refused by server"
        return None, None

    if resp.status_code == 200:
        sessionkey = resp.json()['loginresponse']['sessionkey']
    elif resp.status_code == 405:
        print "Method not allowed, unauthorized access on URL: %s" % url
        session = None
        sessionkey = None
    elif resp.status_code == 531:
        print "Error authenticating at %s, with username: %s" \
              ", and password: %s" % (url, username, password)
        session = None
        sessionkey = None
    else:
        resp.raise_for_status()

    return session, sessionkey


def logout(url, session):
    if session is None:
        return
    session.get(url, params={'command': 'logout'})


def make_request_with_password(command, args, logger, url, credentials):

    error = None
    username = credentials['username']
    password = credentials['password']

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
            session, sessionkey = login(url, username, password)
            if not (session and sessionkey):
                return None, 'Error authenticating'
            credentials['session'] = session
            credentials['sessionkey'] = sessionkey

        args['sessionkey'] = sessionkey

        # make the api call
        resp = session.get(url, params=args)
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
            error = "%s: %s" %\
                    (str(resp.status_code), resp.headers.get('X-Description'))
            result = None
            retry = False

    return result, error

def make_request(command, args, logger, url, 
                 credentials, expires, region):

    response = None
    error = None

    if not url.startswith('http'):
        error = "Server URL should start with 'http' or 'https', " + \
                "please check and fix the url"
        return None, error

    if args is None:
        args = {}

    args["command"] = command
    args["region"] = region
    args["response"] = "json"
    args["signatureversion"] = "3"
    expirationtime = datetime.utcnow() + timedelta(seconds=int(expires))
    args["expires"] = expirationtime.strftime('%Y-%m-%dT%H:%M:%S+0000')

    # try to use the apikey/secretkey method by default
    # followed by trying to check if we're using integration port
    # finally use the username/password method
    if not credentials['apikey'] and not ("8096" in url):
        try:
            return make_request_with_password(command, args,
                                              logger, url, credentials)
        except (requests.exceptions.ConnectionError, Exception), e:
            return None, e

    args['apiKey'] = credentials['apikey']
    secretkey = credentials['secretkey']
    request = zip(args.keys(), args.values())
    request.sort(key=lambda x: x[0].lower())

    request_url = "&".join(["=".join([r[0], urllib.quote_plus(str(r[1]))])
                           for r in request])
    hashStr = "&".join(["=".join([r[0].lower(),
                       str.lower(urllib.quote_plus(str(r[1]))).replace("+",
                       "%20")]) for r in request])

    sig = urllib.quote_plus(base64.encodestring(hmac.new(secretkey, hashStr,
                            hashlib.sha1).digest()).strip())
    request_url += "&signature=%s" % sig

    request_url = "%s?%s" % (url, request_url)
    ##print("Request sent: %s" % request_url) #Debug/test output

    try:
        logger_debug(logger, "Request sent: %s" % request_url)
        connection = urllib2.urlopen(request_url)
        response = connection.read()
    except HTTPError, e:
        error = "%s: %s" % (e.msg, e.info().getheader('X-Description'))
    except URLError, e:
        error = e.reason

    logger_debug(logger, "Response received: %s" % response)
    if error is not None:
        logger_debug(logger, "Error: %s" % (error))
        return response, error

    return response, error

def monkeyrequest(command, args, isasync, asyncblock, logger, url,
                  credentials, timeout, expires, region):
    response = None
    error = None
    logger_debug(logger, "======== START Request ========")
    logger_debug(logger, "Requesting command=%s, args=%s" % (command, args))

    response, error = make_request(command, args, logger, url,
                                   credentials, expires, region)

    logger_debug(logger, "======== END Request ========\n")

    if error is not None:
        return response, error

    def process_json(response):
        try:
            response = json.loads(str(response))
        except ValueError, e:
            logger_debug(logger, "Error processing json: %s" % e)
            print "Error processing json:", str(e)
            response = None
            error = e
        return response

    response = process_json(response)
    if response is None:
        return response, error

    isasync = isasync and (asyncblock == "true")
    responsekey = filter(lambda x: 'response' in x, response.keys())[0]

    if isasync and 'jobid' in response[responsekey]:
        jobid = response[responsekey]['jobid']
        command = "queryAsyncJobResult"
        request = {'jobid': jobid}
        timeout = int(timeout)
        pollperiod = 2
        progress = 1
        while timeout > 0:
            print '\r' + '.' * progress,
            sys.stdout.flush()
            time.sleep(pollperiod)
            timeout = timeout - pollperiod
            progress += 1
            logger_debug(logger, "Job %s to timeout in %ds" % (jobid, timeout))
            response, error = make_request(command, request, logger, url,
                                           credentials, expires, region)

            if error is not None:
                return response, error

            response = process_json(response)
            responsekeys = filter(lambda x: 'response' in x, response.keys())

            if len(responsekeys) < 1:
                continue

            result = response[responsekeys[0]]
            jobstatus = result['jobstatus']
            if jobstatus == 2:
                jobresult = result["jobresult"]
                error = "\rAsync job %s failed\nError %s, %s" % (
                        jobid, jobresult["errorcode"], jobresult["errortext"])
                return response, error
            elif jobstatus == 1:
                print "\r" + " " * progress,
                return response, error
            else:
                logger_debug(logger, "We should not arrive here!")
                sys.stdout.flush()

        error = "Error: Async query timeout occurred for jobid %s" % jobid

    return response, error
