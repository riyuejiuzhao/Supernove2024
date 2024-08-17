package test_show_router

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/metrics"
	"Supernove2024/svr"
	"Supernove2024/util"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"log"
	"net/http"
	"testing"
	"time"
)

const jAddress = "127.0.0.1:6831"

func SetupClient(targetService string) {

	//链路追踪
	jCfg := jaegerConfig.Configuration{
		ServiceName: "Client",
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jAddress,
		},
	}
	tracer, closer, err := jCfg.NewTracer()
	if err != nil {
		log.Fatalln(err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)
	//====================================================
	httpClient := &http.Client{}
	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		log.Fatalln(err)
	}
	var router *sdk.ProcessRouterResult = nil
	for {
		func() {
			router, err = discoveryAPI.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:          util.KVRouterType,
				SrcInstanceName: "",
				DstService:      targetService,
				Key: map[string]string{
					"Key0": "Client",
					"Key1": "Client",
					"Key2": "Client"},
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
			fmt.Println("client send")
		}()
	}
}

func SetupServer(
	service string,
	configFile string,
	host string,
	port int32,
	name string,
	weight int32,
) {
	mt, err := metrics.Instance()
	if err != nil {
		util.Error("%v", err)
		return
	}
	//链路追踪
	jCfg := jaegerConfig.Configuration{
		ServiceName: name,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jAddress,
		},
	}
	tracer, closer, err := jCfg.NewTracer()
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
		Name:        name,
		Host:        host,
		Port:        port,
		Weight:      &weight,
	})
	if err != nil {
		log.Fatalln(err)
	} else {
		util.Info("Register %s successful", name)
	}

	defer registerAPI.Deregister(&sdk.DeregisterArgv{
		ServiceName: service,
		InstanceID:  reply.InstanceID,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		defer func() {
			mt.MetricsUpload(begin, prometheus.Labels{"Method": name}, nil)
		}()
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		span := tracer.StartSpan("hello-world-reply", opentracing.ChildOf(spanCtx))
		defer span.Finish()

		fmt.Printf("%s received\n", name)
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
	util.ClearEtcd("127.0.0.1:2301", t)
	util.ClearEtcd("127.0.0.1:2311", t)

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

	go SetupServer(
		"HTTP", "client.yaml",
		"127.0.0.1", 20000,
		"A0", 30)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(
		"HTTP", "client.yaml",
		"127.0.0.1", 20001,
		"A1", 20)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(
		"HTTP", "client.yaml",
		"127.0.0.1", 20002,
		"A2", 10)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(
		"HTTP", "client.yaml",
		"127.0.0.1", 21000,
		"B0", 1000)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(
		"HTTP", "client.yaml",
		"127.0.0.1", 21001,
		"C0", 50)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(
		"HTTP", "client.yaml",
		"127.0.0.1", 21002,
		"C1", 50)
	time.Sleep(50 * time.Millisecond)

	err = registerAPI.AddTable(&sdk.AddTableArgv{
		ServiceName: "HTTP",
		Tags:        []string{"Key0", "Key1", "Key2"},
	})
	if err != nil {
		t.Fatal(err)
	}

	timeout := int64(10000000)
	_, err = registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"Key0": "Client"},
		DstServiceName:  "HTTP",
		DstInstanceName: []string{"A0", "A1", "A2"},
		Timeout:         &timeout,
		NextRouterType:  util.WeightedRouterType,
	})
	if err != nil {
		log.Fatalln(err)
	}

	go SetupClient("HTTP")

	time.Sleep(60 * time.Second)
	_, err = registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic: map[string]string{
			"Key0": "Client",
			"Key1": "Client",
		},
		DstServiceName:  "HTTP",
		DstInstanceName: []string{"B0"},
		Timeout:         &timeout,
		NextRouterType:  util.WeightedRouterType,
	})
	if err != nil {
		log.Fatalln(err)
	}

	time.Sleep(60 * time.Second)
	_, err = registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic: map[string]string{
			"Key0": "Client",
			"Key1": "Client",
			"Key2": "Client",
		},
		DstServiceName:  "HTTP",
		DstInstanceName: []string{"C0", "C1"},
		Timeout:         &timeout,
		NextRouterType:  util.WeightedRouterType,
	})
	if err != nil {
		log.Fatalln(err)
	}
	time.Sleep(60 * time.Second)
}
