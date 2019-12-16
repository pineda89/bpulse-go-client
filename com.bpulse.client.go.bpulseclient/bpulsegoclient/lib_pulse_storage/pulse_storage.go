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

package lib_pulse_storage

import(
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"

    "strings"
    "fmt"
    "github.com/golang/protobuf/proto"
)

/*
    This component implements the storage of the received pulses.

    This storage is done in a Fifo.

    Each entry into the Fifo, is a []* collector.Pulse

    Each slice into the fifo has a fixed length. When this length is reached,
    a new slice is created.
*/

/*
   POOL de pulse storage adapters. These, will use a sempahore to write to the common storage
*/

var Container IPulseStorageContainer

func init(){
    s_active_storages := lib_gc_config.ConfigurationProvider.GetPropertySTRING("PULSE_STORAGE","PULSE_STORAGE_ACTIVE_STORAGES")
    if _container, err := lib_gc_container.GenericContainerFactory.GetContainer(strings.Split(s_active_storages, ","), "STORAGES");err!=nil{
        panic(err)
    }else {
        Container = &PulseStorageContainer{Container:_container}
    }

    lib_gc_log.Info.Printf("lib_pulse_storage container package initalizaed. Active storages: %s \n", fmt.Sprint(s_active_storages))
}


type PulseRecord struct {
    Key string
    Pulses []proto.Message
    Pos int64
}

type IPulsesStorageAdapter interface{
    AddPulseRecord(record *PulseRecord) error
    GetPulseRecord() (*PulseRecord, error)
    lib_gc_container.IContainerStatusListener
    lib_gc_container.IGenericContents
}

type IPulseStorageContainer interface {
    AddPulseStorage(string, IPulsesStorageAdapter) error
    GetPulseStorageAdapter(string) (IPulsesStorageAdapter, error)
    IsStorageActivated(string) bool
    lib_gc_container.IContainerStatusListener
    lib_gc_container.IGenericContainerGetter
}

type PulseStorageContainer struct {
    Container lib_gc_container.IGenericContainer
}

func (container *PulseStorageContainer) GetGenericContainer() lib_gc_container.IGenericContainer{
    return container.Container
}

func(container *PulseStorageContainer) AddPulseStorage(name string, adapter IPulsesStorageAdapter) error {
    container.Container.AddItem(name, adapter)
    return nil
}

func(container *PulseStorageContainer) GetPulseStorageAdapter(name string) (IPulsesStorageAdapter, error) {
    if adapter,err := container.Container.GetItem(name);err!=nil{
        return nil, err
    }else{
        return adapter.(IPulsesStorageAdapter), nil
    }
}

func (container *PulseStorageContainer) IsStorageActivated(name string) bool{
    return container.Container.IsItemActivated(name)
}

func (container *PulseStorageContainer) Shutdown() error{
    return container.Container.Shutdown()
}

func (container *PulseStorageContainer) Start() error{
    return container.Container.Start()
}

func (container *PulseStorageContainer) Stop() error{
    return container.Container.Stop()
}