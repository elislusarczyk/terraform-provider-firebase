package firebase

import (
	// "context"
	// "encoding/json"
	// "reflect"
	// "strings"
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	// "google.golang.org/api/identitytoolkit/v3"

	"firebase.google.com/go/auth"
)

const defaultProviderID = "firebase"

var testUser = &auth.UserRecord{
	UserInfo: &auth.UserInfo{
		UID:         "2d5ae085-679b-4a92-89e7-97cced6d4c79",
		Email:       "john.doe@example.com",
		PhoneNumber: "+14155552671",
		DisplayName: "John Doe",
		PhotoURL:    "http://www.example.com/2d5ae085-679b-4a92-89e7-97cced6d4c79/photo.png",
		ProviderID:  defaultProviderID,
	},
	Disabled: false,

	EmailVerified: true,
	ProviderUserInfo: []*auth.UserInfo{
		{
			ProviderID:  "password",
			DisplayName: "John Doe",
			PhotoURL:    "http://www.example.com/2d5ae085-679b-4a92-89e7-97cced6d4c79/photo.png",
			Email:       "john.doe@example.com",
			UID:         "2d5ae085-679b-4a92-89e7-97cced6d4c79",
		}, {
			ProviderID:  "phone",
			PhoneNumber: "+14155552671",
			UID:         "2d5ae085-679b-4a92-89e7-97cced6d4c79",
		},
	},
	TokensValidAfterMillis: 1494364393000,
	UserMetadata: &auth.UserMetadata{
		CreationTimestamp:  1234567890000,
		LastLogInTimestamp: 1233211232000,
	},
	CustomClaims: map[string]interface{}{"admin": true, "package": "gold"},
}

func TestAccFirebaseUser(t *testing.T) {
	var v auth.UserRecord

	rInt := acctest.RandInt()

	testCheck := func(rInt int) func(*terraform.State) error {
		return func(*terraform.State) error {
			if v.UserInfo.UID != testUser.UserInfo.UID {
				return fmt.Errorf("incorrect UID: %#v", v.UserInfo.UID)
			}
			if v.UserInfo.Email != testUser.UserInfo.Email {
				return fmt.Errorf("incorrect Email: %#v", v.UserInfo.Email)
			}
			if v.UserInfo.PhoneNumber != testUser.UserInfo.PhoneNumber {
				return fmt.Errorf("incorrect PhoneNumber: %#v", v.UserInfo.PhoneNumber)
			}
			if v.UserInfo.DisplayName != testUser.UserInfo.DisplayName {
				return fmt.Errorf("incorrect DisplayName: %#v", v.UserInfo.DisplayName)
			}
			if v.UserInfo.PhotoURL != testUser.UserInfo.PhotoURL {
				return fmt.Errorf("incorrect PhotoURL: %#v", v.UserInfo.PhotoURL)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "firebase_user.john_doe",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(testUser),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("firebase_user.john_doe", &v),
					testCheck(rInt),
					resource.TestCheckResourceAttr(
						"firebase_user.john_doe",
						"uid",
						"2d5ae085-679b-4a92-89e7-97cced6d4c79"),
					resource.TestCheckResourceAttr(
						"firebase_user.john_doe",
						"email",
						"john.doe@example.com"),
				),
			},
			{
				Config: testAccUserConfig(testUser),
				Check: func(*terraform.State) error {
					auth := testAccProvider.Meta().(Client).Auth
					return auth.DeleteUser(context.Background(), testUser.UID)
				},
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	return testAccCheckUserDestroyWithProvider(s, testAccProvider)
}

func testAccCheckUserDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	client := provider.Meta().(Client).Auth
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "firebase_user" {
			continue
		}
		userRecord, err := client.GetUser(context.Background(), testUser.UserInfo.UID)
		if err == nil {
			return fmt.Errorf("Found existing user: %s", userRecord.UserInfo.UID)
		}
		return err
	}
	return nil
}

func testAccCheckUserExists(n string, u *auth.UserRecord) resource.TestCheckFunc {
	return testAccCheckUserExistsWithProvider(n, u, func() *schema.Provider { return testAccProvider })
}

func testAccCheckUserExistsWithProvider(n string, u *auth.UserRecord, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		provider := providerF()

		client := provider.Meta().(Client).Auth
		_, err := client.GetUser(context.Background(), testUser.UserInfo.UID)
		if err != nil {
			return err
		}

		return fmt.Errorf("User not found")
	}
}

func testAccUserConfig(u *auth.UserRecord) string {
	return fmt.Sprintf(`
resource "firebase_user" "john_doe" {
	uid            = "%s"
	display_name   = "%s"
	disabled       = false
	email          = "%s"
	email_verified = true
	password       = "password123"
	phone_number   = "%s"
	photo_url      = "%s"
}
`, u.UserInfo.UID,
		u.UserInfo.DisplayName,
		u.UserInfo.Email,
		u.UserInfo.PhoneNumber,
		u.UserInfo.PhotoURL)
}
