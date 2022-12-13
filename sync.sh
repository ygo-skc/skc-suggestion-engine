server=$1
user="ec2-user"

if [ $# -eq 0 ]
then
  echo "Need server name"
fi

bash artifact-transfer.sh "$1"
bash docker-restage.sh "$1"