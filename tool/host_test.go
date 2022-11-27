package tool

import "testing"

func Test_isValidIP(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{
				addr: "www.baidu.com",
			},
			want: false,
		},
		{
			name: "test",
			args: args{
				addr: "192.168.1.1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := isValidIP(tt.args.addr)
			t.Logf("b=%#v", b)
		})
	}
}
