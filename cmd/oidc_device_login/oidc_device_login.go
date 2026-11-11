package oidc_device_login

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var FlagIssuer string
var FlagClientID string
var FlagClientSecret string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagIssuer, "issuer", "", "Issuer")
	Cmd.MarkFlagRequired("issuer")
	Cmd.Flags().StringVar(&FlagClientID, "client-id", "", "Client ID")
	Cmd.MarkFlagRequired("client-id")
	Cmd.Flags().StringVar(&FlagClientSecret, "client-secret", "", "Client Secret")
	Cmd.MarkFlagRequired("client-secret")
}

var Cmd = &cobra.Command{
	Use:   "oidc-device-login",
	Short: "OIDC Device Authorization Grant login, prints ID Token to stdout",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		oidcDeviceLogin(FlagIssuer, FlagClientID, FlagClientSecret)
	},
}

func oidcDeviceLogin(issuer, clientID, clientSecret string) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatal(err)
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		// Endpoint contains DeviceAuthURL from the discovery document
		// (device_authorization_endpoint)
		Endpoint: provider.Endpoint(),
		Scopes:   []string{oidc.ScopeOpenID},
	}

	// oauth2 sends only client_id here, confidential client needs client_secret too
	deviceAuth, err := config.DeviceAuth(ctx, oauth2.SetAuthURLParam("client_secret", clientSecret))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, "Open in browser:", deviceAuth.VerificationURIComplete)
	fmt.Fprintln(os.Stderr, "Or open:        ", deviceAuth.VerificationURI)
	fmt.Fprintln(os.Stderr, "and enter code: ", deviceAuth.UserCode)
	fmt.Fprintln(os.Stderr)

	// Polls the token endpoint until the user approves (or denies)
	// the login or the code expires
	oauth2Token, err := config.DeviceAccessToken(ctx, deviceAuth)
	if err != nil {
		log.Fatal(err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Fatal("No id_token field in oauth2 token.")
	}

	_, err = provider.Verifier(&oidc.Config{ClientID: clientID}).Verify(ctx, rawIDToken)
	if err != nil {
		log.Fatal("Failed to verify ID Token: " + err.Error())
	}

	// Keycloak Access Token has "aud": "account" unless an audience mapper
	// for this client is configured, so skip the client ID (audience) check
	_, err = provider.Verifier(&oidc.Config{SkipClientIDCheck: true}).Verify(ctx, oauth2Token.AccessToken)
	if err != nil {
		log.Fatal("Failed to verify Access Token: " + err.Error())
	}

	fmt.Println(rawIDToken)
}
