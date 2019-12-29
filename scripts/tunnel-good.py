from sshtunnel import SSHTunnelForwarder,create_logger
import psycopg2
import logging

logger = logging.getLogger('sshtunnel')
logger.setLevel(logging.DEBUG)
 
sh =  SSHTunnelForwarder(
    ('10.82.152.96,eu-west-2'),
    ssh_username="kodjo_afriyie01",
     ssh_private_key="~/.ssh/id_rsa",
     logger=create_logger(loglevel=10),
    local_bind_address=('', 5432),
    remote_bind_address=('mr1b3ep0y3f0w55.c66kz9sr8urn.eu-west-2.rds.amazonaws.com', 5432)
) 
sh.start()
print("****SSH Tunnel Established****:"+str(tunnel))
 
connection = psycopg2.connect(user = "ugc-cleaner",
                                  password = "",
                                  host = "127.0.0.1",
                                  port = "5432",
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

