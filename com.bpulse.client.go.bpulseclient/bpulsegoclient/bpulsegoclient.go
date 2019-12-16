/*
Copyright 2015 - Serhstourism, S.A

Author Jaume Arús - jaumearus@gmail.com

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package bpulsegoclient

import (
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_aggregator"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_storage"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_send_channel_adapter"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulses_sender"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container/lib_gc_contents_loader"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_event"
    "errors"
    "strings"
    "container/list"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_panic_catching"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_bpulse_event"
)

func init() {
    // Load the container's contents.
    _ = Dummy

    // Initial BPulse client status
    status = BPULSECLIENT_NOT_STARTED
    
    lib_gc_log.Info.Println("Go BPulse client initialized !!!")
}

type BPULSECLIENT_STATUS int
const  (
  BPULSECLIENT_NOT_STARTED = iota
  BPULSECLIENT_STARTED
)
var status BPULSECLIENT_STATUS


/*
  This method waits for the specified container starts. This container may be for aggreators, storages, senders or sender channel adapters.
*/
func waitForContainerStarts(container lib_gc_container.IGenericContainer) error{
    for{
        if err := container.Start(); err!=nil {
            if !strings.Contains(err.Error(),"005-003") {
                return err
            }
        }else{
            break
        }
    }
    return nil
}

/*
  This method waits for the specified container is shutdown. This container may be for aggreators, storages, senders or sender channel adapters.
*/
func waitForContainerShutdown(container lib_gc_container.IGenericContainer) error{
    for{
        if err := container.Shutdown(); err!=nil {
            return err
        }else{
            break
        }
    }
    return nil
}

/*
  To start a BPulse client, apart from set the BPulse client configuration in BPulseGoClient.ini file, 
  Is necessary to pass to the client the rq and rs builders and a rs asynchronous processor. This last parameter is optional.
  The two first parameters are necessary because the exact implementation of the pulse to be sent won't be de same for each application.
  At the end, the RQ maker must to return a proto message complain with the definition of BPulse RQ message.
  The same for the RS maker.
  
  You can see what type of proto messages have to be sent at https://github.com/pineda89/bpulse-protobuf-go
*/


/*
    This method starts the Go BPulse client.

    Parameters:
        rq_maker: BPulse RQ builder from a slice of pulses
        rs_processor: BPulse responses writer. (optional)
        rs_instance_getter: BPulse RS intance getter.
        
    Returns
      error: The fired up error if some issue occurs.
*/
func StartBPulseClient(rq_maker lib_pulses_sender.PulseRQMaker, rs_processor lib_pulses_sender.PulseRSProcessor, rs_instance_getter lib_send_channel_adapter.PulseRSInstanceGetter) error {

    // panic catching
    defer lib_gc_panic_catching.PanicCatching("StartBPulseClient")

  if status == BPULSECLIENT_STARTED {
    msg,_ := lib_gc_event.NotifyEvent("008-003","",&[]string{})
    return errors.New(msg)
  }else{
    // Start the contents.
    lib_gc_contents_loader.InitializeContainerContents()

    // start event logger container
    if err := waitForContainerStarts(lib_bpulse_event.Container.GetGenericContainer());err!=nil {
      return err
    }else{
      // start storage container
      if err := waitForContainerStarts(lib_pulse_storage.Container.GetGenericContainer()); err != nil {
        return err
      }else {
        // start aggregator container
        if err := waitForContainerStarts(lib_pulse_aggregator.Container.GetGenericContainer()); err != nil {
          return err
        }else {

          // Set the bpulse rs instance getter for unmarsal bpulse respones. It's mandatory
          if rs_instance_getter == nil {
            msg, _ := lib_gc_event.NotifyEvent("008-002", "", &[]string{})
            return errors.New(msg)
          }else {
            lib_send_channel_adapter.Container.GetGenericContainer().AddParameter(lib_send_channel_adapter.RS_INSTANCE_GETTER, rs_instance_getter)

            // start sender channel adapter container
            if err := waitForContainerStarts(lib_send_channel_adapter.Container.GetGenericContainer()); err != nil {
              return err
            }else {

              // Set the rq maker. It's mandatory.
              if rq_maker == nil {
                msg, _ := lib_gc_event.NotifyEvent("008-001", "", &[]string{})
                return errors.New(msg)
              }else {
                lib_pulses_sender.Container.GetGenericContainer().AddParameter(lib_pulses_sender.PULSE_RQ_MAKER, rq_maker)

                // RS processor is optional.
                if rs_processor != nil {
                  lib_pulses_sender.Container.GetGenericContainer().AddParameter(lib_pulses_sender.PULSE_RS_PROCESSOR, rs_processor)
                }

                if err := waitForContainerStarts(lib_pulses_sender.Container.GetGenericContainer()); err != nil {
                  return err
                }
              }
            }
          }
        }
      }
    }

    status = BPULSECLIENT_STARTED
    return nil
  }
}


/*
 Shutdown the BPulse client
 
 Parameters:
 
 Returns
  The list of fired up errors if some issue has occurred at shutdown the containers.
*/ 
func ShutdownBPulseClient() *list.List{
  l := list.New()
  if status == BPULSECLIENT_NOT_STARTED{
    msg,_ := lib_gc_event.NotifyEvent("008-004","",&[]string{})
    l.PushBack(errors.New(msg))
  }else {

    if err := waitForContainerShutdown(lib_bpulse_event.Container.GetGenericContainer()); err != nil {
      l.PushBack(err)
    }else{
      if err := waitForContainerShutdown(lib_pulse_aggregator.Container.GetGenericContainer()); err != nil {
        l.PushBack(err)
      }else {
        if err := waitForContainerShutdown(lib_pulse_storage.Container.GetGenericContainer()); err != nil {
          l.PushBack(err)
        }else {
          if err := waitForContainerShutdown(lib_pulses_sender.Container.GetGenericContainer()); err != nil {
            l.PushBack(err)
          }else {
            if err := waitForContainerShutdown(lib_send_channel_adapter.Container.GetGenericContainer()); err != nil {
              l.PushBack(err)
            }
          }
        }
      }
    }
  }
  return l
}
