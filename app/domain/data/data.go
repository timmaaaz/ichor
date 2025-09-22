package data

import (
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

type App struct {
	auth    *auth.Auth
	databus *tablebuilder.ConfigStore
}

func NewApp(configStore *tablebuilder.ConfigStore) *App {
	return &App{
		databus: configStore,
	}
}

func NewAppWithAuth(configStore *tablebuilder.ConfigStore, ath *auth.Auth) *App {
	return &App{
		auth:    ath,
		databus: configStore,
	}
}
