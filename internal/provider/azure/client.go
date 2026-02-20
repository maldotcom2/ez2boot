package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

func getVMClient(cred *azidentity.DefaultAzureCredential, subID string) (*armcompute.VirtualMachinesClient, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(subID, cred, nil)
	if err != nil {
		return nil, err
	}

	return vmClient, nil
}
