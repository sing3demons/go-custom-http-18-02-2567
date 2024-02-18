cert:
	mkdir -p certificates
	openssl genpkey -algorithm RSA -out key.pem -pkeyopt rsa_keygen_bits:2048
	openssl req -new -x509 -key key.pem -out cert.pem -days 365
	mv cert.pem certificates
	mv key.pem certificates
create-postgres:
	docker-compose up -d postgres
