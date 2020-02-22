package helper

import (
	"context"
	"fmt"
	"log"
	"os"

	servicebus "github.com/Azure/azure-service-bus-go"
)

type ServiceBusResponse struct {
	AllCountDetails map[string]*servicebus.CountDetails
}

// ServiceBusHelper handles all things sendgrid...
type ServiceBusHelper struct {
	ConnectionString string // connection string to servicebus
	namespace        *servicebus.Namespace
	qMgr             *servicebus.QueueManager
	tMgr             *servicebus.TopicManager
}

// NewServiceBusHelper creates a new instance of ServiceBusHelper
func NewServiceBusHelper() *ServiceBusHelper {
	sb := ServiceBusHelper{}
	sb.ConnectionString = os.Getenv("SERVICEBUS_CONNECTIONSTRING")

	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(sb.ConnectionString))
	if err != nil {
		// handle error
		log.Fatal("BOOM")
	}
	sb.namespace = ns
	qMgr := ns.NewQueueManager()
	sb.qMgr = qMgr

	tMgr := ns.NewTopicManager()
	sb.tMgr = tMgr

	return &sb
}

// CheckQueue checks queue for messages
func (sb *ServiceBusHelper) CheckQueue(name string) (ServiceBusResponse, error) {

	l1, err := sb.qMgr.List(context.Background())
	if err != nil {
		log.Fatal("BOOM on queue")
	}

	resp := ServiceBusResponse{}
	resp.AllCountDetails = make(map[string]*servicebus.CountDetails)

	if name != "" {
		for _, i := range l1 {
			if i.Name == name {
				resp.AllCountDetails[i.Name] = i.CountDetails
			}
		}
	} else {
		for _, i := range l1 {
			resp.AllCountDetails[i.Name] = i.CountDetails
		}
	}

	fmt.Printf("RES SIZE %d\n", len(resp.AllCountDetails))
	return resp, nil
}

// CheckTopic checks queue for messages
func (sb *ServiceBusHelper) CheckTopic(name string) (ServiceBusResponse, error) {

	l1, err := sb.tMgr.List(context.Background())
	if err != nil {
		log.Fatal("BOOM on topic")
	}

	resp := ServiceBusResponse{}
	resp.AllCountDetails = make(map[string]*servicebus.CountDetails)

	if name != "" {
		for _, i := range l1 {
			if i.Name == name {
				resp.AllCountDetails[i.Name] = i.CountDetails
			}
		}
	} else {
		for _, i := range l1 {
			resp.AllCountDetails[i.Name] = i.CountDetails
		}
	}

	return resp, nil
}
