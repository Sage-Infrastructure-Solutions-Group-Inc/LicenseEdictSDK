package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// Validate the license
	license, err := client.Validate(token)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
	fmt.Printf("License valid: %v (plan: %s)\n", license.Valid, license.Plan)

	// Start background heartbeat -- returns events channel
	events, err := client.StartHeartbeat(licenseedict.HeartbeatOptions{
		InstanceID: "machine-001",
		Hostname:   "dev-workstation",
	})
	if err != nil {
		log.Fatalf("Failed to start heartbeat: %v", err)
	}
	fmt.Println("Heartbeat started. Press Ctrl+C to stop.")

	// Listen for events in a goroutine
	go func() {
		for event := range events {
			switch event.Type {
			case licenseedict.EventHeartbeatOK:
				fmt.Println("  Heartbeat: OK")
			case licenseedict.EventHeartbeatRejected:
				fmt.Println("  Heartbeat: REJECTED (seat limit)")
			case licenseedict.EventHeartbeatError:
				fmt.Printf("  Heartbeat: ERROR (%s)\n", event.Message)
			case licenseedict.EventSeatReleased:
				fmt.Println("  Seat released")
			}
		}
	}()

	// Wait for interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("\nChecking out...")
	if err := client.Checkout(); err != nil {
		log.Printf("Checkout failed: %v", err)
	}
}
