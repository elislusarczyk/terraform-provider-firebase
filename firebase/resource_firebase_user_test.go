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

// func TestMakeExportedUser(t *testing.T) {
// 	rur := &identitytoolkit.UserInfo{
// 		LocalId:          "testuser",
// 		Email:            "testuser@example.com",
// 		PhoneNumber:      "+1234567890",
// 		EmailVerified:    true,
// 		DisplayName:      "Test User",
// 		Salt:             "salt",
// 		PhotoUrl:         "http://www.example.com/testuser/photo.png",
// 		PasswordHash:     "passwordhash",
// 		ValidSince:       1494364393,
// 		Disabled:         false,
// 		CreatedAt:        1234567890000,
// 		LastLoginAt:      1233211232000,
// 		CustomAttributes: `{"admin": true, "package": "gold"}`,
// 		ProviderUserInfo: []*identitytoolkit.UserInfoProviderUserInfo{
// 			{
// 				ProviderId:  "password",
// 				DisplayName: "Test User",
// 				PhotoUrl:    "http://www.example.com/testuser/photo.png",
// 				Email:       "testuser@example.com",
// 				RawId:       "testuid",
// 			}, {
// 				ProviderId:  "phone",
// 				PhoneNumber: "+1234567890",
// 				RawId:       "testuid",
// 			}},
// 	}
//
// 	want := &auth.ExportedUserRecord{
// 		UserRecord:   testUser,
// 		PasswordHash: "passwordhash",
// 		PasswordSalt: "salt",
// 	}
// 	exported, err := makeExportedUser(rur)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !reflect.DeepEqual(exported.UserRecord, want.UserRecord) {
// 		// zero in
// 		t.Errorf("makeExportedUser() = %#v; want: %#v \n(%#v)\n(%#v)", exported.UserRecord, want.UserRecord,
// 			exported.UserMetadata, want.UserMetadata)
// 	}
// 	if exported.PasswordHash != want.PasswordHash {
// 		t.Errorf("PasswordHash = %q; want = %q", exported.PasswordHash, want.PasswordHash)
// 	}
// 	if exported.PasswordSalt != want.PasswordSalt {
// 		t.Errorf("PasswordSalt = %q; want = %q", exported.PasswordSalt, want.PasswordSalt)
// 	}
// }

// func TestCreateUser(t *testing.T) {
// 	resp := `{
// 		"kind": "identitytoolkit#SignupNewUserResponse",
// 		"localId": "expectedUserID"
// 	}`
// 	c := Client{StorageClient}
// 	defer s.Close()
//
// 	cases := []struct {
// 		params *auth.UserToCreate
// 		req    map[string]interface{}
// 	}{
// 		{
// 			nil,
// 			map[string]interface{}{},
// 		},
// 		{
// 			&auth.UserToCreate{},
// 			map[string]interface{}{},
// 		},
// 		{
// 			(&auth.UserToCreate{}).Password("123456"),
// 			map[string]interface{}{"password": "123456"},
// 		},
// 		{
// 			(&auth.UserToCreate{}).UID("1"),
// 			map[string]interface{}{"localId": "1"},
// 		},
// 		{
// 			(&auth.UserToCreate{}).UID(strings.Repeat("a", 128)),
// 			map[string]interface{}{"localId": strings.Repeat("a", 128)},
// 		},
// 		{
// 			(&auth.UserToCreate{}).PhoneNumber("+1"),
// 			map[string]interface{}{"phoneNumber": "+1"},
// 		},
// 		{
// 			(&auth.UserToCreate{}).DisplayName("a"),
// 			map[string]interface{}{"displayName": "a"},
// 		},
// 		{
// 			(&auth.UserToCreate{}).Email("a@a"),
// 			map[string]interface{}{"email": "a@a"},
// 		},
// 		{
// 			(&auth.UserToCreate{}).Disabled(true),
// 			map[string]interface{}{"disabled": true},
// 		},
// 		{
// 			(&auth.UserToCreate{}).Disabled(false),
// 			map[string]interface{}{"disabled": false},
// 		},
// 		{
// 			(&auth.UserToCreate{}).EmailVerified(true),
// 			map[string]interface{}{"emailVerified": true},
// 		},
// 		{
// 			(&auth.UserToCreate{}).EmailVerified(false),
// 			map[string]interface{}{"emailVerified": false},
// 		},
// 		{
// 			(&auth.UserToCreate{}).PhotoURL("http://some.url"),
// 			map[string]interface{}{"photoUrl": "http://some.url"},
// 		},
// 	}
// 	for _, tc := range cases {
// 		uid, err := s.Client.createUser(context.Background(), tc.params)
// 		if uid != "expectedUserID" || err != nil {
// 			t.Errorf("createUser(%v) = (%q, %v); want = (%q, nil)", tc.params, uid, err, "expectedUserID")
// 		}
// 		want, err := json.Marshal(tc.req)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		if !reflect.DeepEqual(s.Rbody, want) {
// 			t.Errorf("createUser() request = %v; want = %v", string(s.Rbody), string(want))
// 		}
// 	}
// }
