module trading-platform/services/user-service

go 1.23.0

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	trading-platform/pkg/authjwt v0.0.0
	trading-platform/pkg/dbx v0.0.0
	trading-platform/pkg/httpx v0.0.0
	trading-platform/pkg/web v0.0.0
)

replace (
	trading-platform/pkg/authjwt => ../../pkg/authjwt
	trading-platform/pkg/dbx => ../../pkg/dbx
	trading-platform/pkg/httpx => ../../pkg/httpx
	trading-platform/pkg/web => ../../pkg/web
)
