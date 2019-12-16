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

package lib_gc_config

import (
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
	"github.com/pineda89/go-misc/com.theskyinflames.go.misc/lib_gc_config"
)

func init() {
	ConfigurationProvider = lib_gc_config.PtrConfigProvider

	lib_gc_log.Info.Println("lib_gc_config package for bpulse client library initialized.")
}

var ConfigurationProvider lib_gc_config.IConfigurationProvider
