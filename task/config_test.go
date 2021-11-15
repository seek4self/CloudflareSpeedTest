// update v2ray config

package task

import "testing"

func Test_replaceAddr(t *testing.T) {
	tests := []struct {
		name string
		args string
	}{
		// TODO: Add test cases.
		{"1", "192.168.1.111"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replaceAddr(tt.args)
		})
	}
}
