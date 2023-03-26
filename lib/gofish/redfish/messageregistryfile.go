//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"encoding/json"
	common2 "hardware_exporter/lib/gofish/common"
)

// MessageRegistryFileLocation is a location
// information for the Message Registry file.
type MessageRegistryFileLocation struct {
	Language string `json:"Language"`
	URI      string `json:"Uri"`
}

// MessageRegistryFile describes the Message Registry file locator Resource.
type MessageRegistryFile struct {
	common2.Entity

	// ODataContext is the odata context.
	ODataContext string `json:"@odata.context"`
	// ODataType is the odata type.
	ODataType string `json:"@odata.type"`
	// Description provides a description of this resource.
	Description string
	// Languages is the RFC5646-conformant language codes for the
	// available Message Registries.
	Languages []string
	// Location is the location information for this Message Registry file.
	Location []MessageRegistryFileLocation
	// Registry shall contain the Message Registry name and it major and
	// minor versions, as defined by the Redfish Specification.
	Registry string
}

// GetMessageRegistryFile will get a MessageRegistryFile
// instance from the Redfish service.
func GetMessageRegistryFile(
	c common2.Client,
	uri string,
) (*MessageRegistryFile, error) {
	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var t MessageRegistryFile
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// ListReferencedMessageRegistryFiles gets the collection of MessageRegistryFile.
func ListReferencedMessageRegistryFiles(
	c common2.Client,
	link string,
) ([]*MessageRegistryFile, error) {
	var result []*MessageRegistryFile
	links, err := common2.GetCollection(c, link)
	if err != nil {
		return result, err
	}

	collectionError := common2.NewCollectionError()
	for _, sLink := range links.ItemLinks {
		s, err := GetMessageRegistryFile(c, sLink)
		if err != nil {
			collectionError.Failures[sLink] = err
		} else {
			result = append(result, s)
		}
	}

	if collectionError.Empty() {
		return result, nil
	}

	return result, collectionError
}
