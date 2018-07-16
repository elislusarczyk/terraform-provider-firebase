# Basic Firebase Authentication

After you run `terraform apply` on this configuration, it will
create a user in your firebase user management console.

To run, configure your Firebase provider as described in

https://www.terraform.io/docs/providers/firebase/index.html

Run with a command like this:

```
terraform apply -var 'service_account_key={your_service_account_key}' \
```

For example:

```
terraform apply -var 'service_account_key=$HOME/test-project-112356-firebase-adminsdk-to28a-b223ad5dx1.json'
```

