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

package lib_send_channel_adapter

import (
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_pool"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_event"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
    "github.com/golang/protobuf/proto"
    "strings"
    "errors"
    "fmt"
)

func init(){
    s_active_send_channel_adapters := lib_gc_config.ConfigurationProvider.GetPropertySTRING("SEND_CHANNEL_ADAPTER","SEND_CHANNEL_ADAPTER_ACTIVE_ADAPTERS")
    if _container, err := lib_gc_container.GenericContainerFactory.GetContainer(strings.Split(s_active_send_channel_adapters, ","),"CHANNEL_ADAPTERS");err!=nil{
        panic(err)
    }else {
        Container = &SendChannelAdapterContainer{Container:_container}
    }
    lib_gc_log.Info.Printf("lib_send_channel_adapter container pakge initalizaed. Active storages: %s\n",fmt.Sprint(s_active_send_channel_adapters))
}

var Container ISendChannelAdapterContainer

const RS_INSTANCE_GETTER = "RS_INSTANCE_GETTER"


type PulseRSInstanceGetter func()(proto.Message)



type ISendChannelAdapter interface {
    SendPulseRQ(proto.Message) (proto.Message, error)
    lib_gc_container.IContainerStatusListener
    lib_gc_container.IGenericContents
}


type ISendChannelAdapterContainer interface{
    AddSendChannelAdapter(string, *lib_gc_container.PooledContents) error
    GetSendChannelAdapter(string) (ISendChannelAdapter, error)
    ReleaseSendChannelAdapter(string, ISendChannelAdapter) error
    IsSendChannelAdapterActivated(string) bool
    lib_gc_container.IContainerStatusListener
    lib_gc_container.IGenericContainerGetter
}

type SendChannelAdapterContainer struct{
    Container lib_gc_container.IGenericContainer
}

func (container *SendChannelAdapterContainer) GetGenericContainer() lib_gc_container.IGenericContainer{
    return container.Container
}

func (container *SendChannelAdapterContainer) AddSendChannelAdapter(name string, pool *lib_gc_container.PooledContents) error{
    container.Container.AddItem(name, pool)
    return nil
}

func (container *SendChannelAdapterContainer) GetSendChannelAdapter(name string) (ISendChannelAdapter, error){
    if pool,err := container.takePool(name);err!=nil{
        return nil, err
    }else{
        poolable := <-pool.GetPool()
        return poolable.Item.(ISendChannelAdapter), nil
    }
}

func (container *SendChannelAdapterContainer) ReleaseSendChannelAdapter(name string, adapter ISendChannelAdapter) error{
    if pool,err := container.takePool(name);err!=nil{
        return err
    }else{
        poolable := &lib_gc_pool.Poolable{Item:adapter}
        pool.GetPool()<- poolable
        return nil
    }
}

func (container *SendChannelAdapterContainer) Shutdown() error{
    return container.Container.Shutdown()
}

func (container *SendChannelAdapterContainer) Start() error{
    container.Container.Start()
    return nil
}
func (container *SendChannelAdapterContainer) Stop() error{
    container.Container.Stop()
    return nil
}

func (container *SendChannelAdapterContainer) IsSendChannelAdapterActivated(name string) bool{
    return container.Container.IsItemActivated(name)
}


func (container *SendChannelAdapterContainer) takePool(name string) (lib_gc_pool.Pooler, error){

    if pool,err := container.Container.GetItem(name);err!=nil{
        _msg, _ := lib_gc_event.NotifyEvent("004-001", "", &[]string{name})
        return nil, errors.New(_msg)
    }else {
        return pool.(lib_gc_pool.Pooler), nil
    }
}
