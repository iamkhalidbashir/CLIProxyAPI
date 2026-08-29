package store

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

type podIdentityProvider struct {
	iam       *credentials.IAM
	tokenFile string
}

func (p *podIdentityProvider) Retrieve() (credentials.Value, error) {
	token, err := os.ReadFile(p.tokenFile)
	if err != nil {
		return credentials.Value{}, fmt.Errorf("object store: read pod identity token: %w", err)
	}
	p.iam.Container.AuthorizationToken = strings.TrimSpace(string(token))
	return p.iam.Retrieve()
}

func (p *podIdentityProvider) IsExpired() bool {
	return p.iam.IsExpired()
}

func objectStoreCredentials(cfg ObjectStoreConfig) (*credentials.Credentials, error) {
	hasAccessKey := cfg.AccessKey != ""
	hasSecretKey := cfg.SecretKey != ""
	if hasAccessKey != hasSecretKey {
		return nil, fmt.Errorf("object store: access key and secret key must be configured together")
	}
	if hasAccessKey {
		return credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""), nil
	}

	credentialsURI := strings.TrimSpace(os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI"))
	tokenFile := strings.TrimSpace(os.Getenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE"))
	if credentialsURI == "" || tokenFile == "" {
		return nil, fmt.Errorf("object store: EKS Pod Identity credential URI and token file are required when static keys are absent")
	}
	iam := &credentials.IAM{Client: http.DefaultClient}
	iam.Container.CredentialsFullURI = credentialsURI
	return credentials.New(&podIdentityProvider{iam: iam, tokenFile: tokenFile}), nil
}
