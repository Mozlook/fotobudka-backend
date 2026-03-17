package auth

import "github.com/golang-jwt/jwt/v5"

// Claims contains the JWT claims used by the FotoBudka backend.
type Claims struct {
	jwt.RegisteredClaims
}
