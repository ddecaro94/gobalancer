package config

import (
	"reflect"
	"testing"
)

func TestReadConfigJSON(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantC   *Config
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, err := ReadConfigJSON(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfigJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("ReadConfigJSON() = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}

func TestReadConfigYAML(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantC   *Config
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, err := ReadConfigYAML(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfigYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("ReadConfigYAML() = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}

func TestConfig_Reload(t *testing.T) {
	tests := []struct {
		name    string
		c       *Config
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Reload(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Reload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCluster_Add(t *testing.T) {
	type args struct {
		s Server
	}
	tests := []struct {
		name       string
		c          *Cluster
		args       args
		wantResult bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.c.Add(tt.args.s); gotResult != tt.wantResult {
				t.Errorf("Cluster.Add() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestCluster_Update(t *testing.T) {
	type args struct {
		s Server
	}
	tests := []struct {
		name        string
		c           *Cluster
		args        args
		wantUpdated int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUpdated := tt.c.Update(tt.args.s); gotUpdated != tt.wantUpdated {
				t.Errorf("Cluster.Update() = %v, want %v", gotUpdated, tt.wantUpdated)
			}
		})
	}
}
