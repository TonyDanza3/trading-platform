module trading-platform/services/api-gateway

go 1.23.0

require (
	trading-platform/pkg/httpx v0.0.0
	trading-platform/pkg/web v0.0.0
)

require github.com/google/uuid v1.6.0 // indirect

replace (
	trading-platform/pkg/httpx => ../../pkg/httpx
	trading-platform/pkg/web => ../../pkg/web
)
