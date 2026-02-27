package inwx

import (
	"fmt"
	"time"

	goinwx "github.com/nrdcg/goinwx"
)

const zonesCacheTTL = 5 * time.Minute

type ClientWrapper struct {
	client         *goinwx.Client
	zonesCache     []string
	zonesCacheTime time.Time
}

type AbstractClientWrapper interface {
	login() (*goinwx.LoginResponse, error)
	logout() error
	getRecords(domain string) (*[]goinwx.NameserverRecord, error)
	getZones() (*[]string, error)
	createRecord(request *goinwx.NameserverRecordRequest) error
	deleteRecord(recID string) error
}

func NewClientWrapper(username, password string, sandbox bool) *ClientWrapper {
	return &ClientWrapper{
		client: goinwx.NewClient(username, password, &goinwx.ClientOptions{Sandbox: sandbox}),
	}
}

func (w *ClientWrapper) login() (*goinwx.LoginResponse, error) {
	return w.client.Account.Login()
}

func (w *ClientWrapper) logout() error {
	return w.client.Account.Logout()
}

func (w *ClientWrapper) getRecords(domain string) (*[]goinwx.NameserverRecord, error) {
	zone, err := w.client.Nameservers.Info(&goinwx.NameserverInfoRequest{Domain: domain})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve records for zone %s: %w", domain, err)
	}
	return &zone.Records, nil
}

func (w *ClientWrapper) getZones() (*[]string, error) {
	if w.zonesCache != nil && time.Since(w.zonesCacheTime) < zonesCacheTTL {
		zones := w.zonesCache
		return &zones, nil
	}

	zones := []string{}
	page := 1
	for {
		response, err := w.client.Nameservers.ListWithParams(&goinwx.NameserverListRequest{
			Domain:    "*",
			Page:      page,
			PageLimit: 100,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list nameserver zones (page %d): %w", page, err)
		}
		for _, domain := range response.Domains {
			zones = append(zones, domain.Domain)
		}
		if len(response.Domains) == 0 || len(zones) >= response.Count {
			break
		}
		page++
	}

	w.zonesCache = zones
	w.zonesCacheTime = time.Now()

	return &zones, nil
}

func (w *ClientWrapper) createRecord(request *goinwx.NameserverRecordRequest) error {
	_, err := w.client.Nameservers.CreateRecord(request)
	return err
}

func (w *ClientWrapper) deleteRecord(recID string) error {
	return w.client.Nameservers.DeleteRecord(recID)
}
