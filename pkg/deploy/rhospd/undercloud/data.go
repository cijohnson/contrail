package undercloud

import (
	"context"

	"github.com/Juniper/contrail/pkg/apisrv/client"
	"github.com/Juniper/contrail/pkg/models"
	"github.com/Juniper/contrail/pkg/services"
)

// Data is the representation of cloud manager details.
type Data struct {
	cloudManagerInfo  *models.RhospdCloudManager
	overcloudNetworks []*OvercloudNetworkData
	client            *client.HTTP
}

// OvercloudNetworkData is the representation of overcloud network details.
type OvercloudNetworkData struct {
	overcloudNetwork *models.RhospdOvercloudNetwork
	virtualNetworks  []*models.VirtualNetwork
	tags             []*models.Tag
	client           *client.HTTP
}

// NewData creates a undercloud data
func NewData(apiClient *client.HTTP) *Data {
	return &Data{
		client: apiClient,
	}
}

// NewOvercloudNetworkData creates a undercloud data
func NewOvercloudNetworkData(overcloudNetwork *models.RhospdOvercloudNetwork,
	apiClient *client.HTTP) *OvercloudNetworkData {
	return &OvercloudNetworkData{
		overcloudNetwork: overcloudNetwork,
		client:           apiClient,
	}
}

func (d *Data) getCloudManagerDetails(undercloudID string) error {
	if err := d.updateUndercloudDetails(undercloudID); err != nil {
		return err
	}
	if err := d.updateOvercloudNetworkDetails(); err != nil {
		return err
	}
	return nil
}

func (d *Data) updateUndercloudDetails(undercloudID string) error {
	request := new(services.GetRhospdCloudManagerRequest)
	request.ID = undercloudID

	resp, err := d.client.GetRhospdCloudManager(context.Background(), request)
	if err != nil {
		return err
	}
	d.cloudManagerInfo = resp.GetRhospdCloudManager()
	if err := d.updateOvercloudChildren(); err != nil {
		return err
	}
	return nil
}

func (d *Data) updateOvercloudChildren() error {
	for i, overcloud := range d.cloudManagerInfo.RhospdOverclouds {
		request := new(services.GetRhospdOvercloudRequest)
		request.ID = overcloud.UUID
		resp, err := d.client.GetRhospdOvercloud(context.Background(), request)
		if err != nil {
			return err
		}
		d.cloudManagerInfo.RhospdOverclouds[i] = resp.GetRhospdOvercloud()
	}
	return nil
}

func (d *Data) updateOvercloudNetworkDetails() error {
	overcloudNetworks := d.cloudManagerInfo.RhospdOverclouds[0].RhospdOvercloudNetworks
	for _, overcloudNetwork := range overcloudNetworks {
		request := new(services.GetRhospdOvercloudNetworkRequest)
		request.ID = overcloudNetwork.UUID

		resp, err := d.client.GetRhospdOvercloudNetwork(context.Background(), request)
		if err != nil {
			return err
		}
		overcloudNetworkData := NewOvercloudNetworkData(resp.GetRhospdOvercloudNetwork(), d.client)
		if err := overcloudNetworkData.update(); err != nil {
			return err
		}
		d.overcloudNetworks = append(d.overcloudNetworks, overcloudNetworkData)
	}
	return nil
}

func (o *OvercloudNetworkData) update() error {
	if err := o.updateTags(); err != nil {
		return err
	}
	if err := o.updateVirtualNetworks(); err != nil {
		return err
	}
	return nil
}

func (o *OvercloudNetworkData) updateTags() error {
	for _, tagRef := range o.overcloudNetwork.TagRefs {
		request := new(services.GetTagRequest)
		request.ID = tagRef.UUID

		resp, err := o.client.GetTag(context.Background(), request)
		if err != nil {
			return err
		}
		o.tags = append(o.tags, resp.GetTag())
	}
	return nil
}

func (o *OvercloudNetworkData) updateVirtualNetworks() error {
	for _, vnRef := range o.overcloudNetwork.VirtualNetworkRefs {
		request := new(services.GetVirtualNetworkRequest)
		request.ID = vnRef.UUID

		resp, err := o.client.GetVirtualNetwork(context.Background(), request)
		if err != nil {
			return err
		}
		o.virtualNetworks = append(o.virtualNetworks, resp.GetVirtualNetwork())
	}
	return nil
}
