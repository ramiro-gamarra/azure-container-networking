package nmagentclient

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/Azure/azure-container-networking/cns/logger"
	"github.com/Azure/azure-container-networking/common"
)

const (
	//GetNmAgentSupportedApiURLFmt Api endpoint to get supported Apis of NMAgent
	GetNmAgentSupportedApiURLFmt     = "http://%s/machine/plugins/?comp=nmagent&type=GetSupportedApis"
	GetNetworkContainerVersionURLFmt = "http://%s/machine/plugins/?comp=nmagent&type=NetworkManagement/interfaces/%s/networkContainers/%s/version/authenticationToken/%s/api-version/1"
)

//WireServerIP - wire server ip
var WireserverIP = "168.63.129.16"

// NMANetworkContainerResponse - NMAgent response.
type NMANetworkContainerResponse struct {
	ResponseCode       string `json:"httpStatusCode"`
	NetworkContainerID string `json:"networkContainerId"`
	Version            string `json:"version"`
}

type NMAgentSupportedApisResponseXML struct {
	SupportedApis []string `xml:"type"`
}

// JoinNetwork joins the given network
func JoinNetwork(
	networkID string,
	joinNetworkURL string) (*http.Response, error) {
	logger.Printf("[NMAgentClient] JoinNetwork: %s", networkID)

	// Empty body is required as wireserver cannot handle a post without the body.
	var body bytes.Buffer
	json.NewEncoder(&body).Encode("")
	response, err := common.GetHttpClient().Post(joinNetworkURL, "application/json", &body)

	if err == nil && response.StatusCode == http.StatusOK {
		defer response.Body.Close()
	}

	logger.Printf("[NMAgentClient][Response] Join network: %s. Response: %+v. Error: %v",
		networkID, response, err)

	return response, err
}

// PublishNetworkContainer publishes given network container
func PublishNetworkContainer(
	networkContainerID string,
	createNetworkContainerURL string,
	requestBodyData []byte) (*http.Response, error) {
	logger.Printf("[NMAgentClient] PublishNetworkContainer NC: %s", networkContainerID)

	requestBody := bytes.NewBuffer(requestBodyData)
	response, err := common.GetHttpClient().Post(createNetworkContainerURL, "application/json", requestBody)

	logger.Printf("[NMAgentClient][Response] Publish NC: %s. Response: %+v. Error: %v",
		networkContainerID, response, err)

	return response, err
}

// UnpublishNetworkContainer unpublishes given network container
func UnpublishNetworkContainer(
	networkContainerID string,
	deleteNetworkContainerURL string) (*http.Response, error) {
	logger.Printf("[NMAgentClient] UnpublishNetworkContainer NC: %s", networkContainerID)

	// Empty body is required as wireserver cannot handle a post without the body.
	var body bytes.Buffer
	json.NewEncoder(&body).Encode("")
	response, err := common.GetHttpClient().Post(deleteNetworkContainerURL, "application/json", &body)

	logger.Printf("[NMAgentClient][Response] Unpublish NC: %s. Response: %+v. Error: %v",
		networkContainerID, response, err)

	return response, err
}

// GetNetworkContainerVersion :- Retrieves NC version from NMAgent
func GetNetworkContainerVersion(
	networkContainerID,
	getNetworkContainerVersionURL string) (*http.Response, error) {
	logger.Printf("[NMAgentClient] GetNetworkContainerVersion NC: %s, URL: %s", networkContainerID, getNetworkContainerVersionURL)

	response, err := common.GetHttpClient().Get(getNetworkContainerVersionURL)

	logger.Printf("[NMAgentClient][Response] GetNetworkContainerVersion NC: %s. Response: %+v. Error: %v",
		networkContainerID, response, err)
	return response, err
}

// GetNmAgentSupportedApis :- Retrieves Supported Apis from NMAgent
func GetNmAgentSupportedApis(
	httpc *http.Client,
	getNmAgentSupportedApisURL string) ([]string, error) {
	var (
		returnErr error
	)

	if getNmAgentSupportedApisURL == "" {
		getNmAgentSupportedApisURL = fmt.Sprintf(
			GetNmAgentSupportedApiURLFmt, WireserverIP)
	}

	response, err := httpc.Get(getNmAgentSupportedApisURL)
	if err != nil {
		returnErr = fmt.Errorf(
			"Failed to retrieve Supported Apis from NMAgent with error %v",
			err.Error())
		logger.Errorf("[Azure-CNS] %s", returnErr)
		return nil, returnErr
	}
	if response == nil {
		returnErr = fmt.Errorf(
			"Response from getNmAgentSupportedApis call is <nil>")
		logger.Errorf("[Azure-CNS] %s", returnErr)
		return nil, returnErr
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		returnErr = fmt.Errorf(
			"Failed to retrieve Supported Apis from NMAgent with StatusCode: %d",
			response.StatusCode)
		logger.Errorf("[Azure-CNS] %s", returnErr)
		return nil, returnErr
	}

	var xmlDoc NMAgentSupportedApisResponseXML
	decoder := xml.NewDecoder(response.Body)
	err = decoder.Decode(&xmlDoc)
	if err != nil {
		returnErr = fmt.Errorf(
			"Failed to decode XML response of Supported Apis from NMAgent with error %v",
			err.Error())
		logger.Errorf("[Azure-CNS] %s", returnErr)
		return nil, returnErr
	}

	logger.Printf("[NMAgentClient][Response] GetNmAgentSupportedApis. Response: %+v.", response)
	return xmlDoc.SupportedApis, nil
}
