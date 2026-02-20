package azure

import (
	"ez2boot/internal/server"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

// Iterate the VM properties to find power state
func getPowerState(vm *armcompute.VirtualMachine) string {
	if vm.Properties == nil || vm.Properties.InstanceView == nil {
		return "unknown"
	}
	for _, status := range vm.Properties.InstanceView.Statuses {
		if strings.HasPrefix(*status.Code, "PowerState/") {
			return strings.TrimPrefix(*status.Code, "PowerState/")
		}
	}
	return "unknown"
}

// Map provider specific states to generic
func mapState(state string) server.ServerState {
	switch state {
	case "running":
		return server.ServerOn
	case "starting", "stopping", "deallocating":
		return server.ServerTransitioning
	default:
		return server.ServerOff
	}
}

// Azure VM ID is a long string with important values embedded
// /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines/{name}
func parseVMID(id string) (resourceGroup, vmName string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) < 9 {
		return "", "", fmt.Errorf("invalid VM ID: %s", id)
	}
	return parts[4], parts[8], nil
}
