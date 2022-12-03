package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestDial(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts []ClientOption
	}
	tests := []struct {
		name    string
		args    args
		want    *grpc.ClientConn
		wantErr bool
	}{
		{
			name: "test dial tcp address",
			args: args{
				ctx: context.Background(),
				opts: []ClientOption{
					WithEndpoint("user.svc.cluster.local:8100"),
					WithInSecure(true),
					WithTracing(true),
					WithOptions(),
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Dial(tt.args.ctx, tt.args.opts...)
			if err != nil {
				t.Fatal(err)
				return
			}
			// orderGrpcClient := order.NewOrderClient(*grpc.ClientConn g)
		})
	}
}
