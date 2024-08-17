package test_http_client

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/svr"
	"Supernove2024/util"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"log"
	"net/http"
	"testing"
	"time"
)

const jAddress = "127.0.0.1:6831"

func SetupClient(
	service string,
	targetService string,
	configFile string,
	host string,
	port int32,
	selfKey string,
	targetKey string,
) {
	//链路追踪
	jCfg := jaegerConfig.Configuration{
		ServiceName: selfKey,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jAddress,
		},
	}
	tracer, closer, err := jCfg.NewTracer(
		jaegerConfig.Logger(jaeger.StdLogger),
		jaegerConfig.Metrics(metrics.NullFactory),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	//启动服务器
	address := fmt.Sprintf("%s:%v", host, port)
	config.GlobalConfigFilePath = configFile
	registerAPI, err := sdk.NewRegisterAPI()
	if err != nil {
		log.Fatalln(err)
	}

	reply, err := registerAPI.Register(&sdk.RegisterArgv{
		ServiceName: service,
		Name:        selfKey,
		Host:        host,
		Port:        port,
	})
	if err != nil {
		log.Fatalln(err)
	}

	defer registerAPI.Deregister(&sdk.DeregisterArgv{
		ServiceName: service,
		InstanceID:  reply.InstanceID,
	})

	go func() {
		healthAPI, err := sdk.NewHealthAPI()
		if err != nil {
			log.Fatalln(err)
		}
		for {
			err = healthAPI.HeartBeatOnce(&sdk.HeartBeatArgv{
				ServiceName: service,
				InstanceID:  reply.InstanceID,
			})
			if err == nil {
				time.Sleep(5 * time.Second)
			} else {
				util.Error("keepalive err: %v", err)
			}
		}
	}()

	timeout := int64(10000000)
	_, err = registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"Key": selfKey},
		DstServiceName:  service,
		DstInstanceName: []string{selfKey},
		Timeout:         &timeout,
	})
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		httpClient := &http.Client{}
		discoveryAPI, err := sdk.NewDiscoveryAPI()
		if err != nil {
			log.Fatalln(err)
		}
		var router *sdk.ProcessRouterResult = nil
		for {
			time.Sleep(1 * time.Second)
			func() {
				router, err = discoveryAPI.ProcessRouter(&sdk.ProcessRouterArgv{
					Method:          util.KVRouterType,
					SrcInstanceName: selfKey,
					DstService:      targetService,
					Key:             map[string]string{"Key": targetKey},
				})

				if router == nil {
					return
				}

				span := tracer.StartSpan("client-request")
				defer span.Finish()

				address := fmt.Sprintf("http://%s:%v", router.DstInstance.GetHost(), router.DstInstance.GetPort())
				req, err := http.NewRequest("GET", address, nil)
				if err != nil {
					util.Error("http get err: %v", err)
					return
				}

				ext.HTTPUrl.Set(span, address)
				ext.HTTPMethod.Set(span, req.Method)
				err = tracer.Inject(span.Context(),
					opentracing.HTTPHeaders,
					opentracing.HTTPHeadersCarrier(req.Header))
				if err != nil {
					util.Error("trace inject err: %v", err)
					return
				}

				response, err := httpClient.Do(req)
				if err != nil {
					util.Error("body close err: %v", err)
					return
				}
				defer response.Body.Close()
			}()

		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		span := tracer.StartSpan("hello-world-reply", opentracing.ChildOf(spanCtx))
		defer span.Finish()

		_, err := fmt.Fprintf(w, "Hello world")
		if err != nil {
			util.Error("http err: %v", err)
		}
	})

	if err = http.ListenAndServe(address, mux); err != nil {
		log.Fatalln(err)
	}
}

func TestHttp(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("client-svr.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		if err = srv.Serve("127.0.0.1:30000"); err != nil {
			log.Fatalln(err)
		}
	}()
	//等服务器启动
	time.Sleep(1 * time.Second)

	config.GlobalConfigFilePath = "client.yaml"
	registerAPI, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}

	err = registerAPI.AddTable(&sdk.AddTableArgv{
		ServiceName: "A",
		Tags:        []string{"Key"},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = registerAPI.AddTable(&sdk.AddTableArgv{
		ServiceName: "B",
		Tags:        []string{"Key"},
	})
	if err != nil {
		t.Fatal(err)
	}

	go SetupClient(
		"A", "B", "client.yaml",
		"127.0.0.1", 20000,
		"keyA0", "keyB0")
	time.Sleep(50 * time.Millisecond)

	go SetupClient(
		"A", "B", "client.yaml",
		"127.0.0.1", 20001,
		"keyA1", "keyB0")
	time.Sleep(50 * time.Millisecond)

	go SetupClient(
		"A", "B", "client.yaml",
		"127.0.0.1", 20002,
		"keyA2", "keyB0")
	time.Sleep(50 * time.Millisecond)

	go SetupClient(
		"B", "A", "client.yaml",
		"127.0.0.1", 21000,
		"keyB0", "keyA0")
	time.Sleep(50 * time.Millisecond)

	go SetupClient(
		"B", "A", "client.yaml",
		"127.0.0.1", 21001,
		"keyB1", "keyA1")
	time.Sleep(50 * time.Millisecond)

	go SetupClient(
		"B", "A", "client.yaml",
		"127.0.0.1", 21002,
		"keyB2", "keyA2")
	time.Sleep(50 * time.Millisecond)

	for {
	}
}
