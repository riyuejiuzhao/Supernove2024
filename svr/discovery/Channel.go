package discovery

/*
func (s *Server) RecvServiceChan(service string, serviceChan <-chan *redis.Message) {
	for msg := range serviceChan {
		log := &pb.InstanceLog{}
		err := proto.Unmarshal([]byte(msg.Payload), log)
		if err != nil {
			util.Error("chan %s err: %v", msg.Channel, err)
			continue
		}
		info := s.ServiceBuffer.GetServiceInfo(service)
		if info.Revision == log.Revision {

		}
	}
}

func (s *Server) RecvAllServiceChannel(allServiceChan <-chan *redis.Message) {
	for msg := range allServiceChan {
		newService := msg.Payload
		newServicePul := s.Rdb.Subscribe(svrutil.ServiceChan(newService))
		_, err := newServicePul.Receive()
		if err != nil {
			util.Error("subscribe %s err: %v", newService, err)
			continue
		}
		newChannel := newServicePul.Channel()
		go s.RecvServiceChan(newService, newChannel)
	}
}

// Channel 订阅信息获取，增量获取，降低流量压力
func (s *Server) Channel() {
	allServicePul := s.Rdb.Subscribe(svrutil.AllServiceChannel)
	_, err := allServicePul.Receive()
	if err != nil {
		log.Fatalf("AllServiceChannel failed %v", err)
	}

	allServiceChan := allServicePul.Channel()
	go s.RecvAllServiceChannel(allServiceChan)
}

func (s *Server) MetricsUpload() {
	for {
		const service = "Discovery"
		//GetInstances
		label := prometheus.Labels{
			svrutil.ServiceTag: service,
			svrutil.MethodTag:  "GetInstance",
		}
		s.MethodMetricsUpload(label,
			&s.GetInstanceRequest,
			&s.GetInstanceReply,
			nil,
			&s.GetInstancesRecvRedis)

		//GetRouters
		label = prometheus.Labels{
			svrutil.ServiceTag: service,
			svrutil.MethodTag:  "GetRouters",
		}
		s.MethodMetricsUpload(label,
			&s.GetRoutersRequest,
			&s.GetRoutersReply,
			nil,
			&s.GetRoutersRecvRedis)

		time.Sleep(1 * time.Second)
	}
}
*/
