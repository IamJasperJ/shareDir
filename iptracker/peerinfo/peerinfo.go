package peerinfo

import (
	"slices"
)

type PeerManager interface {
	Update(clientID string, addr string)
	GetAllClient() []string
	GetIP(clientId string) string
	GetClientID(IP string) string
}

type peermanager struct {
	peersIDList []string
	// todo: mutex
	peersIP map[string]string
}

func GetPeerManager() PeerManager {
	return &peermanager{
		peersIDList: make([]string, 0),
		peersIP:     make(map[string]string),
	}
}

// update peers info
func (manager *peermanager) Update(clientID string, addr string) {
	if !slices.Contains(manager.peersIDList, clientID) {
		manager.peersIDList = append(manager.peersIDList, clientID)
	}
	manager.peersIP[clientID] = addr
}

// return all clients id
func (manager *peermanager) GetAllClient() []string {
	return manager.peersIDList
}

// return client's id according to ip
func (manager *peermanager) GetClientID(IP string) string {
	for id, addr := range manager.peersIP {
		if addr == IP {
			return id
		}
	}
	return ""
}

// return client's ip according to id
func (manager *peermanager) GetIP(clientID string) string {
	return manager.peersIP[clientID]
}
