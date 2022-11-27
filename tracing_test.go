package net

import (
	"context"
	"testing"

	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

func Test_configAgent(t *testing.T) {
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
	app := New(opts...)
	type args struct {
		ctx        context.Context
		agentAddr  string
		service    string
		version    string
		attributes []KeyValue
	}
	tests := []struct {
		name    string
		args    args
		want    *sdkTrace.TracerProvider
		wantErr bool
	}{
		{
			name: "test config agent",
			args: args{
				ctx:        app.ctx,
				agentAddr:  app.opts.agentAddr,
				service:    app.opts.name,
				version:    app.opts.version,
				attributes: app.opts.attributes,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := configAgent(tt.args.ctx, tt.args.agentAddr, tt.args.service, tt.args.version, tt.args.attributes...)
			if err != nil {
				t.Fatal(err)
			}
			ctx, s := got.Tracer("trace").Start(tt.args.ctx, "iSpan")
			t.Logf("ctx = %#v", ctx)
			t.Logf("s = %s", s)
			t.Logf("s trace=%#v", TraceID(ctx))
			t.Logf("span=%#v", SpanID(ctx))
			span := Span(ctx)
			t.Logf("span.id=%#v trace.id=%#v", span.SpanContext().SpanID().String(), span.SpanContext().TraceID().String())
		})
	}
}
