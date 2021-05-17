package models

import (
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"reflect"
	"testing"
)

func TestBuildWorldsMessage(t *testing.T) {
	type args struct {
		gameConfigs configs.GameServerConfigs
		worldId     int
	}
	tests := []struct {
		name string
		args args
		want []*login_proto_messages.World
	}{{
		name: "build_default_worlds_message_id_0",
		args: args{gameConfigs: configs.GameServerConfigs{
			Port:     defaultNumber,
			Name:     defaultString,
			IP:       defaultString,
			Location: defaultString,
		}, worldId: 11},
		want: []*login_proto_messages.World{{
			Id:                         uint32(0),
			ExternalAddress:            defaultString,
			ExternalAddressProtected:   defaultString,
			ExternalAddressUnprotected: defaultString,
			ExternalPort:               uint32(defaultNumber),
			ExternalPortProtected:      uint32(defaultNumber),
			ExternalPortUnprotected:    uint32(defaultNumber),
			Location:                   defaultString,
			Name:                       defaultString,
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildWorldsMessage(tt.args.gameConfigs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildWorldsMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadWorldsFromMessage(t *testing.T) {
	type args struct {
		worldsMsg []*login_proto_messages.World
	}
	tests := []struct {
		name string
		args args
		want []World
	}{{
		"one_world_id_10",
		args{[]*login_proto_messages.World{createWorldMessage(10)}},
		[]World{{
			ID:                         10,
			ExternalPort:               defaultNumber,
			ExternalPortProtected:      defaultNumber,
			ExternalPortUnprotected:    defaultNumber,
			ExternalAddress:            defaultString,
			ExternalAddressProtected:   defaultString,
			ExternalAddressUnprotected: defaultString,
			Location:                   defaultString,
			Name:                       defaultString,
		}}}, {
		"one_world_id_5",
		args{[]*login_proto_messages.World{createWorldMessage(5)}},
		[]World{{
			ID:                         5,
			ExternalPort:               defaultNumber,
			ExternalPortProtected:      defaultNumber,
			ExternalPortUnprotected:    defaultNumber,
			ExternalAddress:            defaultString,
			ExternalAddressProtected:   defaultString,
			ExternalAddressUnprotected: defaultString,
			Location:                   defaultString,
			Name:                       defaultString,
		}}},
		{
			"two_worlds_different_ids",
			args{[]*login_proto_messages.World{createWorldMessage(1), createWorldMessage(5)}},
			[]World{{
				ID:                         1,
				ExternalPort:               defaultNumber,
				ExternalPortProtected:      defaultNumber,
				ExternalPortUnprotected:    defaultNumber,
				ExternalAddress:            defaultString,
				ExternalAddressProtected:   defaultString,
				ExternalAddressUnprotected: defaultString,
				Location:                   defaultString,
				Name:                       defaultString,
			}, {
				ID:                         5,
				ExternalPort:               defaultNumber,
				ExternalPortProtected:      defaultNumber,
				ExternalPortUnprotected:    defaultNumber,
				ExternalAddress:            defaultString,
				ExternalAddressProtected:   defaultString,
				ExternalAddressUnprotected: defaultString,
				Location:                   defaultString,
				Name:                       defaultString,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadWorldsFromMessage(tt.args.worldsMsg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadWorldsFromMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildWorldMessage(t *testing.T) {
	type args struct {
		gameConfigs configs.GameServerConfigs
		worldId     int
	}
	tests := []struct {
		name string
		args args
		want *login_proto_messages.World
	}{{
		name: "default_config_world_id_11",
		args: args{gameConfigs: configs.GameServerConfigs{
			Port:     defaultNumber,
			Name:     defaultString,
			IP:       defaultString,
			Location: defaultString,
		}, worldId: 11},
		want: &login_proto_messages.World{
			Id:                         uint32(11),
			ExternalAddress:            defaultString,
			ExternalAddressProtected:   defaultString,
			ExternalAddressUnprotected: defaultString,
			ExternalPort:               uint32(defaultNumber),
			ExternalPortProtected:      uint32(defaultNumber),
			ExternalPortUnprotected:    uint32(defaultNumber),
			Location:                   defaultString,
			Name:                       defaultString,
		},
	}, {
		name: "mixed_configs_world_id_0",
		args: args{gameConfigs: configs.GameServerConfigs{
			Port:     7171,
			Name:     "Earth",
			IP:       "123.456.789.0",
			Location: "Solar System",
		}, worldId: 0},
		want: &login_proto_messages.World{
			Id:                         uint32(0),
			ExternalAddress:            "123.456.789.0",
			ExternalAddressProtected:   "123.456.789.0",
			ExternalAddressUnprotected: "123.456.789.0",
			ExternalPort:               uint32(7171),
			ExternalPortProtected:      uint32(7171),
			ExternalPortUnprotected:    uint32(7171),
			Location:                   "Solar System",
			Name:                       "Earth",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildWorldMessage(tt.args.gameConfigs, tt.args.worldId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildWorldMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
