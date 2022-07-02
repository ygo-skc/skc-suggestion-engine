server=$1
user="ec2-user"
dirOnServer="skc-suggestion-engine"

if [ $# -eq 0 ]
	then
		echo "Need server name"
fi

ssh -i ~/.ssh/skc-server.pem "${user}@${server}" << EOF
	mkdir $dirOnServer
EOF

sftp -i ~/.ssh/skc-server.pem "${user}@${server}" << EOF
	cd $dirOnServer
	put docker-compose.yaml
	put -r api/
	put -r db/
	put -r downstream/
	put -r env/
	put main.go
	put go.mod
	put go.sum
	put .env_docker
EOF