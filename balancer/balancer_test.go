package balancer

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/ddecaro94/gobalancer/config"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	type args struct {
		c        *config.Config
		logger   *zap.Logger
		frontend string
	}
	tests := []struct {
		name  string
		args  args
		wantP *Balancer
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotP := New(tt.args.c, tt.args.logger, tt.args.frontend); !reflect.DeepEqual(gotP, tt.wantP) {
				t.Errorf("New() = %v, want %v", gotP, tt.wantP)
			}
		})
	}
}

func TestBalancer_ServeHTTP(t *testing.T) {
	type args struct {
		resp http.ResponseWriter
		req  *http.Request
	}
	tests := []struct {
		name string
		p    *Balancer
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.ServeHTTP(tt.args.resp, tt.args.req)
		})
	}
}

func TestBalancer_Next(t *testing.T) {
	tests := []struct {
		name       string
		p          *Balancer
		wantServer config.Server
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotServer := tt.p.Next(); !reflect.DeepEqual(gotServer, tt.wantServer) {
				t.Errorf("Balancer.Next() = %v, want %v", gotServer, tt.wantServer)
			}
		})
	}
}

func Test_codeToBounce(t *testing.T) {
	type args struct {
		code int
		list []int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := codeToBounce(tt.args.code, tt.args.list); got != tt.want {
				t.Errorf("codeToBounce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_forward(t *testing.T) {
	type args struct {
		w   http.ResponseWriter
		res *http.Response
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forward(tt.args.w, tt.args.res)
		})
	}
}

func Test_getWeightedIndex(t *testing.T) {
	type args struct {
		cluster *config.Cluster
	}
	tests := []struct {
		name      string
		args      args
		wantIndex int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIndex := getWeightedIndex(tt.args.cluster); gotIndex != tt.wantIndex {
				t.Errorf("getWeightedIndex() = %v, want %v", gotIndex, tt.wantIndex)
			}
		})
	}
}
