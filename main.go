package main

import (
	"os"

	"github.com/antimatter-studios/cert-manager-webhook-inwx/inwx"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName, inwx.NewSolver())
}
