package things

import (
	"io"
	"testing"
)

func TestErrSet(t *testing.T) {
	type args struct {
		err   error
		key   interface{}
		value interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "key, value retreive",
			args: args{
				err:   io.EOF,
				key:   "hello",
				value: "world",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oerr := tt.args.err
			ErrSet(&tt.args.err, tt.args.key, tt.args.value)
			val := ErrGet(tt.args.err, tt.args.key)
			if val != tt.args.value {
				t.Errorf("wrong value, got %v but wanted %v", val, tt.args.value)
			}
			if !ErrIs(tt.args.err, oerr) {
				t.Errorf("ErrIs doesn't return true; error is %v and should wrap %v", tt.args.err, oerr)
			}
			if _, ok := tt.args.err.(WrappedError); !ok {
				t.Errorf("error is not wrapped: have %v instead", tt.args.err)
			}
		})
	}
}
