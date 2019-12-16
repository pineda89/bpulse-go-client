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

package lib_pulses_sender

import (
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "io"
    "strings"
    "fmt"
    "github.com/golang/protobuf/proto"
)


/*
  This component sends a pulserq to the bpulse, on demand.

  URL
  Conmmand
  Auth (user, passwd)
  Timeout
*/

func init(){
    s_active_senders := lib_gc_config.ConfigurationProvider.GetPropertySTRING("PULSE_SENDER","PULSE_SENDER_ACTIVE_SENDERS")
    if _container, err := lib_gc_container.GenericContainerFactory.GetContainer(strings.Split(s_active_senders, ","), "SENDERS");err!=nil{
        panic(err)
    }else {
        Container = &PulseSenderContainer{Container:_container}
    }

    lib_gc_log.Info.Printf("lib_pulses_sender container package initalizaed. Active senders: %v\n",fmt.Sprint(s_active_senders))
}

var Container IPulseSenderContainer

type PulseRQMaker func(pulses []proto.Message) (proto.Message, error)
type PulseRSProcessor func(pulseRS proto.Message, rs_writer io.Writer) error
const PULSE_RQ_MAKER string = "PULSE_RQ_MAKER"
const PULSE_RS_PROCESSOR string = "PULSE_RS_PROCESSOR"


type IPulsesSender interface {
    lib_gc_container.IContainerStatusListener
    lib_gc_container.IGenericContents
    SetPulseRQMaker(PulseRQMaker) error
    SetPulseRSProcessor(PulseRSProcessor) error
}

type IPulseSenderContainer interface{
    AddPulseSender(string, IPulsesSender) error
    GetPulseSender(string) (IPulsesSender, error)
    IsPulseSenderActivated(string) bool
    lib_gc_container.IContainerStatusListener
    lib_gc_container.IGenericContainerGetter
}

type PulseSenderContainer struct{
    Container lib_gc_container.IGenericContainer
    response_writer io.ByteWriter
}

func (container *PulseSenderContainer) GetGenericContainer() lib_gc_container.IGenericContainer{
return container.Container
}

func(container *PulseSenderContainer) AddPulseSender(name string, sender IPulsesSender) error{
    container.Container.AddItem(name, sender)
    return nil
}

func(container *PulseSenderContainer) GetPulseSender(name string) (IPulsesSender, error){
    if sender,err:=container.Container.GetItem(name);err!=nil{
        return nil, err
    }else {
        return sender.(IPulsesSender), nil
    }
}

func(container *PulseSenderContainer) IsPulseSenderActivated(name string) bool{
    return container.Container.IsItemActivated(name)
}



func(container *PulseSenderContainer) Shutdown() error{
    return container.Container.Shutdown()
}

func(container *PulseSenderContainer) Start() error{
    container.Container.Start()
    return nil
}

func(container *PulseSenderContainer) Stop() error{
    container.Container.Stop()
    return nil
}
