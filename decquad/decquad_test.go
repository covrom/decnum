package decquad

import (
	"reflect"
	"testing"
)

func TestDecFloatFromString(t *testing.T) {
	type args struct {
		s   string
		set *DecContext
	}
	tests := []struct {
		name    string
		args    args
		want    DecQuad
		wantErr bool
	}{
		// {
		// 	args: args{
		// 		s:   "123.45",
		// 		set: &DecContext{},
		// 	},
		// 	want: DecQuad([4]uint32{0x000049c5, 0x00000000, 0x00000000, 0x22078000}),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecFloatFromString(tt.args.s, tt.args.set)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecFloatFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecFloatFromString() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
