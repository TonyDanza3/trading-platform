module trading-platform/pkg/authjwt

go 1.23.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	trading-platform/pkg/web v0.0.0
)

replace trading-platform/pkg/web => ../web
