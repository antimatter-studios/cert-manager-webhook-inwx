package inwx

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	goinwx "github.com/nrdcg/goinwx"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"k8s.io/client-go/rest"
)

const defaultTTL = 60

type INWXDNSSolver struct {
	client AbstractClientWrapper
	logger *slog.Logger
}

func NewSolver() *INWXDNSSolver {
	return &INWXDNSSolver{
		logger: slog.Default(),
	}
}

func NewSolverWithClient(client AbstractClientWrapper) *INWXDNSSolver {
	return &INWXDNSSolver{
		client: client,
		logger: slog.Default(),
	}
}

func (s *INWXDNSSolver) Name() string {
	return "inwx"
}

func (s *INWXDNSSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	username := os.Getenv("INWX_USERNAME")
	password := os.Getenv("INWX_PASSWORD")
	sandboxStr := os.Getenv("INWX_SANDBOX")

	if username == "" || password == "" {
		return fmt.Errorf("INWX_USERNAME and INWX_PASSWORD environment variables must be set")
	}

	sandbox, _ := strconv.ParseBool(sandboxStr)

	s.client = NewClientWrapper(username, password, sandbox)

	if _, err := s.client.login(); err != nil {
		s.logger.Error("startup: failed to login to INWX", "err", err)
	} else {
		if zones, err := s.client.getZones(); err != nil {
			s.logger.Error("startup: failed to list zones", "err", err)
		} else {
			s.logger.Info("INWX zones available", "count", len(*zones), "zones", strings.Join(*zones, ", "))
		}
		if err := s.client.logout(); err != nil {
			s.logger.Error("startup: failed to logout", "err", err)
		}
	}

	return nil
}

// Present creates a TXT record for the ACME DNS-01 challenge.
// cert-manager provides ResolvedFQDN (e.g. "_acme-challenge.example.com.") and
// ResolvedZone (e.g. "example.com.") with trailing dots, and Key as the challenge token.
func (s *INWXDNSSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	zone := strings.TrimSuffix(ch.ResolvedZone, ".")
	fqdn := strings.TrimSuffix(ch.ResolvedFQDN, ".")

	s.logger.Info("presenting ACME challenge", "fqdn", fqdn, "zone", zone)

	if _, err := s.client.login(); err != nil {
		return fmt.Errorf("failed to login to INWX: %w", err)
	}
	defer func() {
		if err := s.client.logout(); err != nil {
			s.logger.Error("failed to logout", "err", err)
		}
	}()

	recordName := extractRecordName(fqdn, zone)

	err := s.client.createRecord(&goinwx.NameserverRecordRequest{
		Domain:  zone,
		Name:    recordName,
		Type:    "TXT",
		Content: ch.Key,
		TTL:     defaultTTL,
	})
	if err != nil {
		if isObjectExistsError(err) {
			s.logger.Info("TXT record already exists, skipping", "fqdn", fqdn)
			return nil
		}
		return fmt.Errorf("failed to create TXT record for %s: %w", fqdn, err)
	}

	s.logger.Info("created TXT record", "fqdn", fqdn, "zone", zone)
	return nil
}

// CleanUp deletes the TXT record that was created for the ACME DNS-01 challenge.
// Only the record matching both the name and the specific challenge key is deleted,
// to support concurrent validations for the same domain.
func (s *INWXDNSSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	zone := strings.TrimSuffix(ch.ResolvedZone, ".")
	fqdn := strings.TrimSuffix(ch.ResolvedFQDN, ".")

	s.logger.Info("cleaning up ACME challenge", "fqdn", fqdn, "zone", zone)

	if _, err := s.client.login(); err != nil {
		return fmt.Errorf("failed to login to INWX: %w", err)
	}
	defer func() {
		if err := s.client.logout(); err != nil {
			s.logger.Error("failed to logout", "err", err)
		}
	}()

	records, err := s.client.getRecords(zone)
	if err != nil {
		return fmt.Errorf("failed to get records for zone %s: %w", zone, err)
	}

	recordName := extractRecordName(fqdn, zone)

	for _, rec := range *records {
		if rec.Type == "TXT" && rec.Name == recordName && rec.Content == ch.Key {
			if err := s.client.deleteRecord(rec.ID); err != nil {
				return fmt.Errorf("failed to delete TXT record %s: %w", rec.ID, err)
			}
			s.logger.Info("deleted TXT record", "fqdn", fqdn, "zone", zone, "id", rec.ID)
			return nil
		}
	}

	s.logger.Info("no matching TXT record found to delete", "fqdn", fqdn, "zone", zone)
	return nil
}

// extractRecordName computes the INWX record name from a full DNS name and zone.
// For example: extractRecordName("_acme-challenge.sub.example.com", "example.com") returns "_acme-challenge.sub"
func extractRecordName(dnsName string, zone string) string {
	if dnsName == zone {
		return ""
	}
	return strings.TrimSuffix(dnsName, "."+zone)
}

func isObjectExistsError(err error) bool {
	var apiErr *goinwx.ErrorResponse
	if errors.As(err, &apiErr) {
		return apiErr.Code == 2302
	}
	return false
}
