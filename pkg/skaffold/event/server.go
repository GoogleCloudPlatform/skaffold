/*
Copyright 2019 The Skaffold Authors

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

package event

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/event/proto"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

type server struct{}

func (s *server) GetState(context.Context, *empty.Empty) (*proto.State, error) {
	return ev.state, nil
}

func (s *server) EventLog(stream proto.SkaffoldService_EventLogServer) error {
	ev.logLock.Lock()
	for _, entry := range ev.eventLog {
		if err := stream.Send(&entry); err != nil {
			return err
		}
	}
	ev.logLock.Unlock()
	c := make(chan proto.LogEntry)
	ev.RegisterListener(c)
	var entry proto.LogEntry
	for {
		entry = <-c
		if err := stream.Send(&entry); err != nil {
			return err
		}
	}
}

func (s *server) Handle(ctx context.Context, event *proto.Event) (*empty.Empty, error) {
	if event != nil {
		handle(event)
	}
	return &empty.Empty{}, nil
}

// newStatusServer creates the grpc server for serving the state and event log.
func newStatusServer(originalPort int) (func() error, error) {
	if originalPort == -1 {
		return func() error { return nil }, nil
	}
	port := util.GetAvailablePort(originalPort, &sync.Map{})
	if port != originalPort && originalPort != constants.DefaultRPCPort {
		logrus.Warnf("provided port %d already in use: using %d instead", originalPort, port)
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return func() error { return nil }, errors.Wrap(err, "creating listener")
	}
	logrus.Infof("starting gRPC server on port %d", port)

	s := grpc.NewServer()
	proto.RegisterSkaffoldServiceServer(s, &server{})

	go func() {
		if err := s.Serve(l); err != nil {
			logrus.Errorf("failed to start grpc server: %s", err)
		}
	}()
	return func() error {
		s.Stop()
		return l.Close()
	}, nil
}
