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

package lib_pulse_aggregator

import (
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_storage"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_pool"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
    "strings"
    "fmt"
    "github.com/golang/protobuf/proto"
)

/*
 This component implements two functions:

     * Receives and store the pulses, using its pulses_storage
     * In intervals of time, check if is necessary send the pending pulses to bpulse.
       This check can be done by timing, or if a maximum number of pending pulses reached.
*/

/*
   POOL de pulse aggregators.
*/

func init(){
    s_active_aggregators := lib_gc_config.ConfigurationProvider.GetPropertySTRING("PULSE_AGGREGATOR","PULSE_AGGREGATOR_ACTIVE_AGGREGATORS")
    if _container, err := lib_gc_container.GenericContainerFactory.GetContainer(strings.Split(s_active_aggregators, ","), "AGGREGATORS");err!=nil{
        panic(err)
    }else {
        Container = &PulseAggregatorContainer{Container:_container}
    }

    lib_gc_log.Info.Printf("lib_pulse_aggregator container package initalizaed. Active aggregators: %s\n",fmt.Sprint(s_active_aggregators))
}

var Container IPulseAggregatorContainer

type IPulseAggregator interface{
    SetPulsesStorage(lib_pulse_storage.IPulsesStorageAdapter) error
    GetPulsesAggregationChannel() (chan<- proto.Message,error)
    lib_gc_container.IGenericContents
    lib_gc_container.IContainerStatusListener
}

type IPulseAggregatorContainer interface{
    AddPulseAggregatorPool(name string , pool *lib_gc_container.PooledContents) error
    GetPulseAggregator(name string) (IPulseAggregator, error)
    ReleasePulseAggregator(name string, aggregator IPulseAggregator) error
    IsAggregatorActivated(name string) bool
    lib_gc_container.IContainerStatusListener
    lib_gc_container.IGenericContainerGetter
}
type PulseAggregatorContainer struct {
    Container lib_gc_container.IGenericContainer
}

func (container *PulseAggregatorContainer) GetGenericContainer() lib_gc_container.IGenericContainer{
    return container.Container
}

func (container *PulseAggregatorContainer) AddPulseAggregatorPool(name string , pool *lib_gc_container.PooledContents) error {
    container.Container.AddItem(name, pool)
    return nil
}

func (container *PulseAggregatorContainer) GetPulseAggregator(name string) (IPulseAggregator, error) {
    if pool,err := container.Container.GetItem(name);err!=nil{
        return nil, err
    }else{
        poolable := <-pool.(lib_gc_pool.Pooler).GetPool()
        return poolable.Item.(IPulseAggregator), nil
    }
}

func (container *PulseAggregatorContainer) ReleasePulseAggregator(name string, aggregator IPulseAggregator) error {
    if pool,err := container.Container.GetItem(name);err!=nil{
        return err
    }else{
        pool.(lib_gc_pool.Pooler).GetPool()<- &lib_gc_pool.Poolable{Item:aggregator}
        return nil
    }
}
func (container *PulseAggregatorContainer) IsAggregatorActivated(name string) bool {
    return container.Container.IsItemActivated(name)
}


func (container *PulseAggregatorContainer) Shutdown() error{
    println("Container aggregator, Shutdown")
    return container.Container.Shutdown()
}

func (container *PulseAggregatorContainer) Start() error{
    println("Container aggregator, Start", container.Container)
    return container.Container.Start()
}

func (container *PulseAggregatorContainer) Stop() error{
    println("Container aggregator, Stop")
    return container.Container.Stop()
}

