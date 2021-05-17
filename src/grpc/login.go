package grpc_login_server

import (
	"context"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

const DefaultLoginErrorCode = 3

func (ls *GrpcServer) Login(ctx context.Context, in *login_proto_messages.LoginRequest) (*login_proto_messages.LoginResponse, error) {
	acc, err := database.LoadAccount(in.Email, in.Password, ls.DB)
	if err != nil {
		return &login_proto_messages.LoginResponse{
			Error: &login_proto_messages.Error{
				Code:    DefaultLoginErrorCode,
				Message: err.Error(),
			},
		}, nil
	}

	players := &database.Players{AccountID: acc.ID}

	err = database.LoadPlayers(ls.DB, players)
	if err != nil {
		return nil, err
	}

	return buildLoginResponsePayload(acc.GetGrpcSession(), *players), nil
}

func buildLoginResponsePayload(
	session *login_proto_messages.Session,
	players database.Players,
) *login_proto_messages.LoginResponse {
	var characters []*login_proto_messages.Character
	for _, player := range players.Players {
		character := player.ToCharacterMessage()
		characters = append(characters, character)

		if session.LastLogin < character.Info.LastLogin {
			session.LastLogin = character.Info.LastLogin
		}
	}

	return &login_proto_messages.LoginResponse{
		PlayData: &login_proto_messages.PlayData{
			Characters: characters,
			Worlds:     loadWorldsMessage(),
		},
		Session: session,
	}
}

func loadWorldsMessage() []*login_proto_messages.World {
	return []*login_proto_messages.World{loadWorldMessage(configs.GetGameServerConfigs())}
}

func loadWorldMessage(gameConfigs configs.GameServerConfigs) *login_proto_messages.World {
	return &login_proto_messages.World{
		ExternalAddress:            gameConfigs.IP,
		ExternalAddressProtected:   gameConfigs.IP,
		ExternalAddressUnprotected: gameConfigs.IP,
		ExternalPort:               uint32(gameConfigs.Port),
		ExternalPortProtected:      uint32(gameConfigs.Port),
		ExternalPortUnprotected:    uint32(gameConfigs.Port),
		Location:                   gameConfigs.Location,
		Name:                       gameConfigs.Name,
	}
}
