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
    "github.com/golang/protobuf/proto"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_pool"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_event"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_send_channel_adapter"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container/lib_gc_contents_loader"
    "bytes"
    "time"
    "io/ioutil"
    "errors"
    "net/http"
    "sync"
)

/*
This component sends a pulse rq to bpulse, on demand.

URL
Conmmand
Auth (user, passwd)
Timeout
*/

func init() {
    // Locks the container until this content has been put in it.
    wg := lib_send_channel_adapter.Container.(*lib_send_channel_adapter.SendChannelAdapterContainer).Container.GetWg()
    wg.Add(1)
    lib_gc_contents_loader.AddInitiator("DEFAULT_RESTFUL", Initialize)
    wg.Done()
}

func Initialize() error {
    if lib_send_channel_adapter.Container.IsSendChannelAdapterActivated(RESTFUL_SEND_CHANNEL_ADAPTER) {
        // Create the client
        timeout := time.Duration(lib_gc_config.ConfigurationProvider.GetPropertyINT64("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_TIMEOUT_MS")) * time.Millisecond
        max_idle_connections := lib_gc_config.ConfigurationProvider.GetPropertyINT("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_MAX_IDLE_CONNECTIONS")

        pool_size := lib_gc_config.ConfigurationProvider.GetPropertyINT("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_POOL_SIZE")
        url_protocol := lib_gc_config.ConfigurationProvider.GetPropertySTRING("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_PROTOCOL")
        url_host := lib_gc_config.ConfigurationProvider.GetPropertySTRING("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_HOST")
        url_port := lib_gc_config.ConfigurationProvider.GetPropertySTRING("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_URL_PORT")
        url_path := lib_gc_config.ConfigurationProvider.GetPropertySTRING("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_URL_PATH")
        url := url_protocol+"://"+url_host+":"+url_port+url_path
        method := lib_gc_config.ConfigurationProvider.GetPropertySTRING("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_METHOD")
        user := lib_gc_config.ConfigurationProvider.GetPropertySTRING("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_USER")
        pwd := lib_gc_config.ConfigurationProvider.GetPropertySTRING("RESTFUL_SEND_CHANNEL_ADAPTER", "RESTFUL_SEND_CHANNEL_ADAPTER_URL_PWD")

        pool := lib_gc_pool.POOLMaker.GetPool(int32(pool_size))
        for z := 0; z<pool_size; z++ {
            http_client := &http.Client{
                Transport: &http.Transport{
                    MaxIdleConnsPerHost: max_idle_connections,
                },
                Timeout: time.Duration(timeout) * time.Second,
            }
            adapter := &RESTfulSendChannelAdapter{http_client:http_client, url:url, method:method, user:user, pwd:pwd, once:&sync.Once{}}
            adapter.Contents = lib_gc_container.Contents{lib_send_channel_adapter.Container.(*lib_send_channel_adapter.SendChannelAdapterContainer).Container, lib_gc_container.CONTAINER_STATUS_STOPPED, adapter}
            pool.GetPool() <- &lib_gc_pool.Poolable{Item:adapter}
        }
        lib_send_channel_adapter.Container.AddSendChannelAdapter(RESTFUL_SEND_CHANNEL_ADAPTER, &lib_gc_container.PooledContents{Pooler:pool})

        lib_gc_log.Info.Printf("%s RESTful channel adapter initialized.\n", RESTFUL_SEND_CHANNEL_ADAPTER)
    }
    return nil
}

const RESTFUL_SEND_CHANNEL_ADAPTER = "DEFAULT"

var (

)

type RESTfulSendChannelAdapter struct {
    lib_gc_container.Contents
    http_client *http.Client
    url string
    method string
    user string
    pwd string
    pulse_rs_instance_getter lib_send_channel_adapter.PulseRSInstanceGetter
    once *sync.Once
}


func (adapter *RESTfulSendChannelAdapter) Shutdown() error {
    lib_gc_log.Info.Printf("ADAPTER content [%p, %s], container [%p, %s] SHUTDOWN catched. \n", &adapter.Contents, RESTFUL_SEND_CHANNEL_ADAPTER, adapter.Container, adapter.Container.GetName())
    return nil
}

