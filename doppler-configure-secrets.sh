# Download - Dev
doppler secrets download -p skc-suggestion-engine -c dev --no-file --format env > .env
doppler secrets download -p skc-suggestion-engine -c dev_docker --no-file --format env > .env_docker_local

# Download - Prod
doppler secrets download -p skc-suggestion-engine -c prod --no-file --format env > .env_docker