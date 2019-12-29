import paramiko
import os

print(os.name)
ssh_config = paramiko.SSHConfig()
# Try to read SSH_CONFIG_FILE
try:
            # open the ssh config file
            with open(os.path.expanduser("~/.ssh/config"), 'r') as f:
                ssh_config.parse(f)
            # looks for information for the destination system
            hostname_info = ssh_config.lookup("10.82.152.85")
            # gather settings for user, port and identity file
            # last resort: use the 'login name' of the user
            ssh_username = (
                hostname_info.get('User')
            )
            ssh_pkey = (
                hostname_info.get('identityfile', [None])[0]
            )
            ssh_host = hostname_info.get('hostname')
            ssh_port = hostname_info.get('port')

            proxycommand = hostname_info.get('proxycommand')
            ssh_proxy = (paramiko.ProxyCommand(proxycommand) if
                                      proxycommand else None)
            compression = hostname_info.get('compression', '')
            compression = True if compression.upper() == 'YES' else False
            print(ssh_username)
except IOError:
            if logger:
                logger.warning(
                    'Could not read SSH configuration file: {0}'
                    .format(ssh_config_file)
                )
