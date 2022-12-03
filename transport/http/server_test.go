package http

import (
	"reflect"
	"testing"
)

func TestNewServer(t *testing.T) {
	type args struct {
		opts []ServerOption
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "new server",
			args: args{
				opts: []ServerOption{
					WithMiddleware(),
					WithAddress("127.0.0.1"),
					WithPrometheus(true),
					WithProfile(true),
					WithTracing(true),
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
