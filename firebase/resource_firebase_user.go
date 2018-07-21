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
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
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
	log.Printf("[INFO] Creating user uid: %s", d.Get("uid").(string))

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
	log.Printf("[INFO] UID: %s", userRecord.UserInfo.UID)

	// Store the resulting UID so we can look this up later
	d.SetId(userRecord.UserInfo.UID)

	log.Printf("[DEBUG] Waiting for user (%s) to become created", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleted"},
		Target:     []string{"created"},
		Refresh:    userStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for user (%s) state to be created: %s", d.Id(), err)
	}

	return resourceFirebaseUserUpdate(d, meta)
}

func resourceFirebaseUserRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Reading user uid: %s", d.Id())
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
	log.Printf("[DEBUG] update resource %+v\n", d)

	if !d.IsNewResource() {
		d.Partial(true)
	}

	changed := false
	if d.HasChange("uid") && !d.IsNewResource() {
		changed = true
		d.SetPartial("uid")
	}
	if d.HasChange("email") && !d.IsNewResource() {
		changed = true
		d.SetPartial("email")
	}
	if d.HasChange("display_name") && !d.IsNewResource() {
		changed = true
		d.SetPartial("display_name")
	}
	if d.HasChange("email_verified") && !d.IsNewResource() {
		changed = true
		d.SetPartial("email_verified")
	}
	if d.HasChange("phone_number") && !d.IsNewResource() {
		changed = true
		d.SetPartial("phone_number")
	}
	if d.HasChange("password") && !d.IsNewResource() {
		changed = true
		d.SetPartial("password")
	}
	if d.HasChange("photo_url") && !d.IsNewResource() {
		changed = true
		d.SetPartial("photo_url")
	}

	if changed {
		log.Printf("[INFO] Updating uid: %s", d.Id())
		client := meta.(Client).Auth
		var u auth.UserToUpdate

		u.Email(d.Get("email").(string))
		u.DisplayName(d.Get("display_name").(string))
		u.EmailVerified(d.Get("email_verified").(bool))
		u.PhoneNumber(d.Get("phone_number").(string))
		u.Password(d.Get("password").(string))
		u.PhotoURL(d.Get("photo_url").(string))

		_, err := client.UpdateUser(context.Background(), d.Id(), &u)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceFirebaseUserDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Deleting uid: %s", d.Id())

	client := meta.(Client).Auth

	err := client.DeleteUser(context.Background(), d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for user (%s) to become deleted", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"created"},
		Target:     []string{"deleted"},
		Refresh:    userStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for user (%s) state to be deleted: %s", d.Id(), err)
	}

	return nil
}

func userStateRefreshFunc(client *auth.Client, uid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Checking user (%s) state\n", uid)
		userRecord, err := client.GetUser(context.Background(), uid)
		log.Printf("[DEBUG] UserRecord: %+v)\n", userRecord)
		if err != nil {
			log.Printf("[DEBUG] The user (%s) doesn't exist state (deleted)\n", uid)
			return auth.UserInfo{}, "deleted", nil
		}
		log.Printf("[DEBUG] The user (%s) exists state (created)\n", uid)
		return userRecord, "created", nil
	}
}
