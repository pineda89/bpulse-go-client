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
    STORAGE "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_storage/contents"
    AGGREGATOR "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_aggregator/contents"
    CHANNEL_ADAPTER "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_send_channel_adapter/contents"
    SENDER "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulses_sender/contents"
    EVENTS "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_bpulse_event/contents"
)

func init(){
    _ = STORAGE.DEFAULT_STORAGE_NAME
    _ = AGGREGATOR.DEFAULT_AGGREGATOR_NAME
    _ = CHANNEL_ADAPTER.RESTFUL_SEND_CHANNEL_ADAPTER
    _ = SENDER.DEFAULT_SENDER
    _ = EVENTS.DEFAULT_EVENT_LOGGER
}

var Dummy struct{}

