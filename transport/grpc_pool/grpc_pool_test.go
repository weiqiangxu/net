package grpc_pool

import (
	"context"
	"testing"

	netGrpc "github.com/weiqiangxu/net/transport/grpc"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	type args struct {
		poolConfig *Config
	}
	tests := []struct {
		name    string
		args    args
		want    Pool
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				poolConfig: &Config{
					InitialCap: 5,
					MaxCap:     10,
					Factory: func() (*grpc.ClientConn, error) {
						return netGrpc.Dial(context.Background())
					},
					Close:       nil,
					Ping:        nil,
					IdleTimeout: 0,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.poolConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			clientConn, _ := got.Get()
			t.Logf("%+v", clientConn)
		})
	}
}