func (adapter *RESTfulSendChannelAdapter) Stop() error {
    lib_gc_log.Info.Printf("ADAPTER content [%p, %s], container [%p, %s] STOP catched. \n", &adapter.Contents, RESTFUL_SEND_CHANNEL_ADAPTER, adapter.Container, adapter.Container.GetName())
    return nil
}

func (adapter *RESTfulSendChannelAdapter) Start() error {
    lib_gc_log.Info.Printf("ADAPTER content [%p, %s], container [%p, %s] START catched. \n", &adapter.Contents, RESTFUL_SEND_CHANNEL_ADAPTER, adapter.Container, adapter.Container.GetName())

    // Set bpulse rs instance gettter
    // It's necessary to unmarshal the responses
    f := func() {
        if rs_instance_getter, err := adapter.Container.(lib_gc_container.IGenericContainer).GetParameter(lib_send_channel_adapter.RS_INSTANCE_GETTER);err!=nil {
            msg,_ := lib_gc_event.NotifyEvent("004-006","",&[]string{})
            lib_gc_log.Error.Println(msg)
        }else{
            adapter.pulse_rs_instance_getter = rs_instance_getter.(lib_send_channel_adapter.PulseRSInstanceGetter)
        }
    }
   adapter.once.Do(f)

    return nil
}

func (adapter *RESTfulSendChannelAdapter) SendPulseRQ(pulserq proto.Message) (proto.Message, error) {

    // Marshal the rq
    if _bytes, err := proto.Marshal(pulserq); err!=nil {
        msg, _ := lib_gc_event.NotifyEvent("004-005", "", &[]string{adapter.url, err.Error()})
        return nil, errors.New(msg)
    }else {
        lib_gc_log.Trace.Printf("ADAPTER content [%p, %s], container [%p, %s] PULSE RQ: %v+ \n", &adapter.Contents, RESTFUL_SEND_CHANNEL_ADAPTER, adapter.Container, adapter.Container.GetName(), pulserq)

            // build the request
            if request,err := http.NewRequest(adapter.method, adapter.url, bytes.NewBuffer(_bytes)); err!=nil {
                msg, _ := lib_gc_event.NotifyEvent("004-001", "", &[]string{adapter.url, err.Error()})
                return nil, errors.New(msg)
            }else {

                // Set headers
                request.Header.Set("Content-Type", "application/x-protobuf")
                request.Header.Set("Accept", "application/x-protobuf")

                // Set the request auth
                request.SetBasicAuth(adapter.user, adapter.pwd)

                // Do the request
                if response, err := adapter.http_client.Do(request); err!=nil {
                    msg, _ := lib_gc_event.NotifyEvent("004-002", "", &[]string{adapter.url, err.Error()})
                    return nil, errors.New(msg)
                }else {

                    // Close the response body to reuse the connection.
                    defer response.Body.Close()

                    // Read the response
                    if body, err := ioutil.ReadAll(response.Body); err!=nil {
                        msg, _ := lib_gc_event.NotifyEvent("004-003", "", &[]string{adapter.url, err.Error()})
                        return nil, errors.New(msg)
                    }else {
                        rs := adapter.pulse_rs_instance_getter()
                        if err := proto.Unmarshal(body, rs); err!=nil {
                            msg, _ := lib_gc_event.NotifyEvent("004-004", "", &[]string{adapter.url, err.Error(), string(body)})
                            return nil, errors.New(msg)
                        }else {
                            lib_gc_log.Info.Printf("ADAPTER content [%p, %s], container [%p, %s] Sent %d bytes, RS [%d] received from BPulse: %v \n", &adapter.Contents, RESTFUL_SEND_CHANNEL_ADAPTER, adapter.Container, adapter.Container.GetName(), len(_bytes), response.StatusCode, rs.String())
                            return rs, nil
                        }
                    }
                }
            }
    }
}

