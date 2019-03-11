package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

// Provider returns the actual provider instance.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credentials": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CREDENTIALS",
					"GOOGLE_CLOUD_KEYFILE_JSON",
					"GCLOUD_KEYFILE_JSON",
				}, nil),
				ValidateFunc: validateCredentials,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"greseller_cloud_billing_account": resourceBillingAccount(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	credentials := d.Get("credentials").(string)

	pc := ProviderClient{}

	if err := pc.init(credentials); err != nil {
		return nil, errors.Wrap(err, "failed to create provider client")
	}

	return &pc, nil
}

func validateCredentials(v interface{}, k string) (warns []string, errs []error) {
	if v == nil || v.(string) == "" {
		return
	}
	credentialsFile := v.(string)

	if _, err := apiClientFromCredentialsFile(credentialsFile); err != nil {
		errs = append(errs, errors.Wrap(err, "Invalid credentials"))
	}

	return
}
