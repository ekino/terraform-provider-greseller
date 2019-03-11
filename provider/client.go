package provider

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	billing "google.golang.org/api/cloudbilling/v1"
)

var OAuthScopes = []string{
	billing.CloudPlatformScope,
}

// Client is a structure that will be provided to terraform as "meta"
type ProviderClient struct {
	billing *billing.APIService
}

func test() error {
	return nil
}

func apiClientFromCredentialsFile(crendentialsFile string) (*http.Client, error) {
	log.Printf("[INFO] Authenticating using provided credentials file")
	fileContent, _, err := pathorcontents.Read(crendentialsFile)
	if err != nil {
		return nil, err
	}

	jwtConfig, err := google.JWTConfigFromJSON([]byte(fileContent), OAuthScopes...)
	if err != nil {
		return nil, err
	}

	return jwtConfig.Client(context.Background()), nil
}

func apiClientFromDefault() (*http.Client, error) {
	log.Printf("[INFO] Authenticating using DefaultClient")
	return google.DefaultClient(context.Background(), OAuthScopes...)
}

func (pc *ProviderClient) init(crendentialsFile string) error {
	var apiClient *http.Client
	var err error
	
	if crendentialsFile != "" {
		apiClient, err = apiClientFromCredentialsFile(crendentialsFile)
		if err != nil {
			return errors.Wrap(err, "Authentication error, invalid credentials file")
		}
	} else {
		apiClient, err = apiClientFromDefault()
		if err != nil {
			return errors.Wrap(err, "Authentication error, failed to use default credentials")
		}
	}

	// Use a custom user-agent string. This helps google with analytics and it's
	// just a nice thing to do.
	apiClient.Transport = logging.NewTransport("Google", apiClient.Transport)
	userAgent := fmt.Sprintf("(%s %s) Terraform/%s",
		runtime.GOOS, runtime.GOARCH, terraform.VersionString())

	// Create billing service
	billingSvc, err := billing.New(apiClient)
	if err != nil {
		return errors.Wrap(err, "Couldn't instantiate billing api service")
	}

	billingSvc.UserAgent = userAgent
	pc.billing = billingSvc

	return nil
}
