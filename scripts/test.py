from sshtunnel import SSHTunnelForwarder

server = SSHTunnelForwarder(
    'int-ugc-postgres.c66kz9sr8urn.eu-west-2.rds.amazonaws.com',
    remote_bind_address=('10.82.152.96,eu-west-2',5432)
)

server.start()

print(server.local_bind_port)  # show assigned local port
# work with `SECRET SERVICE` through `server.local_bind_port`.

server.stop()
