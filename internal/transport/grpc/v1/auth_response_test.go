package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	"github.com/google/uuid"
	pb "github.com/weit145/proto-startup/auth"
	grpcapi "google.golang.org/grpc"
)

type authResponseClient struct {
	pb.AuthClient
	response *pb.AuthResponse
}

func (c authResponseClient) RegisterComplete(context.Context, *pb.RegisterCompleteRequest, ...grpcapi.CallOption) (*pb.AuthResponse, error) {
	return c.response, nil
}

func (c authResponseClient) Login(context.Context, *pb.LoginRequest, ...grpcapi.CallOption) (*pb.AuthResponse, error) {
	return c.response, nil
}

func (c authResponseClient) RefreshToken(context.Context, *pb.RefreshRequest, ...grpcapi.CallOption) (*pb.AuthResponse, error) {
	return c.response, nil
}

func TestAuthResponsesWithoutUsername(t *testing.T) {
	id := uuid.New()
	methods := map[string]func(*Client) (auth.User, auth.Tokens, error){
		"register_complete": func(c *Client) (auth.User, auth.Tokens, error) {
			return c.RegisterComplete(context.Background(), "test-password", uuid.New())
		},
		"login": func(c *Client) (auth.User, auth.Tokens, error) {
			return c.Login(context.Background(), "owner@example.com", "test-password")
		},
		"refresh": func(c *Client) (auth.User, auth.Tokens, error) {
			return c.RefreshToken(context.Background(), "old-refresh-token")
		},
	}
	cases := []struct {
		name       string
		invalidate func(*pb.AuthResponse)
	}{
		{name: "valid_without_username"},
		{name: "missing_user", invalidate: func(r *pb.AuthResponse) { r.User = nil }},
		{name: "missing_id", invalidate: func(r *pb.AuthResponse) { r.User.Id = "" }},
		{name: "invalid_id", invalidate: func(r *pb.AuthResponse) { r.User.Id = "invalid" }},
		{name: "missing_email", invalidate: func(r *pb.AuthResponse) { r.User.Email = "" }},
		{name: "invalid_email", invalidate: func(r *pb.AuthResponse) { r.User.Email = "invalid" }},
		{name: "missing_role", invalidate: func(r *pb.AuthResponse) { r.User.Role = "" }},
		{name: "invalid_role", invalidate: func(r *pb.AuthResponse) { r.User.Role = "unknown" }},
		{name: "missing_access_token", invalidate: func(r *pb.AuthResponse) { r.AccessToken = "" }},
		{name: "missing_refresh_token", invalidate: func(r *pb.AuthResponse) { r.RefreshToken = "" }},
	}
	for method, call := range methods {
		for _, tc := range cases {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				response := &pb.AuthResponse{
					User:         &pb.User{Id: id.String(), Email: "owner@example.com", Role: "user"},
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
				}
				if tc.invalidate != nil {
					tc.invalidate(response)
				}
				client := &Client{
					api:        authResponseClient{response: response},
					rpcTimeout: TimeSettings{Default: time.Second},
				}
				user, tokens, err := call(client)
				if tc.invalidate != nil {
					if err == nil {
						t.Fatal("invalid response accepted")
					}
					return
				}
				if err != nil {
					t.Fatalf("response without username rejected: %v", err)
				}
				if user.ID() != id || user.Email() != response.User.Email || user.Role() != response.User.Role {
					t.Fatalf("unexpected user: %+v", user)
				}
				if tokens.AccessToken().Token() != response.AccessToken || tokens.RefreshToken().Token() != response.RefreshToken {
					t.Fatal("tokens do not match the gRPC response")
				}
			})
		}
	}
}
