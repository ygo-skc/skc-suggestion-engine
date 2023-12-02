if [ $# -eq 0 ]
	then
		echo "Need server name"
		exit 1
fi

SERVER=$1
USER="ec2-user"

echo "Using server $SERVER to sync prod deployment w/ local changes"

bash artifact-transfer.sh "$1"
bash docker-restage.sh "$1"
bash aws-cert-update.sh