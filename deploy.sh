if [ $# -eq 0 ]
	then
		echo "Need server name"
		exit 1
fi

SERVER=$1
USER="ec2-user"
DIR_ON_SERVER="api/skc-suggestion-engine"

echo "Building API"
env GOOS=linux GOARCH=arm64 go build .

echo "Using server $SERVER and directory $DIR_ON_SERVER to sync prod API"

echo "Uploading API files"
rsync --rsync-path="mkdir -p ${DIR_ON_SERVER} && rsync" -avzh --delete --progress -e "ssh -i ~/.ssh/skc-server.pem" skc-suggestion-engine data certs .env_prod docker-compose.yaml "${USER}@${SERVER}:${DIR_ON_SERVER}/"

echo -e "\n\nRestaging API"
ssh -i ~/.ssh/skc-server.pem "${USER}@${SERVER}" << EOF
	cd $DIR_ON_SERVER
	docker-compose kill
	docker-compose rm -f
	docker-compose pull
	docker-compose up -d
EOF

bash aws-cert-update.sh