package azure

import (
	"context"
	"ez2boot/internal/server"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

// Scrape Azure to retrieve servers.
func (s *Service) Scrape() error {
	s.Logger.Debug("Scraping Azure", "domain", "azure")

	pager := s.VMClient.NewListAllPager(nil)

	servers := []server.Server{}
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			s.Logger.Error("Failed to list Azure VMs", "domain", "azure", "error", err)
			return err
		}

		for _, vm := range page.Value {
			// Filter by tag key
			if _, ok := vm.Tags[s.Config.TagKey]; !ok {
				continue
			}

			resourceGroup, vmName, err := parseVMID(*vm.ID)
			if err != nil {
				s.Logger.Error("Failed to parse VM ID", "id", *vm.ID, "domain", "azure", "error", err)
				continue
			}

			// Fetch instance view for power state
			detail, err := s.VMClient.Get(context.Background(), resourceGroup, vmName, &armcompute.VirtualMachinesClientGetOptions{
				Expand: to.Ptr(armcompute.InstanceViewTypesInstanceView),
			})
			if err != nil {
				s.Logger.Error("Failed to get VM instance view", "name", vmName, "domain", "azure", "error", err)
				continue
			}

			svr := server.Server{
				UniqueID:    *vm.ID,
				Name:        vmName,
				State:       mapState(getPowerState(&detail.VirtualMachine)),
				ServerGroup: *vm.Tags[s.Config.TagKey],
				TimeAdded:   time.Now().Unix(),
			}

			servers = append(servers, svr)
		}
	}

	s.Logger.Debug("Scraped and found number of matching VMs", "domain", "azure", "count", len(servers))
	s.ServerService.UpdateServers(servers)

	return nil
}

// Start required Azure servers
func (s *Service) Start() error {
	s.Logger.Debug("Starting requested VMs", "domain", "azure")

	// Get start VM IDs
	vmIDs, err := s.ServerService.GetPending("off", "on")
	if err != nil {
		s.Logger.Error("Failed to get VM IDs pending on", "domain", "azure", "error", err)
		return err
	}

	// Nothing to do
	if len(vmIDs) == 0 {
		s.Logger.Debug("No VMs to start", "domain", "azure")
		return nil
	}

	// Loop and turn each on
	for _, id := range vmIDs {
		resourceGroup, vmName, err := parseVMID(id)
		if err != nil {
			s.Logger.Error("Failed to parse VM ID", "id", id, "domain", "azure", "error", err)
			continue
		}

		s.Logger.Debug("Starting VM", "name", vmName, "domain", "azure")

		_, err = s.VMClient.BeginStart(context.Background(), resourceGroup, vmName, nil)
		if err != nil {
			s.Logger.Error("Failed to start VM", "name", vmName, "domain", "azure", "error", err)
			continue
		}

		s.Logger.Info("VM start initiated", "name", vmName, "domain", "azure")
	}

	return nil
}

// Stop no longer required Azure servers
func (s *Service) Stop() error {
	s.Logger.Debug("Stopping requested VMs", "domain", "azure")

	vmIDs, err := s.ServerService.GetPending("on", "off")
	if err != nil {
		s.Logger.Error("Failed to get VM IDs pending off", "domain", "azure", "error", err)
		return err
	}

	if len(vmIDs) == 0 {
		s.Logger.Debug("No VMs to stop", "domain", "azure")
		return nil
	}

	for _, id := range vmIDs {
		resourceGroup, vmName, err := parseVMID(id)
		if err != nil {
			s.Logger.Error("Failed to parse VM ID", "id", id, "domain", "azure", "error", err)
			continue
		}

		s.Logger.Debug("Stopping VM", "name", vmName, "domain", "azure")

		_, err = s.VMClient.BeginDeallocate(context.Background(), resourceGroup, vmName, nil)
		if err != nil {
			s.Logger.Error("Failed to stop VM", "name", vmName, "domain", "azure", "error", err)
			continue
		}

		s.Logger.Info("VM stop initiated", "name", vmName, "domain", "azure")
	}

	return nil
}
