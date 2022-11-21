# Download - Dev
doppler secrets download -p skc-suggestion-engine -c dev --no-file --format env > .env
doppler secrets download -p skc-suggestion-engine -c dev_docker --no-file --format env > .env_docker_local

# Download - Prod
doppler secrets download -p skc-suggestion-engine -c prod --no-file --format env > .env_docker

# Download Certs
mkdir certs
doppler secrets get -p skc-suggestion-engine -c prod MONGODB_X509 --plain  > certs/skc-suggestion-engine-db.pem

doppler secrets get -p skc-suggestion-engine -c prod SSL_CERTIFICATE --plain  > certs/certificate.crt
doppler secrets get -p skc-suggestion-engine -c prod SSL_PRIVATE_KEY --plain  > certs/private.key
doppler secrets get -p skc-suggestion-engine -c prod SSL_CA_BUNDLE --plain  > certs/ca_bundle.crt