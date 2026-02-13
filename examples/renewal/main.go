package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Sage-Infrastructure-Solutions-Group-Inc/LicenseEdictSDK"
)

func main() {
	token := os.Getenv("LICENSE_TOKEN")
	pubKeyB64 := os.Getenv("LICENSE_PUBLIC_KEY")

	if token == "" || pubKeyB64 == "" {
		log.Fatal("Set LICENSE_TOKEN and LICENSE_PUBLIC_KEY environment variables")
	}

	client, err := licenseedict.NewClient(
		licenseedict.WithPublicKey(pubKeyB64),
		licenseedict.WithAppInfo("MyApp", "MyCompany"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Validate the license first
	license, err := client.Validate(token)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Printf("Current license: plan=%s, expires=%v\n", license.Plan, license.ExpiresAt)

	if license.IsExpired() {
		fmt.Println("License is expired, attempting renewal...")
	} else {
		fmt.Println("License is active, attempting early renewal...")
	}

	// Renew the license -- returns *License directly
	renewed, err := client.Renew()
	if err != nil {
		log.Fatalf("Renewal failed: %v", err)
	}

	fmt.Printf("Renewed license valid: %v, expires: %v\n", renewed.Valid, renewed.ExpiresAt)

	// For detailed renewal metadata, use RenewResult() instead:
	//   result, err := client.RenewResult()
	//   fmt.Printf("Renewal status:   %s\n", result.Status)
	//   fmt.Printf("New expires at:   %s\n", result.ExpiresAt)
	//   fmt.Printf("Previous expires: %s\n", result.PreviousExpiresAt)
}
