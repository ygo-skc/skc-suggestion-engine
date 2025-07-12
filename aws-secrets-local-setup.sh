mkdir -p certs

aws secretsmanager get-secret-value --secret-id "/prod/skc/suggestion-engine/ssl" --region us-east-2 |
	jq -r '.SecretString' |
	jq -r "with_entries(select(.key | startswith(\"SSL\")))" >certs/base64-certs-json
jq -r ".SSL_PRIVATE_KEY" <certs/base64-certs-json | base64 -d >certs/private.key
jq -r ".SSL_CA_BUNDLE_CRT" <certs/base64-certs-json | base64 -d >certs/ca_bundle.crt
jq -r ".SSL_CERTIFICATE_CRT" <certs/base64-certs-json | base64 -d >certs/certificate.crt

aws secretsmanager get-secret-value --secret-id "/prod/skc/suggestion-engine/db" --region us-east-2 |
	jq -r '.SecretString' >certs/base64-certs-json
jq -r ".DB_PEM" <certs/base64-certs-json | base64 -d >certs/skc-suggestion-engine-db.pem

rm certs/base64-certs-json

#############################################
createEnvFile() {
	local FILE_NAME=$1
	local ENV_SECRETS_ID=$2

	aws secretsmanager get-secret-value --secret-id "${ENV_SECRETS_ID}" --region us-east-2 |
		jq -r '.SecretString' | jq -r ". + $DB_HOST | to_entries|map(\"\(.key)=\\\"\(.value|tostring)\\\"\")|.[]" >"$FILE_NAME"
}

DB_HOST=$(aws secretsmanager get-secret-value --secret-id "/prod/skc/suggestion-engine/db" --region us-east-2 |
	jq -r '.SecretString' |
	jq -c "with_entries(select(.key | startswith(\"DB_HOST\")))")

createEnvFile ".env_prod" "/prod/skc/suggestion-engine/env"
createEnvFile ".env" "/dev/skc/suggestion-engine/env"
createEnvFile ".env_docker_local" "/prod/skc/suggestion-engine/env"
