server=$1
user="ec2-user"

if [ $# -eq 0 ]
	then
		echo "Need server name"
fi

ssh -i ~/.ssh/skc-server.pem "${user}@${server}" << EOF
	cd skc-suggestion-engine
	docker-compose kill
	docker-compose rm -f
	docker-compose up -d
EOF