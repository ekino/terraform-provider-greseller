package provider

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	billing "google.golang.org/api/cloudbilling/v1"
)

func resourceBillingAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceBillingAccountCreate,
		Read:   resourceBillingAccountRead,
		Update: resourceBillingAccountUpdate,
		Delete: resourceBillingAccountDelete,
		Importer: &schema.ResourceImporter{
			State: resourceBillingAccountImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"open": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"master_billing_account": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBillingAccountCreate(d *schema.ResourceData, meta interface{}) error {
	providerClient := meta.(*ProviderClient)

	billingAccount := &billing.BillingAccount{
		DisplayName:          d.Get("display_name").(string),
		MasterBillingAccount: d.Get("master_billing_account").(string),
		Open:                 d.Get("open").(bool),
	}

	var createdBillingAccount *billing.BillingAccount
	var err error
	err = retry(func() error {
		createdBillingAccount, err = providerClient.billing.BillingAccounts.Create(billingAccount).Do()
		return err
	})

	if err != nil {
		return fmt.Errorf("Error creating billing account: %s", err)
	}

	d.SetId(createdBillingAccount.Name)
	log.Printf("[INFO] Created billing account: %s", createdBillingAccount.DisplayName)
	return resourceBillingAccountRead(d, meta)
}

func resourceBillingAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	providerClient := meta.(*ProviderClient)

	billingAccount := &billing.BillingAccount{}
	shouldUpdate := false

	if d.HasChange("display_name") {
		log.Printf("[DEBUG] Updating billing account display name: %s", d.Get("display_name").(string))
		billingAccount.DisplayName = d.Get("display_name").(string)
		shouldUpdate = true
	}

	if d.HasChange("open") {
		log.Printf("[WARN] Can not close or open a billing account. This is not supported by the api currently. This change will be ignored")
		//billingAccount.Open = d.Get("open").(bool)
	}

	if shouldUpdate {
		var updatedBillingAccount *billing.BillingAccount
		var err error
		err = retry(func() error {
			updatedBillingAccount, err = providerClient.billing.BillingAccounts.Patch(d.Id(), billingAccount).UpdateMask("display_name").Do()
			return err
		})

		if err != nil {
			return fmt.Errorf("Error updating billing account: %s", err)
		}

		log.Printf("[INFO] Updated billing account: %s", updatedBillingAccount.DisplayName)
	}

	return resourceBillingAccountRead(d, meta)
}

func resourceBillingAccountRead(d *schema.ResourceData, meta interface{}) error {
	providerClient := meta.(*ProviderClient)

	var billingAccount *billing.BillingAccount
	var err error
	err = retry(func() error {
		billingAccount, err = providerClient.billing.BillingAccounts.Get(d.Id()).Do()
		return err
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Billing account %q", d.Get("display_name").(string)))
	}

	d.SetId(billingAccount.Name)
	d.Set("display_name", billingAccount.DisplayName)
	d.Set("open", billingAccount.Open)
	d.Set("master_billing_account", billingAccount.MasterBillingAccount)

	return nil
}

func resourceBillingAccountDelete(d *schema.ResourceData, meta interface{}) error {
	/*providerClient := meta.(*ProviderClient)

	billingAccount := &billing.BillingAccount{
		Open: false,
	}

	if d.HasChange("display_name") {
		log.Printf("[DEBUG] Updating billing account display name: %s", d.Get("display_name").(string))
		billingAccount.DisplayName = "[Closed] " + d.Get("display_name").(string)
	}

	var updatedBillingAccount *billing.BillingAccount
	var err error
	err = retry(func() error {
		updatedBillingAccount, err = providerClient.billing.BillingAccounts.Patch(d.Id(), billingAccount).Do()
		return err
	})

	if err != nil {
		return fmt.Errorf("Error updating billing account: %s", err)
	}*/

	log.Printf("[WARN] Can not close or open a billing account. This is not supported by the api currently."+
		"The billing account %s will be removed from tf state, but won't be deleted nor closed", d.Id())

	d.SetId("")
	return nil
}

// Allow importing using name
func resourceBillingAccountImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	providerClient := meta.(*ProviderClient)

	billingAccount, err := providerClient.billing.BillingAccounts.Get(d.Id()).Do()

	if err != nil {
		return nil, fmt.Errorf("Error fetching billing account. Make sure the billing account exists: %s ", err)
	}

	d.SetId(billingAccount.Name)
	d.Set("display_name", billingAccount.DisplayName)
	d.Set("open", billingAccount.Open)
	d.Set("master_billing_account", billingAccount.MasterBillingAccount)

	return []*schema.ResourceData{d}, nil
}
