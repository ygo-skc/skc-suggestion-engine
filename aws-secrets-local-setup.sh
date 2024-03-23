mkdir -p certs

aws secretsmanager get-secret-value --secret-id "/prod/skc/suggestion-engine/ssl" --region us-east-2 \
  | jq -r '.SecretString' \
  | jq -r "with_entries(select(.key | startswith(\"SSL\")))" > certs/base64-certs-json
jq -r ".SSL_PRIVATE_KEY" < certs/base64-certs-json | base64 -d > certs/private.key
jq -r ".SSL_CA_BUNDLE_CRT" < certs/base64-certs-json | base64 -d > certs/ca_bundle.crt
jq -r ".SSL_CERTIFICATE_CRT" < certs/base64-certs-json | base64 -d > certs/certificate.crt

aws secretsmanager get-secret-value --secret-id "/prod/skc/suggestion-engine/db" --region us-east-2 \
  | jq -r '.SecretString'  > certs/base64-certs-json
jq -r ".DB_PEM" < certs/base64-certs-json | base64 -d > certs/skc-suggestion-engine-db.pem

rm certs/base64-certs-json

#############################################
createEnvFile() {
  local SKC_API_DB_INFO=$1
  local FILE_NAME=$2

	aws secretsmanager get-secret-value --secret-id "/prod/skc/suggestion-engine/env" --region us-east-2 \
	| jq -r '.SecretString' | jq -r ". + $DB_HOST + $SKC_API_DB_INFO | to_entries|map(\"\(.key)=\\\"\(.value|tostring)\\\"\")|.[]" > "$FILE_NAME"
}

DB_HOST=$(aws secretsmanager get-secret-value --secret-id "/prod/skc/suggestion-engine/db" --region us-east-2 \
  | jq -r '.SecretString' \
  | jq -c "with_entries(select(.key | startswith(\"DB_HOST\")))")

SKC_API_DB_INFO=$(aws secretsmanager get-secret-value --secret-id "/prod/skc/skc-api/db" --region us-east-2 \
  | jq -r '.SecretString' \
  | jq -c "with_entries(select(.key | startswith(\"DB\")) | .key |= \"SKC_\(.)\")")
createEnvFile "$SKC_API_DB_INFO" ".env_prod"

SKC_API_DB_INFO=$(aws secretsmanager get-secret-value --secret-id "/local/skc/skc-api/db" --region us-east-2 \
  | jq -r '.SecretString' \
  | jq -c "with_entries(select(.key | startswith(\"DB\")) | .key |= \"SKC_\(.)\")")
createEnvFile "$SKC_API_DB_INFO" ".env"

SKC_API_DB_INFO=$(aws secretsmanager get-secret-value --secret-id "/docker/local/skc/skc-api/db" --region us-east-2 \
  | jq -r '.SecretString' \
  | jq -c "with_entries(select(.key | startswith(\"DB\")) | .key |= \"SKC_\(.)\")")
createEnvFile "$SKC_API_DB_INFO" ".env_docker_local"