package main

import (
	"context"
	"github.com/posthog/posthog-go"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-cli/pkg/cmd"
	"github.com/stripe/stripe-cli/pkg/stripe"
)

func main() {
	ctx := context.Background()

	if stripe.TelemetryOptedOut(os.Getenv("STRIPE_CLI_TELEMETRY_OPTOUT")) || stripe.TelemetryOptedOut(os.Getenv("DO_NOT_TRACK")) {
		// Proceed without the telemetry client if client opted out.
		cmd.Execute(ctx)
	} else {
		// Set up the telemetry client and add it to the context
		httpClient := &http.Client{
			Timeout: time.Second * 3,
		}

		key := "phc_Belh475IYfPoF9bke7r9ReO3m7WIa21C5ftRvD10Pvs"
		client, _ := posthog.NewWithConfig(
			key,
			posthog.Config{
				Endpoint: "https://ph.usecyclone.dev",
			},
		)
		defer client.Close()

		sm := stripe.NewSessionManager()

		telemetryClient := &stripe.AnalyticsTelemetryClient{HTTPClient: httpClient, PostHogKey: key, PostHogClient: client, SessionManager: sm}

		contextWithTelemetry := stripe.WithTelemetryClient(ctx, telemetryClient)

		cmd.Execute(contextWithTelemetry)

		// Wait for all telemetry calls to finish before existing the process
		telemetryClient.Wait()
	}
}
