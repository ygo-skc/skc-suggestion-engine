if [ $# -eq 0 ]
	then
		echo "Need server name"
		exit 1
fi

SERVER=$1
USER="ec2-user"
DIR_ON_SERVER="skc-suggestion-engine"

echo "Using server $SERVER and directory $DIR_ON_SERVER to restage api"

ssh -i ~/.ssh/skc-server.pem "${user}@${server}" << EOF
	cd $DIR_ON_SERVER
	docker-compose kill
	docker-compose rm -f
	docker-compose pull
	docker-compose up -d
EOF