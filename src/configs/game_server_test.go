package configs

import (
	"reflect"
	"testing"
)

func TestGameServerConfigs_Format(t *testing.T) {
	type fields struct {
		Port     int
		Name     string
		IP       string
		Location string
		Config   Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameServerConfigs := &GameServerConfigs{
				Port:     tt.fields.Port,
				Name:     tt.fields.Name,
				IP:       tt.fields.IP,
				Location: tt.fields.Location,
				Config:   tt.fields.Config,
			}
			if got := gameServerConfigs.Format(); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGameServerConfigs(t *testing.T) {
	tests := []struct {
		name string
		want GameServerConfigs
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGameServerConfigs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGameServerConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetServerVocations(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetServerVocations(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServerVocations() = %v, want %v", got, tt.want)
			}
		})
	}
}
