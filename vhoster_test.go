package vhoster

import (
	"testing"
)

func TestGateway_Exists(t *testing.T) {
	type fields struct {
		pws map[string]proxyWrapper
	}
	type args struct {
		vhost string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"exists",
			fields{
				pws: map[string]proxyWrapper{
					"test.example.com": {},
				},
			},
			args{
				vhost: "test.example.com",
			},
			true,
		},
		{
			"not exists",
			fields{
				pws: map[string]proxyWrapper{
					"test.example.com": {},
				},
			},
			args{
				vhost: "foo.example.com",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gateway{
				pws: tt.fields.pws,
			}
			if got := g.Exists(tt.args.vhost); got != tt.want {
				t.Errorf("Gateway.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
