#!/usr/bin/env python
import json
import os
import sys
import stat

subprocess = os.popen("aws sts assume-role-with-web-identity --role-arn $AWS_ROLE_ARN --role-session-name mh9test --web-identity-token file://$AWS_WEB_IDENTITY_TOKEN_FILE --duration-second 3600")

try:
  creds = json.loads(subprocess.read())
except:
  sys.stderr.write("Could not read JSON")
  exit(1)

print(creds)

with open('env.sh', 'w') as f:
    f.write('export AWS_ACCESS_KEY_ID={0}\n'.format(creds['Credentials']['AccessKeyId']))
    f.write('export AWS_SECRET_ACCESS_KEY={0}\n'.format(creds['Credentials']['SecretAccessKey']))
    f.write('export AWS_SESSION_TOKEN={0}\n'.format(creds['Credentials']['SessionToken']))
    f.write('export AWS_DEFAULT_REGION=eu-west-2\n')

st = os.stat('env.sh')
os.chmod('env.sh', st.st_mode | stat.S_IEXEC)