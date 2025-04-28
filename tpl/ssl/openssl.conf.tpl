
[req]
default_bits       = 2048
default_md         = sha256
distinguished_name = req_distinguished_name
req_extensions     = v3_req
prompt            = no

[req_distinguished_name]
C  = SG
ST = Singapore
L  = Singapore
O  = LocalServer
OU = Server
CN = local_server

[v3_req]
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
