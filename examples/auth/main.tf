variable "service_account_key" {}

# Specify the provider and access details
provider "firebase" {
  service_account_key = "${var.service_account_key}"
}

resource "firebase_user" "john_doe" {
  uid            = "2d5ae085-679b-4a92-89e7-97cced6d4c79"
  display_name   = "John Doe"
  disabled       = false
  email          = "john.doe@example.com"
  email_verified = true
  password       = "password123"
  phone_number   = "+14155552671"
  photo_url      = "http://www.example.com/2d5ae085-679b-4a92-89e7-97cced6d4c79/photo.png"
}
