package encore

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hasura/go-graphql-client"
	"go.uber.org/zap"
)

type API interface {
	GetUserInfo(ctx context.Context) (*User, error)
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
}

type APIFactory func(accessToken string) API

type Client struct {
	accessToken string
	baseURL     string
	tokenType   TokenType
	gglClient   *graphql.Client
}

type TokenType int

const (
	BearerTokenType TokenType = iota
	PrivateTokenType
)

var ErrWorkspaceNotFound = errors.New("workspace não encontrado")

type tokenTransport struct {
	Transport http.RoundTripper
	Token     string
	tokenType TokenType
	logger    *zap.Logger
}

func (att tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// operação normal utilizará token bearer. token privado é habilitado
	// para os motivos de rodar testes de integração utilizando o pat
	
	if att.tokenType == BearerTokenType {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", att.Token))
	} else {
		req.Header.Add("PRIVATE-TOKEN", att.Token)
	}

	return att.Transport.RoundTrip(req)
}

func NewClient(logger *zap.Logger, accessToken, baseURL string, tokenType TokenType) *Client {
	client := &http.Client{
		Transport: http.DefaultTransport
	}
	
	client.Transport = tokenTransport{
		logger:    logger,
		tokenType: tokenType,
		Transport: client.Transport,
		Token:     accessToken
	}

	gqlClient := graphql.NewClient(fmt.Sprintf("%s/api/graphql", baseURL), client)

	return &Client{
		accessToken: accessToken,
		baseURL:     baseURL,
		tokenType:   tokenType,
		gqlClient:   gqlClient
	}
}
