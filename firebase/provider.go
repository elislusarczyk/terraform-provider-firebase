package firebase

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"service_account_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("FIREBASE_SERVICE_ACCOUNT_KEY", nil),
				Description: descriptions["service_account_key"],
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"firebase_user": resourceFirebaseUser(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"service_account_key": "Firebase Admin SDK Service Account Key File",
		"firebase_user":       "Firebase User",
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		ServiceAccountKey: d.Get("service_account_key").(string),
	}
	return config.Client()
}
