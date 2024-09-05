#.PHONY: test

test: docker-test-up run-tests docker-test-down

docker-test-up:
	docker-compose -f docker-compose.test.yml up -d
	sleep 10 # Ждем, пока база данных будет готова

run-tests:
	go test -v ./...

docker-test-down:
	docker-compose -f docker-compose.test.yml down

up:
	docker-compose -f docker-compose.yml up --build