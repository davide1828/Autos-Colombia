.PHONY: db test test-integration lint build run
GO := cd backend && go

db:                       ## base de datos de desarrollo
	docker compose up -d db
test:                     ## pruebas de backend (con race) y frontend
	$(GO) test -race -cover ./... && cd frontend && npm test
test-integration:         ## incluye PostgreSQL real (base desechable)
	cd backend && TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" go test -race -count=1 ./...
lint:
	@test -z "$$(cd backend && gofmt -l .)" || (echo "gofmt pendiente"; exit 1)
	$(GO) vet ./... && cd frontend && npm run lint
build:                    ## frontend -> backend/static
	cd frontend && npm ci && npm run build && rm -rf ../backend/static && cp -r dist ../backend/static
run: build
	cd backend && API_TOKEN=$${API_TOKEN:?defina API_TOKEN} go run ./cmd/api
