// +build !integration

package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strings"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
	ServiceReconfigure
	ConsulAddress      string
	ReconfigureBaseUrl string
	RemoveBaseUrl      string
	ReconfigureUrl     string
	RemoveUrl          string
	ResponseWriter     *ResponseWriterMock
	RequestReconfigure *http.Request
	RequestRemove      *http.Request
}

func (s *ServerTestSuite) SetupTest() {
	s.ConsulAddress = "http://1.2.3.4:1234"
	s.ServiceName = "myService"
	s.ServiceColor = "pink"
	s.ServiceDomain = "my-domain.com"
	s.ServicePath = []string{"/path/to/my/service/api", "/path/to/my/other/service/api"}
	s.ReconfigureBaseUrl = "/v1/docker-flow-proxy/reconfigure"
	s.RemoveBaseUrl = "/v1/docker-flow-proxy/remove"
	s.ReconfigureUrl = fmt.Sprintf(
		"%s?serviceName=%s&serviceColor=%s&servicePath=%s&serviceDomain=%s",
		s.ReconfigureBaseUrl,
		s.ServiceName,
		s.ServiceColor,
		strings.Join(s.ServicePath, ","),
		s.ServiceDomain,
	)
	s.RemoveUrl = fmt.Sprintf(
		"%s?serviceName=%s",
		s.RemoveBaseUrl,
		s.ServiceName,
	)
	s.ResponseWriter = getResponseWriterMock()
	s.RequestReconfigure, _ = http.NewRequest("GET", s.ReconfigureUrl, nil)
	s.RequestRemove, _ = http.NewRequest("GET", s.RemoveUrl, nil)
	httpListenAndServe = func(addr string, handler http.Handler) error {
		return nil
	}
	server = Server{
		BaseReconfigure: BaseReconfigure{
			ConsulAddress: s.ConsulAddress,
		},
	}
	NewReconfigure = func(baseData BaseReconfigure, serviceData ServiceReconfigure) Reconfigurable {
		return getReconfigureMock("")
	}
	logPrintf = func(format string, v ...interface{}) {}
}

// Execute

func (s *ServerTestSuite) Test_Execute_InvokesHTTPListenAndServe() {
	server := Server{
		IP:   "myIp",
		Port: "1234",
	}
	var actual string
	expected := fmt.Sprintf("%s:%s", server.IP, server.Port)
	httpListenAndServe = func(addr string, handler http.Handler) error {
		actual = addr
		return nil
	}
	server.Execute([]string{})
	s.Equal(expected, actual)
}

func (s *ServerTestSuite) Test_Execute_ReturnsError_WhenHTTPListenAndServeFails() {
	orig := httpListenAndServe
	defer func() {
		httpListenAndServe = orig
	}()
	httpListenAndServe = func(addr string, handler http.Handler) error {
		return fmt.Errorf("This is an error")
	}

	actual := server.Execute([]string{})

	s.Error(actual)
}

func (s *ServerTestSuite) Test_Execute_InvokesRunExecute() {
	orig := NewRun
	defer func() {
		NewRun = orig
	}()
	mockObj := getRunMock("")
	NewRun = func() Executable {
		return mockObj
	}

	server.Execute([]string{})

	mockObj.AssertCalled(s.T(), "Execute", []string{})
}

func (s *ServerTestSuite) Test_Execute_InvokesReloadAllServices() {
	mockObj := getReconfigureMock("")
	NewReconfigure = func(baseData BaseReconfigure, serviceData ServiceReconfigure) Reconfigurable {
		return mockObj
	}

	server.Execute([]string{})

	mockObj.AssertCalled(s.T(), "ReloadAllServices", s.ConsulAddress)
}

func (s *ServerTestSuite) Test_Execute_ReturnsErrro_WhenReloadAllServicesFails() {
	mockObj := getReconfigureMock("ReloadAllServices")
	mockObj.On("ReloadAllServices", mock.Anything).Return(fmt.Errorf("This is an error"))
	NewReconfigure = func(baseData BaseReconfigure, serviceData ServiceReconfigure) Reconfigurable {
		return mockObj
	}

	actual := server.Execute([]string{})

	s.Error(actual)
}

