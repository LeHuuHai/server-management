package authdomain

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor checks if the user's role has the required scope for the gRPC method.
func AuthInterceptor(methodScopes map[string]Scope) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 1. Check if the method requires authorization
		requiredScope, exists := methodScopes[info.FullMethod]
		if !exists {
			// If not listed in methodScopes, bypass auth check (permit by default)
			return handler(ctx, req)
		}

		// 2. Extract gRPC metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing incoming metadata context")
		}

		roles := md.Get("x-user-role")
		if len(roles) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing user role in metadata")
		}

		// 3. Resolve role and verify it has the required scope
		userRole := Role(roles[0])
		userScopes := userRole.Scopes()

		authorized := false
		for _, s := range userScopes {
			if s == requiredScope {
				authorized = true
				break
			}
		}

		if !authorized {
			return nil, status.Errorf(
				codes.PermissionDenied,
				"user with role %q does not have the required scope %q for method %q",
				userRole,
				requiredScope,
				info.FullMethod,
			)
		}

		// 4. Authorized, proceed to execute actual RPC method
		return handler(ctx, req)
	}
}

// StreamAPIKeyInterceptor validates a shared API key from incoming gRPC metadata for streaming endpoints.
func StreamAPIKeyInterceptor(validAPIKey string) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 1. Extract metadata from stream context
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing incoming metadata context")
		}

		// 2. Retrieve the API key from metadata
		keys := md.Get("x-api-key")
		if len(keys) == 0 {
			return status.Error(codes.Unauthenticated, "missing API key in metadata")
		}

		// 3. Verify the API key
		if keys[0] != validAPIKey {
			return status.Error(codes.PermissionDenied, "invalid API key")
		}

		// 4. Authorized, proceed with stream handler
		return handler(srv, ss)
	}
}
