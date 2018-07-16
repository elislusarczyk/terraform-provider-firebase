package firebase

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/terraform"
)

func resourceFirebaseUserMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found Firebase User State v0; migrating to v1")
		return is, nil
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}
