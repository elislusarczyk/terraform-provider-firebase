package firebase

import (
	"context"
	"fmt"
	"log"
	"time"

	"firebase.google.com/go/auth"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceFirebaseUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirebaseUserCreate,
		Read:   resourceFirebaseUserRead,
		Update: resourceFirebaseUserUpdate,
		Delete: resourceFirebaseUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 0,
		MigrateState:  resourceFirebaseUserMigrateState,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"uid": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"email": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateEmail,
			},
			"email_verified": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(6, 128),
			},
			"phone_number": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateE164PhoneNumber,
			},
			"photo_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateURL,
			},
		},
	}
}

func resourceFirebaseUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(Client).Auth

	var u auth.UserToCreate

	u.UID(d.Get("uid").(string))
	u.Email(d.Get("email").(string))
	u.DisplayName(d.Get("display_name").(string))
	u.EmailVerified(d.Get("email_verified").(bool))
	u.PhoneNumber(d.Get("phone_number").(string))
	u.Password(d.Get("password").(string))
	u.PhotoURL(d.Get("photo_url").(string))

	userRecord, err := client.CreateUser(context.Background(), &u)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] userRecord: %+v", userRecord)

	log.Printf("[INFO] UID: %s", userRecord.UserInfo.UID)

	// Store the resulting UID so we can look this up later
	d.SetId(userRecord.UserInfo.UID)

	log.Printf("[DEBUG] Waiting for user (%s) to become created", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleted"},
		Target:     []string{"created"},
		Refresh:    userStateRefreshFunc(client, d.Id(), []string{}),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for user (%s) to be created: %s", d.Id(), err)
	}

	return resourceFirebaseUserUpdate(d, meta)
}

func resourceFirebaseUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(Client).Auth

	userRecord, err := client.GetUser(context.Background(), d.Id())
	if err != nil {
		return err
	}

	d.Set("uid", userRecord.UserInfo.UID)
	d.Set("display_name", userRecord.UserInfo.DisplayName)
	d.Set("disabled", userRecord.Disabled)
	d.Set("email", userRecord.UserInfo.Email)
	d.Set("email_verified", userRecord.EmailVerified)
	d.Set("phone_number", userRecord.UserInfo.PhoneNumber)
	d.Set("photo_url", userRecord.UserInfo.PhotoURL)

	return nil
}

func resourceFirebaseUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(Client).Auth

	d.Partial(true)

	changed := false
	if d.HasChange("uid") && !d.IsNewResource() {
		changed = true
	}
	if d.HasChange("email") && !d.IsNewResource() {
		changed = true
	}
	if d.HasChange("display_name") && !d.IsNewResource() {
		changed = true
	}
	if d.HasChange("email_verified") && !d.IsNewResource() {
		changed = true
	}
	if d.HasChange("phone_number") && !d.IsNewResource() {
		changed = true
	}
	if d.HasChange("password") && !d.IsNewResource() {
		changed = true
	}
	if d.HasChange("photo_url") && !d.IsNewResource() {
		changed = true
	}

	if changed {
		var u auth.UserToUpdate

		log.Printf("[INFO] Updating uid: %s", d.Id())

		u.Email(d.Get("email").(string))
		u.DisplayName(d.Get("display_name").(string))
		u.EmailVerified(d.Get("email_verified").(bool))
		u.PhoneNumber(d.Get("phone_number").(string))
		u.Password(d.Get("password").(string))
		u.PhotoURL(d.Get("photo_url").(string))

		userRecord, err := client.UpdateUser(context.Background(), d.Id(), &u)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] userRecord: %+v", userRecord)
	}
	return nil
}

func resourceFirebaseUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(Client).Auth
	log.Printf("[INFO] Deleting uid: %s", d.Id())
	err := client.DeleteUser(context.Background(), d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for user (%s) to become deleted", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"created"},
		Target:     []string{"deleted"},
		Refresh:    userStateRefreshFunc(client, d.Id(), []string{}),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for user (%s) to be created: %s", d.Id(), err)
	}

	return nil
}

func userStateRefreshFunc(client *auth.Client, uid string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Checking user (%s) state", uid)
		userRecord, err := client.GetUser(context.Background(), uid)
		if err != nil {
			log.Printf("[DEBUG] The user (%s) doesn't exist state (deleted)", uid)
			return nil, "deleted", nil
		}
		log.Printf("[DEBUG] The user (%s) exists state (created)", uid)
		return userRecord, "created", nil
	}
}
