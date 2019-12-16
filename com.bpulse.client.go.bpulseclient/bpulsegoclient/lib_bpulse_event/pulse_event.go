package lib_bpulse_event

import(
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"

	"strings"
	"fmt"
)


func init(){

	s_active_loggers := lib_gc_config.ConfigurationProvider.GetPropertySTRING("BPULSE_EVENT_LOGGER_SECTION","BPULSE_EVENT_LOGGER_ACTIVE_LOGGERS")
	if _container, err := lib_gc_container.GenericContainerFactory.GetContainer(strings.Split(s_active_loggers, ","), "EVENT_LOGGERS");err!=nil{
		panic(err)
	}else {
		Container = &EventLoggerContainer{Container:_container}
	}

	lib_gc_log.Info.Printf("lib_pulses_sender container package initalizaed. Active senders: %v\n",fmt.Sprint(s_active_loggers))
}
var Container IEventLoggerContainer

var Messages_chan chan *Event_message

type EVENT_TYPE string
const (
	WARN EVENT_TYPE = "WARN"
	ERROR EVENT_TYPE = "ERROR"
	NEWS EVENT_TYPE = "NEWS"
)
type Event_message struct {
	Type EVENT_TYPE
	Title string
	Description string
	StartDate int64
	EndDate int64
}

type IEventLogger interface{
	Send_event(*Event_message) error
}

type IEventLoggerContainer interface{
	lib_gc_container.IContainerStatusListener
	lib_gc_container.IGenericContainerGetter
}

type EventLoggerContainer struct{
	Container lib_gc_container.IGenericContainer
}

func(container *EventLoggerContainer) IsPulseSenderActivated(name string) bool{
	return container.Container.IsItemActivated(name)
}

func(container *EventLoggerContainer) GetGenericContainer() lib_gc_container.IGenericContainer{
	return container.Container
}


func(container *EventLoggerContainer) Shutdown() error{
	return container.Container.Shutdown()
}

func(container *EventLoggerContainer) Start() error{
	container.Container.Start()
	return nil
}

func(container *EventLoggerContainer) Stop() error{
	container.Container.Stop()
	return nil
}