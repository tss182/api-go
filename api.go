package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

const (
	TypeJson      = "application/json"
	TypeUrlEncode = "application/x-www-form-urlencoded"
	TypeMultipart = "multipart/form-data"

	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodDELETE = "DELETE"
	MethodPATCH  = "PATCH"
)

type Api struct {
	Url         string
	ContentType string
	Method      string
	Body        interface{}
	Header      map[string]string
	Username    string
	Password    string
	BasicAuth   bool
	result      string
	Status      string
	req         *http.Request
}

func (api *Api) Do() error {
	if api.Url == "" || api.Method == "" || api.ContentType == "" {
		return errors.New("url,method and content type is required")
	}
	//method check
	switch api.Method {
	case MethodGET, MethodPOST, MethodPUT, MethodPATCH, MethodDELETE:
	default:
		return errors.New(api.Method + " doesn't support")
	}

	var err error

	//contentType
	switch api.ContentType {
	case TypeJson:
		err = api.jsonProccess()
	case TypeUrlEncode:
		err = api.urlEncodeProccess()
	case TypeMultipart:
		err = api.multipartProccess()
	default:
		err = errors.New(api.ContentType + " doesn't support")
	}
	if err != nil {
		return err
	}

	//set basic auth
	if api.BasicAuth {
		api.req.SetBasicAuth(api.Username, api.Password)
	}
	//set header
	for i, v := range api.Header {
		api.req.Header.Set(i, v)
	}
	client := &http.Client{}
	resp, err := client.Do(api.req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	api.result = string(body)
	api.Status = strings.Split(resp.Status, " ")[0]
	return nil

}

func (api *Api) jsonProccess() error {
	r, err := json.Marshal(api.Body)
	if err != nil {
		return err
	}
	var reader = strings.NewReader(string(r))
	api.req, err = http.NewRequest(api.Method, api.Url, reader)
	return err
}

func (api *Api) urlEncodeProccess() error {
	var err error
	param := url.Values{}
	data, ok := api.Body.(map[string]interface{})
	if !ok {
		return errors.New("body must be map[string]interface{}")
	}
	for i, dt := range data {
		switch v := dt.(type) {
		case string:
			param.Add(i, v)
		case []string:
			for _, v2 := range v {
				param.Add(i+"[]", v2)
			}
		case int, int8, int16, int32, int64:
			reflectValue := reflect.ValueOf(v)
			param.Add(i, strconv.Itoa(int(reflectValue.Int())))
		case []int:
			for _, v2 := range v {
				param.Add(i+"[]", strconv.Itoa(v2))
			}
		case uint, uint8, uint16, uint32, uint64:
			reflectValue := reflect.ValueOf(v)
			param.Add(i, strconv.Itoa(int(reflectValue.Uint())))
		case []uint:
			for _, v2 := range v {
				param.Add(i+"[]", strconv.Itoa(int(v2)))
			}
		case map[string]string:
			for i2, v2 := range v {
				param.Add(i+"["+i2+"]", v2)
			}

		default:
			return errors.New(reflect.TypeOf(v).String() + " doesn't support")
		}
	}
	payload := strings.NewReader(param.Encode())
	api.req, err = http.NewRequest(api.Method, api.Url, payload)
	return err
}

func (api *Api) multipartProccess() error {
	var err error
	payload := &bytes.Buffer{}
	param := multipart.NewWriter(payload)
	data, ok := api.Body.(map[string]interface{})
	if !ok {
		return errors.New("body must be map[string]interface{}")
	}
	for i, dt := range data {
		switch v := dt.(type) {
		case string:
			_ = param.WriteField(i, v)
		case []string:
			for _, v2 := range v {
				_ = param.WriteField(i+"[]", v2)
			}
		case int, int8, int16, int32, int64:
			reflectValue := reflect.ValueOf(v)
			_ = param.WriteField(i, strconv.Itoa(int(reflectValue.Int())))
		case []int:
			for _, v2 := range v {
				_ = param.WriteField(i+"[]", strconv.Itoa(v2))
			}
		case uint, uint8, uint16, uint32, uint64:
			reflectValue := reflect.ValueOf(v)
			_ = param.WriteField(i, strconv.Itoa(int(reflectValue.Uint())))
		case []uint:
			for _, v2 := range v {
				_ = param.WriteField(i+"[]", strconv.Itoa(int(v2)))
			}
		case map[string]string:
			for i2, v2 := range v {
				_ = param.WriteField(i+"["+i2+"]", v2)
			}
		case *multipart.FileHeader:
			file, _ := v.Open()
			write, _ := param.CreateFormFile(i, v.Filename)
			_, _ = io.Copy(write, file)
			_ = file.Close()
		case []*multipart.FileHeader:
			for _, v2 := range v {
				file, _ := v2.Open()
				f, _ := param.CreateFormFile(i+"[]", v2.Filename)
				_, _ = io.Copy(f, file)
				_ = file.Close()
			}

		default:
			return errors.New(reflect.TypeOf(v).String() + " doesn't support")
		}
	}
	err = param.Close()
	if err != nil {
		return err
	}

	api.req, err = http.NewRequest(api.Method, api.Url, payload)
	return err
}

func (api *Api) Get(data interface{}) error {
	err := json.Unmarshal([]byte(api.result), &data)
	if err != nil {
		return err
	}
	return nil
}

func (api *Api) GetXml(data interface{}) error {
	err := xml.Unmarshal([]byte(api.result), &data)
	if err != nil {
		return err
	}
	return nil
}

func (api *Api) GetRaw() string {
	return api.result
}
