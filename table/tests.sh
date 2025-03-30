go test ./table_test -tags=""
go test ./table_test -tags="unsafe"
go test ./table_test -tags="schema_enabled"
go test ./table_test -tags="unsafe schema_enabled"

go test . -tags=""
go test . -tags="unsafe"
go test . -tags="schema_enabled"
go test . -tags="unsafe schema_enabled"

go test -bench=. ./table_benchmarks -tags=""
go test -bench=. ./table_benchmarks -tags="unsafe"
go test -bench=. ./table_benchmarks -tags="schema_enabled"
go test -bench=. ./table_benchmarks -tags="unsafe schema_enabled"
