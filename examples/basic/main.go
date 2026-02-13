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

	// Simple function API (new simplified entry point)
	license, err := licenseedict.CheckLicense(pubKeyB64, token)
	if err != nil {
		log.Fatalf("License check failed: %v", err)
	}

	fmt.Printf("Valid:      %v\n", license.Valid)
	fmt.Printf("License ID: %s\n", license.LicenseID)
	fmt.Printf("Product:    %s\n", license.ProductID)
	fmt.Printf("Licensee:   %s\n", license.Licensee)
	fmt.Printf("Plan:       %s\n", license.Plan)
	fmt.Printf("Features:   %v\n", license.Features)
	fmt.Printf("Expires:    %v\n", license.ExpiresAt)
	fmt.Printf("Server URL: %s\n", license.ServerURL)

	hasPro, err := licenseedict.CheckFeature(pubKeyB64, token, "PRO")
	if err != nil {
		log.Printf("Feature check failed: %v", err)
	} else if hasPro {
		fmt.Println("PRO feature is enabled!")
	}
}
