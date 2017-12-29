package api

import (
	"reflect"
	"testing"

	"github.com/ddecaro94/gobalancer/config"
)

func TestManager_Start(t *testing.T) {
	tests := []struct {
		name string
		m    *Manager
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.Start()
		})
	}
}

func TestNewManager(t *testing.T) {
	type args struct {
		c *config.Config
	}
	tests := []struct {
		name  string
		args  args
		wantM *Manager
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotM := NewManager(tt.args.c); !reflect.DeepEqual(gotM, tt.wantM) {
				t.Errorf("NewManager() = %v, want %v", gotM, tt.wantM)
			}
		})
	}
}
