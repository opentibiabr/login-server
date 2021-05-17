package grpc_login_server

import (
	"database/sql"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/server"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGrpcServer_GetName(t *testing.T) {
	type fields struct {
		DB                 *sql.DB
		LoginServiceServer login_proto_messages.LoginServiceServer
		ServerInterface    server.ServerInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		"",
		fields{},
		"gRPC",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &GrpcServer{
				DB:                 tt.fields.DB,
				LoginServiceServer: tt.fields.LoginServiceServer,
				ServerInterface:    tt.fields.ServerInterface,
			}
			assert.Equal(t, tt.want, ls.GetName())
		})
	}
}
