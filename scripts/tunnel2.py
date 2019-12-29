from sshtunnel import SSHTunnelForwarder,create_logger
import psycopg2
import logging
import paramiko

logger = logging.getLogger('sshtunnel')
logger.setLevel(logging.DEBUG)

 
proxy = paramiko.ProxyCommand('ssh -g -q -p 22000 bastion-tunnel@2.access.eu-west-2.cloud.bbc.co.uk nc 10.82.152.96 22')
s = SSHTunnelForwarder(
    ('10.82.152.96,eu-west-2'),
    ssh_username="kodjo_afriyie01",
     ssh_private_key="~/.ssh/id_rsa",
     ssh_proxy=proxy, 
    ssh_config_file='~/.ssh/config',  
     logger=create_logger(loglevel=10),
    remote_bind_address=('int-ugc-postgres.c65kz9sr8urn.eu-west-2.rds.amazonaws.com', 5432)
)   
s.start()
print("****SSH Tunnel Established****:"+str(tunnel))
 
connection = psycopg2.connect(user = "ugc-cleaner",
                                  password = "",
                                  host = "127.0.0.1",
                                  port = s.local_bind_port,
                                  database = "ugc")

try:
       campaignId = 'u22249586'
       cursor = connection.cursor()
       cursor.execute("select id, submission_id from ugc_schema.file where submission_id  in (select id from ugc_schema.submission where campaign = %s);", (campaignId,))

       for row in cursor:
         print(row)

finally:
        connection.close()
 
print("YAYY!!")

