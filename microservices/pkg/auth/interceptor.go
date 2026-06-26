package authdomain

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// APIKeyCheckStreamGRPCInterceptor validates a shared API key from incoming gRPC metadata for streaming endpoints.
func APIKeyCheckStreamGRPCInterceptor(validAPIKey string) grpc.StreamServerInterceptor {
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

// APIKeyBindStreamGRPCInterceptor injects the API key into the outgoing metadata of all streaming RPC calls.
func APIKeyBindStreamGRPCInterceptor(apiKey string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", apiKey)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
