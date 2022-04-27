package mocks

import "net/http"

type MockResponseData struct {
	Response *http.Response
	Error    error
}

type MockHTTPRequester struct {
	requestData  []*http.Request
	responseData []MockResponseData
}

func (r *MockHTTPRequester) Do(req *http.Request) (*http.Response, error) {
	r.requestData = append(r.requestData, req)

	if len(r.responseData) == 0 {
		return nil, http.ErrBodyReadAfterClose
	}

	data := r.responseData[0]
	r.responseData = r.responseData[1:]
	return data.Response, nil
}

func (r *MockHTTPRequester) SetResponse(data MockResponseData) {
	r.responseData = append(r.responseData, data)
}

func (r *MockHTTPRequester) GetRequest() *http.Request {
	if len(r.requestData) == 0 {
		return nil
	}

	data := r.requestData[0]
	r.requestData = r.requestData[1:]
	return data
}
