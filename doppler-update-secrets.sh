# Upload - DEV
doppler secrets upload -p skc-suggestion-engine -c dev .env
doppler secrets upload -p skc-suggestion-engine -c dev_docker .env_docker_local

cat certs/skc-suggestion-engine-db.pem | doppler secrets set -p skc-suggestion-engine -c dev MONGODB_X509
cat certs/skc-suggestion-engine-db.pem | doppler secrets set -p skc-suggestion-engine -c dev_docker MONGODB_X509

# Upload - Prod
doppler secrets upload -p skc-suggestion-engine -c prod .env_docker
cat certs/skc-suggestion-engine-db.pem| doppler secrets set -p skc-suggestion-engine -c prod MONGODB_X509