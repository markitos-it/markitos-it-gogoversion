go test ./...                    # todos
go test ./... -v                 # verbose
go test ./... -v -run TestBump   # solo los de bump
go test ./... -cover             # con cobertura
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out