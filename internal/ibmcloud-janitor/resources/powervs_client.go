/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/boskos/common"
	"sigs.k8s.io/boskos/common/ibmcloud"
	"sigs.k8s.io/boskos/internal/ibmcloud-janitor/account"
)

type IBMPowerVSClient struct {
	session *ibmpisession.IBMPISession

	instance *PowervsInstance
	network  *PowervsNetwork
	key      *APIKey
	resource *common.Resource
}

// Returns the virtual server instances in the PowerVS service instance
func (p *IBMPowerVSClient) GetInstances() (*models.PVMInstances, error) {
	return p.instance.instanceClient.GetAll()
}

// Deletes the virtual server instances in the PowerVS service instance
func (p *IBMPowerVSClient) DeleteInstance(id string) error {
	return p.instance.instanceClient.Delete(id)
}

// Returns the networks in the PowerVS service instance
func (p *IBMPowerVSClient) GetNetworks() (*models.Networks, error) {
	return p.network.networkClient.GetAll()
}

// Deletes the network in PowerVS service instance
func (p *IBMPowerVSClient) DeleteNetwork(id string) error {
	return p.network.networkClient.Delete(id)
}

// Returns ports of the network instance
func (p *IBMPowerVSClient) GetPorts(id string) (*models.NetworkPorts, error) {
	return p.network.networkClient.GetAllPorts(id)
}

// Deletes the port of the network
func (p *IBMPowerVSClient) DeletePort(networkID, portID string) error {
	return p.network.networkClient.DeletePort(networkID, portID)
}

// Returns a new PowerVS client
func NewPowerVSClient(options *CleanupOptions) (*IBMPowerVSClient, error) {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	pclient := &IBMPowerVSClient{}
	powervsData, err := ibmcloud.GetResourceData(options.Resource)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the resource data")
	}

	pclient.key = &APIKey{
		serviceIDName: options.Resource.Name,
		value:         &powervsData.APIKey,
	}

	auth, err := account.GetAuthenticator()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the authenticator")
	}

	sclient, err := NewServiceIDClient(auth, pclient.key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create serviceID client")
	}

	account, err := sclient.GetAccount()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the account")
	}

	clientOptions := &ibmpisession.IBMPIOptions{
		Debug:         options.Debug,
		Authenticator: auth,
		Region:        powervsData.Region,
		Zone:          powervsData.Zone,
		UserAccount:   *account,
	}
	pclient.session, err = ibmpisession.NewIBMPISession(clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new session")
	}
	resourceLogger.Info("successfully created PowerVS client")

	pclient.instance = NewInstanceClient(pclient.session, powervsData.ServiceInstanceID)
	pclient.network = NewNetworkClient(pclient.session, powervsData.ServiceInstanceID)
	pclient.resource = options.Resource

	return pclient, nil
}
