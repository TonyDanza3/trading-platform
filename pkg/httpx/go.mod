module trading-platform/pkg/httpx

go 1.23.0

require (
	github.com/google/uuid v1.6.0
	trading-platform/pkg/web v0.0.0
)

replace trading-platform/pkg/web => ../web
