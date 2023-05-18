package main

import (
	"bufio"
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
		// Hijack stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Hijack stderr
		oldStderr := os.Stderr
		rErr, wErr, _ := os.Pipe()
		os.Stderr = wErr

		// Restore stdout and stderr when the program exits
		defer func() {
			os.Stdout = oldStdout
			os.Stderr = oldStderr
		}()

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

		go captureFD(r, oldStdout, "stdout", telemetryClient)
		go captureFD(rErr, oldStderr, "stderr", telemetryClient)

		cmd.Execute(contextWithTelemetry)

		// Wait for all telemetry calls to finish before existing the process
		telemetryClient.Wait()

		// Ensure all data is flushed before the program exits
		w.Close()
		wErr.Close()

		// Restore stdout, stderr
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}
}

func captureFD(r *os.File, oldW *os.File, source string, a *stripe.AnalyticsTelemetryClient) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		a.SendCli(s, source)
		oldW.WriteString(s + "\n")
	}
}
