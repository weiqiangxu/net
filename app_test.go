package net

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	kvList := []KeyValue{
		{
			Key:   "name",
			Value: "jack",
		},
	}
	opts := []Option{
		ID("1"),
		Name("one"),
		Version("v0.0.1"),
		Context(context.Background()),
		Tracing("192.168.1.1", kvList...),
	}
	type args struct {
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want *App
	}{
		{
			name: "get app instance",
			args: args{
				opts: opts,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.opts...)
			t.Logf("%#v", got)
			t.Logf("a.opts.agentAddr =%#v", got.opts.agentAddr)
			t.Logf("trace id=%#v", TraceID(got.ctx))
			t.Logf("span id=%#v", SpanID(got.ctx))
			t.Logf("span =%#v", Span(got.ctx))
		})
	}
}
