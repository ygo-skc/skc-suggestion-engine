if [ $# -eq 0 ]
	then
		echo "Need server name"
		exit 1
fi

SERVER=$1
USER="ec2-user"
DIR_ON_SERVER="skc-suggestion-engine"

echo "Using server $SERVER and directory $DIR_ON_SERVER to upload API files"
rsync -avz --progress -e "ssh -i ~/.ssh/skc-server.pem" docker-compose.yaml api certs db model util validation data main.go go.mod go.sum .env_docker "${USER}@${SERVER}:${DIR_ON_SERVER}/"