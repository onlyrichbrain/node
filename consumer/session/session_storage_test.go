/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package session

import (
	"errors"
	"testing"
	"time"

	"github.com/mysteriumnetwork/node/consumer"
	"github.com/mysteriumnetwork/node/core/connection"
	"github.com/mysteriumnetwork/node/identity"
	discovery_dto "github.com/mysteriumnetwork/node/service_discovery/dto"
	node_session "github.com/mysteriumnetwork/node/session"
	"github.com/stretchr/testify/assert"
)

var (
	stubRetriever = &StubRetriever{}
	stubLocation  = &StubServiceDefinition{}

	errMock         = errors.New("error")
	sessionID       = node_session.ID("sessionID")
	consumerID      = identity.FromAddress("consumerID")
	providerID      = identity.FromAddress("providerID")
	serviceType     = "serviceType"
	providerCountry = "providerCountry"
	startTime       = time.Now()
	sessionStatus   = SessionStatusNew

	mockStatistics = consumer.SessionStatistics{BytesReceived: 10, BytesSent: 15}
	mockSession    = Session{
		SessionID:       sessionID,
		ProviderID:      providerID,
		ServiceType:     serviceType,
		ProviderCountry: providerCountry,
		Started:         startTime,
		Status:          sessionStatus,
	}
	mockPayload = connection.StateEvent{
		State: connection.Connected,
		SessionInfo: connection.SessionInfo{
			SessionID:  sessionID,
			ConsumerID: consumerID,
			Proposal: discovery_dto.ServiceProposal{
				ServiceDefinition: stubLocation,
				ServiceType:       serviceType,
				ProviderID:        providerID.Address,
			},
		},
	}
)

func TestSessionStorageSave(t *testing.T) {
	storer := &StubSessionStorer{}
	storage := NewSessionStorage(storer, stubRetriever)
	err := storage.Save(mockSession)
	assert.Nil(t, err)
	assert.True(t, storer.SaveCalled)
}

func TestSessionStorageSaveReturnsError(t *testing.T) {
	storer := &StubSessionStorer{
		SaveError: errMock,
	}

	storage := NewSessionStorage(storer, stubRetriever)
	err := storage.Save(mockSession)
	assert.Equal(t, errMock, err)
}

func TestSessionStorageUpdate(t *testing.T) {
	storer := &StubSessionStorer{}
	storage := NewSessionStorage(storer, stubRetriever)
	err := storage.Update(sessionID, startTime, mockStatistics, sessionStatus)
	assert.Nil(t, err)
	assert.True(t, storer.UpdateCalled)
}

func TestSessionStorageUpdateReturnsError(t *testing.T) {
	storer := &StubSessionStorer{
		UpdateError: errMock,
	}
	storage := NewSessionStorage(storer, stubRetriever)
	err := storage.Update(sessionID, startTime, mockStatistics, sessionStatus)
	assert.Equal(t, errMock, err)
}

func TestSessionStorageGetAll(t *testing.T) {
	storer := &StubSessionStorer{}
	storage := NewSessionStorage(storer, stubRetriever)
	sessions, err := storage.GetAll()
	assert.Nil(t, err)
	assert.True(t, storer.GetAllCalled)
	assert.Len(t, sessions, 0)
}

func TestSessionStorageGetAllReturnsError(t *testing.T) {
	storer := &StubSessionStorer{
		GetAllError: errMock,
	}
	storage := NewSessionStorage(storer, stubRetriever)
	sessions, err := storage.GetAll()
	assert.NotNil(t, err)
	assert.True(t, storer.GetAllCalled)
	assert.Nil(t, sessions)
}

func TestSessionStorageConsumeEventDisconnectingOK(t *testing.T) {
	storer := &StubSessionStorer{}

	storage := NewSessionStorage(storer, stubRetriever)
	storage.ConsumeStateEvent(connection.StateEvent{
		State: connection.Disconnecting,
	})
	assert.True(t, storer.UpdateCalled)
}

func TestSessionStorageConsumeEventDisconnectingErrors(t *testing.T) {
	storer := &StubSessionStorer{
		UpdateError: errMock,
	}

	storage := NewSessionStorage(storer, stubRetriever)
	assert.NotPanics(t, func() {
		storage.ConsumeStateEvent(connection.StateEvent{State: connection.Disconnecting})
	})

	assert.True(t, storer.UpdateCalled)
}

func TestSessionStorageConsumeEventConnectedOK(t *testing.T) {
	storer := &StubSessionStorer{}

	storage := NewSessionStorage(storer, stubRetriever)
	storage.ConsumeStateEvent(mockPayload)
	assert.True(t, storer.SaveCalled)
}

func TestSessionStorageConsumeEventConnectedError(t *testing.T) {
	storer := &StubSessionStorer{
		SaveError: errMock,
	}
	storage := NewSessionStorage(storer, stubRetriever)
	assert.NotPanics(t, func() {
		storage.ConsumeStateEvent(mockPayload)
	})
	assert.True(t, storer.SaveCalled)
}

// StubSessionStorer allows us to get all sessions, save and update them
type StubSessionStorer struct {
	SaveError    error
	SaveCalled   bool
	UpdateError  error
	UpdateCalled bool
	GetAllCalled bool
	GetAllError  error
}

func (sss *StubSessionStorer) Save(object interface{}) error {
	sss.SaveCalled = true
	return sss.SaveError
}

func (sss *StubSessionStorer) Update(object interface{}) error {
	sss.UpdateCalled = true
	return sss.UpdateError
}

func (sss *StubSessionStorer) GetAll(array interface{}) error {
	sss.GetAllCalled = true
	return sss.GetAllError
}

type StubRetriever struct {
	Value consumer.SessionStatistics
}

func (sr *StubRetriever) Retrieve() consumer.SessionStatistics {
	return sr.Value
}

type StubServiceDefinition struct{}

func (fs *StubServiceDefinition) GetLocation() discovery_dto.Location { return discovery_dto.Location{} }
