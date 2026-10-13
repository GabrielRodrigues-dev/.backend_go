package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Usada para passar dados de forma segura.
type ProviderKey string

const ContextProviderKey ProviderKey = "provider_id"

// AuthMiddleware intercepta a requisição, valida o token e extrai a identidade.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing or invalid token", http.StatusUnauthorized)
			return
		}

		// Remove a palavra "Bearer " para ficar só com a string do token
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Analisa o JWT.
		// Extrai as 'claims'.
		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			http.Error(w, "Unauthorized: Malformed token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized: Invalid claims", http.StatusUnauthorized)
			return
		}

		// O Keycloak pode enviar o ID do cliente em 'clientId', 'client_id' ou 'azp'
		var clientId string
		if id, ok := claims["clientId"].(string); ok {
			clientId = id
		} else if id, ok := claims["client_id"].(string); ok {
			clientId = id
		} else if id, ok := claims["azp"].(string); ok {
			clientId = id
		}

		if clientId == "" {
			http.Error(w, "Unauthorized: Missing clientId in token", http.StatusUnauthorized)
			return
		}

		fmt.Printf(">> Porteiro: Acesso liberado para o provedor [%s]\n", clientId)

		// Injeta o ID do provedor na requisição para o nosso Handler usar depois.
		ctx := context.WithValue(r.Context(), ContextProviderKey, clientId)

		// Passa a requisição modificada adiante
		next(w, r.WithContext(ctx))
	}
}
