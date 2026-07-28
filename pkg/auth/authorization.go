package auth

import (
	"context"
	"error"

	"encore.com/workspaces/encore-workspaces-proxy/pkg/encore"
)

var ErrInvalidUser = errors.New("usuário não possui acesso a este workspace")

func checkAuthorization(ctx context.Context, accessToken string, workspaceID string, apiFactory encore.APIFactory) error {
	api := apiFactory(accessToken)

	currentUser, err := api.GetUserInfo(ctx)
	if err != nil {
		return err
	}

	workspaceInfo, err := api.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	if currentUser.ID != workspaceInfo.User.ID {
		return ErrInvalidUser
	}

	return nil
}
