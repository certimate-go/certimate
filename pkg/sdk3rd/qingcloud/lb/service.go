// An extension SDK client for QingCloud LB service.
// Based on github.com/yunify/qingcloud-sdk-go.
package lb

import (
	"fmt"
	"time"

	"github.com/yunify/qingcloud-sdk-go/config"
	"github.com/yunify/qingcloud-sdk-go/request"
	"github.com/yunify/qingcloud-sdk-go/request/data"
	"github.com/yunify/qingcloud-sdk-go/request/errors"
	"github.com/yunify/qingcloud-sdk-go/service"
)

var (
	_ fmt.State
	_ time.Time
)

type LoadBalancerService struct {
	Config     *config.Config
	Properties *LoadBalancerServiceProperties
}

type LoadBalancerServiceProperties struct {
	Zone *string `json:"zone" name:"zone"`
}

func NewService(config *config.Config) (*LoadBalancerService, error) {
	properties := &LoadBalancerServiceProperties{
		Zone: &config.Zone,
	}

	return &LoadBalancerService{Config: config, Properties: properties}, nil
}

func (s *LoadBalancerService) AssociateServerCertsToLBListener(i *AssociateServerCertsToLBListenerInput) (*AssociateServerCertsToLBListenerOutput, error) {
	if i == nil {
		i = &AssociateServerCertsToLBListenerInput{}
	}
	o := &data.Operation{
		Config:        s.Config,
		Properties:    s.Properties,
		APIName:       "AssociateServerCertsToLBListener",
		RequestMethod: "POST",
	}

	x := &AssociateServerCertsToLBListenerOutput{}
	r, err := request.New(o, i, x)
	if err != nil {
		return nil, err
	}

	err = r.Send()
	if err != nil {
		return nil, err
	}

	return x, err
}

type AssociateServerCertsToLBListenerInput struct {
	LoadBalancerListener *string   `json:"loadbalancer_listener" name:"loadbalancer_listener" location:"params"`
	ServerCertificates   []*string `json:"server_certificates" name:"server_certificates" location:"params"`
}

func (v *AssociateServerCertsToLBListenerInput) Validate() error {
	if v.LoadBalancerListener == nil {
		return errors.ParameterRequiredError{
			ParameterName: "LoadBalancerListener",
			ParentName:    "AssociateServerCertsToLBListenerInput",
		}
	}

	if v.ServerCertificates == nil {
		return errors.ParameterRequiredError{
			ParameterName: "ServerCertificates",
			ParentName:    "AssociateServerCertsToLBListenerInput",
		}
	}

	return nil
}

type AssociateServerCertsToLBListenerOutput struct {
	Message *string `json:"message" name:"message"`
	Action  *string `json:"action" name:"action" location:"elements"`
	RetCode *int    `json:"ret_code" name:"ret_code" location:"elements"`
}

func (s *LoadBalancerService) CreateServerCertificate(i *CreateServerCertificateInput) (*CreateServerCertificateOutput, error) {
	if i == nil {
		i = &CreateServerCertificateInput{}
	}
	o := &data.Operation{
		Config:        s.Config,
		Properties:    s.Properties,
		APIName:       "CreateServerCertificate",
		RequestMethod: "POST",
	}

	x := &CreateServerCertificateOutput{}
	r, err := request.New(o, i, x)
	if err != nil {
		return nil, err
	}

	err = r.Send()
	if err != nil {
		return nil, err
	}

	return x, err
}

type CreateServerCertificateInput = service.CreateServerCertificateInput

type CreateServerCertificateOutput = service.CreateServerCertificateOutput

func (s *LoadBalancerService) DescribeLoadBalancerListeners(i *DescribeLoadBalancerListenersInput) (*DescribeLoadBalancerListenersOutput, error) {
	if i == nil {
		i = &DescribeLoadBalancerListenersInput{}
	}
	o := &data.Operation{
		Config:        s.Config,
		Properties:    s.Properties,
		APIName:       "DescribeLoadBalancerListeners",
		RequestMethod: "GET",
	}

	x := &DescribeLoadBalancerListenersOutput{}
	r, err := request.New(o, i, x)
	if err != nil {
		return nil, err
	}

	err = r.Send()
	if err != nil {
		return nil, err
	}

	return x, err
}

type DescribeLoadBalancerListenersInput = service.DescribeLoadBalancerListenersInput

type DescribeLoadBalancerListenersOutput = service.DescribeLoadBalancerListenersOutput

func (s *LoadBalancerService) DescribeServerCertificates(i *DescribeServerCertificatesInput) (*DescribeServerCertificatesOutput, error) {
	if i == nil {
		i = &DescribeServerCertificatesInput{}
	}
	o := &data.Operation{
		Config:        s.Config,
		Properties:    s.Properties,
		APIName:       "DescribeServerCertificates",
		RequestMethod: "GET",
	}

	x := &DescribeServerCertificatesOutput{}
	r, err := request.New(o, i, x)
	if err != nil {
		return nil, err
	}

	err = r.Send()
	if err != nil {
		return nil, err
	}

	return x, err
}

type DescribeServerCertificatesInput = service.DescribeServerCertificatesInput

type DescribeServerCertificatesOutput = service.DescribeServerCertificatesOutput
