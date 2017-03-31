// Package network handles container networks and interconnections
package network

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	networkTypes "github.com/docker/docker/api/types/network"
	"github.com/kontainerooo/kontainer.ooo/pkg/abstraction"
)

var (
	// ErrNetworkNotExist occurs when a network does not exist
	ErrNetworkNotExist = errors.New("Network does not exist")

	// ErrNetworkAlreadyExists occurs when a network already exists
	ErrNetworkAlreadyExists = errors.New("Network already exists")
)

// Service NetworkService
type Service interface {
	// CreateNetwork creates a new network for a given user
	CreateNetwork(refid uint, cfg *Config) error

	// RemoveNetwork removes a network with a given name
	RemoveNetworkByName(refid uint, name string) error

	// AddContainerToNetwork joins a given container to a given network
	AddContainerToNetwork(refid uint, name string, containerID string) error

	// RemoveContainerFromNetwork removes a container from a given network
	RemoveContainerFromNetwork(refid uint, name string, containerID string) error

	// ExposePortToContainer exposes a port from one container to another
	ExposePortToContainer(refid uint, srcContainerID string, port uint16, destContainerID string) error

	// RemovePortFromContainer removes an exposed port from a container
	RemovePortFromContainer(refid uint, srcContainerID string, port uint16, destContainerID string) error
}

type dbAdapter interface {
	abstraction.DBAdapter
	AutoMigrate(...interface{}) error
	Where(interface{}, ...interface{}) error
	First(interface{}, ...interface{}) error
	Find(interface{}, ...interface{}) error
	Create(interface{}) error
	Delete(interface{}, ...interface{}) error
}

type service struct {
	dcli   abstraction.DCli
	db     dbAdapter
	logger log.Logger
}

func (s *service) InitializeDatabases() error {
	return s.db.AutoMigrate(&Networks{}, &Containers{})
}

func (s *service) getNetworkByName(refid uint, name string) (Networks, error) {
	nw := Networks{}

	err := s.db.Where("user_id = ? AND network_name = ?", refid, name)
	if err != nil {
		return nw, err
	}

	s.db.First(&nw)

	return nw, nil
}

func (s *service) createNetwork(refid uint, cfg *Config, isPrimary bool) error {
	name := cfg.Name

	nw, err := s.getNetworkByName(refid, name)
	if err != nil {
		return err
	}

	if nw.NetworkID != "" {
		return ErrNetworkAlreadyExists
	}

	res, err := s.dcli.NetworkCreate(context.Background(), fmt.Sprintf("%s-%s", string(refid), name), types.NetworkCreate{
		Driver: cfg.Driver,
	})
	if err != nil {
		return err
	}

	s.db.Begin()
	nw = Networks{
		UserID:      uint(refid),
		NetworkName: name,
		NetworkID:   res.ID,
		IsPrimary:   isPrimary,
	}

	err = s.db.Create(&nw)
	fmt.Println(err)
	if err != nil {
		s.db.Rollback()
		// Try to remove the actual network on db error
		s.dcli.NetworkRemove(context.Background(), res.ID)
		return err
	}
	s.db.Commit()
	return nil
}

func (s *service) CreateNetwork(refid uint, cfg *Config) error {
	return s.createNetwork(refid, cfg, false)
}

func (s *service) CreatePrimaryNetworkForContainer(refid uint, cfg *Config, containerID string) error {
	err := s.createNetwork(refid, cfg, true)
	if err != nil {
		return err
	}

	err = s.AddContainerToNetwork(refid, cfg.Name, containerID)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) RemoveNetworkByName(refid uint, name string) error {
	nw, err := s.getNetworkByName(refid, name)
	if err != nil {
		return err
	}

	if nw.NetworkID != "" {
		s.db.Begin()
		err = s.dcli.NetworkRemove(context.Background(), nw.NetworkID)
		if err != nil {
			return err
		}

		err = s.db.Delete(&nw)
		if err != nil {
			s.db.Rollback()
			return err
		}

		cts := Containers{
			NetworkID: nw.NetworkID,
		}
		err = s.db.Delete(&cts)
		if err != nil {
			s.db.Rollback()
			return err
		}
		s.db.Commit()

		return nil
	}
	return ErrNetworkNotExist
}

