package contents

import(
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_bpulse_event"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container/lib_gc_contents_loader"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_panic_catching"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_pool"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_event"

	"github.com/parnurzeal/gorequest"


	"net/http"
	"crypto/tls"

	"time"
	"strconv"
	"errors"
	"encoding/json"
	"sync"
	"strings"
	"encoding/base64"
)

func init(){
	// Locks the container until this content has been put in it.
	wg := lib_bpulse_event.Container.(*lib_bpulse_event.EventLoggerContainer).Container.GetWg()
	wg.Add(1)
	lib_gc_contents_loader.AddInitiator(DEFAULT_EVENT_LOGGER, Initialize)
	wg.Done()
}

const (
	DEFAULT_EVENT_LOGGER = "DEFAULT_EVENT_LOGGER"
	BPULSE_EVENT_LOGGER_SECTION ="BPULSE_EVENT_LOGGER_SECTION"
)



func Initialize() error{

	if lib_bpulse_event.Container.GetGenericContainer().IsItemActivated(DEFAULT_EVENT_LOGGER){

		// Take configuration properties
  	_timeout := time.Duration(lib_gc_config.ConfigurationProvider.GetPropertyINT64(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_TIMEOUT_MS")) * time.Millisecond
		_url_prefix := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_URL_PREFIX")
		_url_path := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BUPLSE_EVENT_LOGGER_URL_PATH")
		_host := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_HOST")
		_port := lib_gc_config.ConfigurationProvider.GetPropertyINT(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_PORT")
		_user := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_USER")
		_password := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_PASSWORD")
		_source := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_SOURCE")
		_data_source_id := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_DATA_SOURCE_ID")
		_editable := lib_gc_config.ConfigurationProvider.GetPropertySTRING(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_EDITABLE")
		_reentries := lib_gc_config.ConfigurationProvider.GetPropertyINT(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_SENDING_RETRIES")
		_server := &event_logger_server{	timeout: _timeout,
																		user:_user,
																		password:_password,
																		source:_source,
																		data_source_id:_data_source_id,
																		editable:_editable,
																		reentries:_reentries}
		_server.make_url(_url_prefix,_host,_url_path,_port)


		pool_size := lib_gc_config.ConfigurationProvider.GetPropertyINT(BPULSE_EVENT_LOGGER_SECTION, "BPULSE_EVENT_LOGGER_POOL_SIZE")
		lib_bpulse_event.Messages_chan = make(chan *lib_bpulse_event.Event_message, pool_size)

		// Create the pool
		pool := lib_gc_pool.POOLMaker.GetPool(int32(pool_size))
		for z := 0; z<pool_size; z++ {
			event_logger := &EventLogger{server:_server, is_running:false,start_once:&sync.Once{}}
			event_logger.Contents = lib_gc_container.Contents{lib_bpulse_event.Container.(*lib_bpulse_event.EventLoggerContainer).Container, lib_gc_container.CONTAINER_STATUS_STOPPED, event_logger}
			pool.GetPool() <- &lib_gc_pool.Poolable{Item:event_logger}
		}
		lib_bpulse_event.Container.GetGenericContainer().AddItem(DEFAULT_EVENT_LOGGER, &lib_gc_container.PooledContents{Pooler:pool})
	}
	return nil
}

type event_logger_server struct{
	timeout time.Duration
	url string
	user string
	password string
	source string
	data_source_id string
	editable string
	version string
	reentries int
}

func (es *event_logger_server) make_url(url_prefix,host,url_path string, port int) {
	es.url = url_prefix+"://"+host+":"+ strconv.FormatInt(int64(port),10)+url_path
}


type EventLogger struct{
	lib_gc_container.Contents
	server *event_logger_server
	is_running bool
	start_once *sync.Once
}

func (ev *EventLogger) Send_event(event_message *lib_bpulse_event.Event_message) error{

	m := map[string]interface{}{
		"id": strconv.FormatInt(time.Now().UnixNano(), 10),
		"type": string(event_message.Type),
		"title" : event_message.Title,
		"description" : event_message.Description,
		"startDate:" : event_message.StartDate,
		"endDate" : event_message.EndDate,
		"source" : ev.server.source,
		"dataSourceId" : ev.server.data_source_id,
		"editable" : ev.server.editable,
	}
	mJson, _ := json.Marshal(m)
	println("\n",string(mJson),"\n")

	var _err error
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
	}
	for z:=0;z<ev.server.reentries;z++ {
		request := gorequest.New().Timeout(time.Duration(ev.server.timeout) * time.Millisecond)
		request.Transport = tr
		request.Header["Authorization"]="Basic "+ev.basicAuth(ev.server.user,ev.server.user)
		_, body, errs := request.Post(ev.server.url).
		Set("Content-Type", "application/json").
		Send(string(mJson)).
		End()

		println("*jas* body: ",body)

		if strings.Contains(body,"error"){
			if errs==nil{
				errs = []error{errors.New(body)}
			}else{
				errs = append(errs, errors.New(body))
			}
		}

		if errs == nil || len(errs)==0{
			break;
		}else{

			str := ""
			for i,s := range errs{
				if i > 0{
					str += ", "+s.Error()
				}else{
					str = s.Error()
				}
			}
			msg,_ := lib_gc_event.NotifyEvent("009-001","",&[]string{str})
			lib_gc_log.Error.Println(msg)
			_err = errors.New(str)

			// Wait for the next retry
			timer := time.NewTimer(500 * time.Millisecond)
			<-timer.C
		}
	}

	if _err!=nil{
		return _err
	}else {
		return nil
	}
}

func (ev *EventLogger) basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (ev *EventLogger) Start() error{
	lib_gc_log.Info.Printf("EVENT_LOGGER content [%p, %s], container [%p, %s] START catched. \n", &ev.Contents, DEFAULT_EVENT_LOGGER, ev.Container, ev.Container.GetName())

	f := func(){
		ev._run()
	}
	ev.start_once.Do(f)

	return nil
}

func (ev *EventLogger) Shutdown() error{
	lib_gc_log.Info.Printf("EVENT_LOGGER content [%p, %s], container [%p, %s] SHUTDOWN catched. \n", &ev.Contents, DEFAULT_EVENT_LOGGER, ev.Container, ev.Container.GetName())

	return nil
}

// Check for the event logger sending works
func (ev *EventLogger) check_for_running(){
	ticker := time.NewTicker(time.Millisecond * 500)
	go func(){
		for t := range ticker.C{
			_ = t
			if !ev.is_running {
				go ev._run()
			}
		}
	}()
}

// Event logger sending
func (ev *EventLogger) _run(){
	defer func(){
		lib_gc_panic_catching.PanicCatching("EventLogger.run")
		ev.is_running = false
	}()

	lib_gc_log.Info.Printf("EVENT_LOGGER content [%p, %s], container [%p, %s] RUN STARTED.\n", &ev.Contents, DEFAULT_EVENT_LOGGER, ev.Container, ev.Container.GetName())
	ev.is_running = true
	for {
		if ev.Contents.Status == lib_gc_container.CONTAINER_STATUS_STARTED {
			if message,ok := <-lib_bpulse_event.Messages_chan;ok{

				// Send the event
				if err :=ev.Send_event(message);err!=nil{
					lib_gc_log.Error.Println("ERROR on sending an event: ",err.Error())
				}
			}else{

				// The channel is closed !!!
				lib_gc_log.Error.Printf("ERROR: The event logger channel is closed !!! \n")
			}
		}else{

			// Wait for the next message
			timer := time.NewTimer(500 * time.Millisecond)
			<-timer.C
		}
	}

}