# below expression will omit certain secrets as they are not needed and since they contain spaces will mess up the .env file API expects
JQ_EXPRESSION='with_entries(select(.key | startswith("DOPPLER") | not) | select([.key] | inside(["MONGODB_X509", "SSL_CERTIFICATE", "SSL_CA_BUNDLE", "SSL_PRIVATE_KEY"]) | not)) | to_entries[] | "\(.key)=\"\(.value)\""'

# Download - Dev
doppler secrets download -p skc-suggestion-engine -c dev --no-file --format json | jq -r "$JQ_EXPRESSION" > .env
doppler secrets download -p skc-suggestion-engine -c dev_docker --no-file --format json | jq -r "$JQ_EXPRESSION" > .env_docker_local

# Download - Prod
doppler secrets download -p skc-suggestion-engine -c prod --no-file --format json | jq -r "$JQ_EXPRESSION" > .env_prod

# Download Certs
mkdir -p certs
doppler secrets get -p skc-suggestion-engine -c prod MONGODB_X509 --plain  > certs/skc-suggestion-engine-db.pem

doppler secrets get -p skc-suggestion-engine -c prod SSL_CERTIFICATE --plain  > certs/certificate.crt
doppler secrets get -p skc-suggestion-engine -c prod SSL_PRIVATE_KEY --plain  > certs/private.key
doppler secrets get -p skc-suggestion-engine -c prod SSL_CA_BUNDLE --plain  > certs/ca_bundle.crt