func (s *service) AddContainerToNetwork(refid uint, name string, containerID string) error {
	nw, err := s.getNetworkByName(refid, name)
	if err != nil {
		return err
	}

	if nw.NetworkID != "" {
		err = s.dcli.NetworkConnect(context.Background(), nw.NetworkID, containerID, &networkTypes.EndpointSettings{})
		if err != nil {
			return err
		}

		s.db.Begin()
		cts := Containers{
			ContainerID: containerID,
			NetworkID:   nw.NetworkID,
		}

		err = s.db.Create(cts)
		if err != nil {
			s.db.Rollback()
			return err
		}
		s.db.Commit()

	} else {
		return ErrNetworkNotExist
	}

	return nil
}

func (s *service) RemoveContainerFromNetwork(refid uint, name string, containerID string) error {
	nw, err := s.getNetworkByName(refid, name)
	if err != nil {
		return err
	}
	if nw.NetworkID != "" {
		err = s.dcli.NetworkDisconnect(context.Background(), nw.NetworkID, containerID, true)
		if err != nil {
			return err
		}
	} else {
		return ErrNetworkNotExist
	}

	return nil
}

func (s *service) getContainerNetworks(containerID string) ([]Containers, error) {
	cts := []Containers{}

	s.db.Where("container_id = ?", containerID)

	err := s.db.Find(&cts)
	if err != nil {
		return []Containers{}, err
	}

	return cts, nil
}

func (s *service) isPrimary(networkID string) bool {
	err := s.db.Where("network_id = ? AND is_primary = ?", networkID, true)
	if err != nil {
		return false
	}

	if s.db.GetValue() != nil {
		return true
	}

	return false
}

func (s *service) getPrimaryNetworkForContainer(containerID string) Networks {
	cts, err := s.getContainerNetworks(containerID)
	if err != nil {
		return Networks{}
	}

	for _, v := range cts {
		if s.isPrimary(v.NetworkID) {
			s.db.Begin()
			var nw Networks
			s.db.Where("network_id = ?", v.NetworkID)
			s.db.First(&nw)
			s.db.Commit()

			if nw != (Networks{}) {
				return nw
			}
		}
	}

	return Networks{}
}

func (s *service) ExposePortToContainer(refid uint, srcContainerID string, port uint16, destContainerID string) error {
	// Check if the containers are in a same network
	srcNetworks, err := s.getContainerNetworks(srcContainerID)
	if err != nil {
		return err
	}

	dstNetworks, err := s.getContainerNetworks(destContainerID)
	if err != nil {
		return err
	}

	for _, srcV := range srcNetworks {
		for _, dstV := range dstNetworks {
			if srcV.NetworkID == dstV.NetworkID {
				return errors.New("Containers are already in the same network")
			}
		}
	}

	srcPrimaryNetwork := s.getPrimaryNetworkForContainer(srcContainerID)
	dstPrimaryNetwork := s.getPrimaryNetworkForContainer(destContainerID)

	if srcPrimaryNetwork == (Networks{}) || dstPrimaryNetwork == (Networks{}) {
		return errors.New("Both containers must have a primary network")
	}

	// TODO: talk to firewall service

	return nil
}

func (s *service) RemovePortFromContainer(refid uint, srcContainerID string, port uint16, destContainerID string) error {
	// TODO: implement
	return nil
}

// NewService creates a new network service
func NewService(dcli abstraction.DCli, db dbAdapter) (Service, error) {
	s := &service{
		dcli: dcli,
		db:   db,
	}

	err := s.InitializeDatabases()
	if err != nil {
		return s, err
	}

	return s, nil
}
