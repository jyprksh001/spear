package config

import (
	"encoding/base64"
	"errors"
	"strings"

	"../../spear/network"
)

//Section refers a list of content under [name]
type Section struct {
	Name    string
	Content map[string]string
}

//Configuration refers to the content of a spear config file
type Configuration []*Section

//CreateClient creates a network.Client from Configuration
func CreateClient(config *Configuration) (*network.Client, error) {
	client := network.Client{}

	sections := config.GetSections("client")
	if len(sections) != 1 {
		return nil, errors.New("Multiple or no [client] found")
	}

	if err := readClientSection(sections[0], &client); err != nil {
		return nil, err
	}

	for _, section := range config.GetSections("peer") {
		if err := readPeerSection(section, &client); err != nil {
			return nil, err
		}
	}

	return &client, nil
}

func readClientSection(section *Section, client *network.Client) error {
	for key, value := range section.Content {
		switch key {
		case "secret":
			data, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return errors.New("Error decoding secret: " + err.Error())
			}
			client.SecretKey = data
		case "candidates":
			for _, addr := range readList(value) {
				parsedAddr, err := ParseAddr(addr)
				if err != nil {
					return err
				}
				client.Addr.Candidates = append(client.Addr.Candidates, parsedAddr)
			}
		default:
			return errors.New("Key " + key + " is not recognized")
		}
	}
	return nil
}

func readPeerSection(section *Section, client *network.Client) error {
	peer := network.Peer{}
	for key, value := range section.Content {
		switch key {
		case "pk":
			data, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return errors.New("Error decoding public key" + key + ": " + err.Error())
			}
			peer.PublicKey = data
		case "candidates":
			for _, addr := range readList(value) {
				parsedAddr, err := ParseAddr(addr)
				if err != nil {
					return err
				}
				peer.Addr.Candidates = append(peer.Addr.Candidates, parsedAddr)
			}
		default:
			return errors.New("Key " + key + " is not recognized")
		}
	}
	client.PeerList = append(client.PeerList, &peer)
	return nil
}

func readList(list string) []string {
	values := strings.Split(list, ",")
	for i, v := range values {
		values[i] = strings.TrimSpace(v)
	}
	return values
}
