package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)
// Datas This is response struct from create
type Datas struct {
	Data struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	} `json:"data"`
}

// ClientApiResponse This is get single api response
type ClientApiResponse struct {
	Data ClientApiData `json:"data"`
}

type ClientApiData struct {
	Data ClientApiResp `json:"data"`
}

type ClientApiResp struct {
	Response map[string]interface{} `json:"response"`
}

type Response struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// NewRequestBody's Data (map) field will be in this structure
//.   fields
// objects_ids []string
// table_slug string
// object_data map[string]interface
// method string
// app_id string

// but all field will be an interface, you must do type assertion

type HttpRequest struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Params  url.Values  `json:"params"`
	Body    []byte      `json:"body"`
}

type AuthData struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type NewRequestBody struct {
	RequestData HttpRequest            `json:"request_data"`
	Auth        AuthData               `json:"auth"`
	Data        map[string]interface{} `json:"data"`
}
type Request struct {
	Data map[string]interface{} `json:"data"`
}

func DoRequest(url string, method string, body interface{}, appId string) ([]byte, error) {
	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("authorization", "API-KEY")
	request.Header.Add("X-API-KEY", appId)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respByte, nil
}

// Handle a serverless request
func Handle(req []byte) string {
	var response Response
	var request NewRequestBody
	const urlConst = "https://api.admin.u-code.io"

	err := json.Unmarshal(req, &request)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling request"}
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}
	if request.Data["app_id"] == nil {
		response.Data = map[string]interface{}{"message": "App id required"}
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}
	appId := request.Data["app_id"].(string)

	// you may change table slug  it's related your business logic
	var tableSlug = "notifications"

	var (
		getListUrl string = fmt.Sprintf("%v/v1/object/get-list/%v?project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", urlConst, tableSlug)
		updateUrl  string = fmt.Sprintf("%v/v1/object/%v?project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", urlConst, tableSlug)
	)
	getListObjectRequest := Request{
		Data: map[string]interface{}{
			"client_id": request.Data["client_id"].(string),
			"is_read":   false,
			"time_take": map[string]interface{}{
				"$lte": time.Now(),
			},
		},
	}
	res, err, response := GetListObject(getListUrl, appId, getListObjectRequest)
	if err != nil {
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}

	for _, v := range res.Data.Data.Response {
		updateRequest := Request{
			Data: map[string]interface{}{
				"guid":    v["guid"].(string),
				"is_read": true,
			},
		}
		err, _ := UpdateObject(updateUrl, appId, updateRequest)
		if err != nil {
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
	}

	response.Data = map[string]interface{}{ }
	response.Status = "done" //if all will be ok else "error"
	responseByte, _ := json.Marshal(response)

	return string(responseByte)
}