// ServeHTTP

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsStatus404WhenURLIsUnknown() {
	req, _ := http.NewRequest("GET", "/this/url/does/not/exist", nil)

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.ResponseWriter.AssertCalled(s.T(), "WriteHeader", 404)
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsStatus200WhenUrlIsTest() {
	for ver := 1; ver <= 2; ver++ {
		rw := getResponseWriterMock()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v%d/test", ver), nil)

		Server{}.ServeHTTP(rw, req)

		rw.AssertCalled(s.T(), "WriteHeader", 200)
	}
}

// ServeHTTP > Reconfigure

func (s *ServerTestSuite) Test_ServeHTTP_SetsContentTypeToJSON_WhenUrlIsReconfigure() {
	var actual string
	httpWriterSetContentType = func(w http.ResponseWriter, value string) {
		actual = value
	}
	req, _ := http.NewRequest("GET", s.ReconfigureUrl, nil)

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.Equal("application/json", actual)
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsJSON_WhenUrlIsReconfigure() {
	expected, _ := json.Marshal(Response{
		Status:        "OK",
		ServiceName:   s.ServiceName,
		ServiceColor:  s.ServiceColor,
		ServicePath:   s.ServicePath,
		ServiceDomain: s.ServiceDomain,
		PathType:      s.PathType,
	})

	Server{}.ServeHTTP(s.ResponseWriter, s.RequestReconfigure)

	s.ResponseWriter.AssertCalled(s.T(), "Write", []byte(expected))
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsJsonWithPathType_WhenPresent() {
	pathType := "path_reg"
	req, _ := http.NewRequest("GET", s.ReconfigureUrl+"&pathType="+pathType, nil)
	expected, _ := json.Marshal(Response{
		Status:        "OK",
		ServiceName:   s.ServiceName,
		ServiceColor:  s.ServiceColor,
		ServicePath:   s.ServicePath,
		ServiceDomain: s.ServiceDomain,
		PathType:      pathType,
	})

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.ResponseWriter.AssertCalled(s.T(), "Write", []byte(expected))
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsJsonWithSkipCheck_WhenPresent() {
	req, _ := http.NewRequest("GET", s.ReconfigureUrl+"&skipCheck=true", nil)
	expected, _ := json.Marshal(Response{
		Status:        "OK",
		ServiceName:   s.ServiceName,
		ServiceColor:  s.ServiceColor,
		ServicePath:   s.ServicePath,
		ServiceDomain: s.ServiceDomain,
		PathType:      s.PathType,
		SkipCheck:     true,
	})

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.ResponseWriter.AssertCalled(s.T(), "Write", []byte(expected))
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsStatus400_WhenUrlIsReconfigureAndServiceNameQueryIsNotPresent() {
	req, _ := http.NewRequest("GET", s.ReconfigureBaseUrl, nil)

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.ResponseWriter.AssertCalled(s.T(), "WriteHeader", 400)
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsStatus400_WhenServicePathQueryIsNotPresent() {
	url := fmt.Sprintf("%s?serviceName=%s", s.ReconfigureBaseUrl, s.ServiceName[0])
	req, _ := http.NewRequest("GET", url, nil)

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.ResponseWriter.AssertCalled(s.T(), "WriteHeader", 400)
}

func (s *ServerTestSuite) Test_ServeHTTP_InvokesReconfigureExecute() {
	mockObj := getReconfigureMock("")
	var actualBase BaseReconfigure
	expectedBase := BaseReconfigure{
		ConsulAddress: s.ConsulAddress,
	}
	var actualService ServiceReconfigure
	NewReconfigure = func(baseData BaseReconfigure, serviceData ServiceReconfigure) Reconfigurable {
		actualBase = baseData
		actualService = serviceData
		return mockObj
	}
	server := Server{BaseReconfigure: expectedBase}

	server.ServeHTTP(s.ResponseWriter, s.RequestReconfigure)

	s.Equal(expectedBase, actualBase)
	s.Equal(s.ServiceReconfigure, actualService)
	mockObj.AssertCalled(s.T(), "Execute", []string{})
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsStatus500_WhenReconfigureExecuteFails() {
	mockObj := getReconfigureMock("Execute")
	mockObj.On("Execute", []string{}).Return(fmt.Errorf("This is an error"))
	NewReconfigure = func(baseData BaseReconfigure, serviceData ServiceReconfigure) Reconfigurable {
		return mockObj
	}

	Server{}.ServeHTTP(s.ResponseWriter, s.RequestReconfigure)

	s.ResponseWriter.AssertCalled(s.T(), "WriteHeader", 500)
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsJson_WhenConsulTemplatePathIsPresent() {
	path := "/path/to/consul/template"
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s?serviceName=%s&consulTemplatePath=%s", s.ReconfigureBaseUrl, s.ServiceName, path),
		nil,
	)
	expected, _ := json.Marshal(Response{
		Status:             "OK",
		ServiceName:        s.ServiceName,
		ConsulTemplatePath: path,
		PathType:           s.PathType,
	})

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.ResponseWriter.AssertCalled(s.T(), "Write", []byte(expected))
}

func (s *ServerTestSuite) Test_ServeHTTP_InvokesReconfigureExecute_WhenConsulTemplatePathIsPresent() {
	path := "/path/to/consul/template"
	mockObj := getReconfigureMock("")
	var actualBase BaseReconfigure
	expectedBase := BaseReconfigure{
		ConsulAddress: s.ConsulAddress,
	}
	expectedService := ServiceReconfigure{
		ServiceName:        s.ServiceName,
		ConsulTemplatePath: path,
		PathType:           s.PathType,
	}
	var actualService ServiceReconfigure
	NewReconfigure = func(baseData BaseReconfigure, serviceData ServiceReconfigure) Reconfigurable {
		actualBase = baseData
		actualService = serviceData
		return mockObj
	}
	server := Server{BaseReconfigure: expectedBase}
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s?serviceName=%s&consulTemplatePath=%s", s.ReconfigureBaseUrl, s.ServiceName, path),
		nil,
	)

	server.ServeHTTP(s.ResponseWriter, req)

	s.Equal(expectedBase, actualBase)
	s.Equal(expectedService, actualService)
	mockObj.AssertCalled(s.T(), "Execute", []string{})
}

// ServeHTTP > Remove

func (s *ServerTestSuite) Test_ServeHTTP_SetsContentTypeToJSON_WhenUrlIsRemove() {
	var actual string
	httpWriterSetContentType = func(w http.ResponseWriter, value string) {
		actual = value
	}
	req, _ := http.NewRequest("GET", s.RemoveUrl, nil)

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.Equal("application/json", actual)
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsJSON_WhenUrlIsRemove() {
	expected, _ := json.Marshal(Response{
		Status:      "OK",
		ServiceName: s.ServiceName,
	})

	Server{}.ServeHTTP(s.ResponseWriter, s.RequestRemove)

	s.ResponseWriter.AssertCalled(s.T(), "Write", []byte(expected))
}

func (s *ServerTestSuite) Test_ServeHTTP_ReturnsStatus400_WhenUrlIsRemoveAndServiceNameQueryIsNotPresent() {
	req, _ := http.NewRequest("GET", s.RemoveBaseUrl, nil)

	Server{}.ServeHTTP(s.ResponseWriter, req)

	s.ResponseWriter.AssertCalled(s.T(), "WriteHeader", 400)
}

func (s *ServerTestSuite) Test_ServeHTTP_InvokesRemoveExecute() {
	mockObj := getRemoveMock("")
	var actual Remove
	expected := Remove{
		ServiceName: s.ServiceName,
	}
	NewRemove = func(serviceName, configsPath, templatesPath string) Removable {
		actual = Remove{
			ServiceName:   serviceName,
			TemplatesPath: templatesPath,
			ConfigsPath:   configsPath,
		}
		return mockObj
	}

	server.ServeHTTP(s.ResponseWriter, s.RequestRemove)

	s.Equal(expected, actual)
	mockObj.AssertCalled(s.T(), "Execute", []string{})
}

// Suite

func TestServerTestSuite(t *testing.T) {
	logPrintf = func(format string, v ...interface{}) {}
	suite.Run(t, new(ServerTestSuite))
}

// Mock

type ServerMock struct {
	mock.Mock
}

func (m *ServerMock) Execute(args []string) error {
	params := m.Called(args)
	return params.Error(0)
}

func (m *ServerMock) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.Called(w, req)
}

func getServerMock() *ServerMock {
	mockObj := new(ServerMock)
	mockObj.On("Execute", mock.Anything).Return(nil)
	mockObj.On("ServeHTTP", mock.Anything, mock.Anything)
	return mockObj
}

type ResponseWriterMock struct {
	mock.Mock
}

func (m *ResponseWriterMock) Header() http.Header {
	m.Called()
	return make(map[string][]string)
}

func (m *ResponseWriterMock) Write(data []byte) (int, error) {
	params := m.Called(data)
	return params.Int(0), params.Error(1)
}

func (m *ResponseWriterMock) WriteHeader(header int) {
	m.Called(header)
}

func getResponseWriterMock() *ResponseWriterMock {
	mockObj := new(ResponseWriterMock)
	mockObj.On("Header").Return(nil)
	mockObj.On("Write", mock.Anything).Return(0, nil)
	mockObj.On("WriteHeader", mock.Anything)
	return mockObj
}
