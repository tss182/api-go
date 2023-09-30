package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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
	TypeText      = "text/plain"

	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodDELETE = "DELETE"
	MethodPATCH  = "PATCH"
)

type Api struct {
	Url            string
	ContentType    string
	Method         string
	header         map[string]string
	Body           interface{}
	bodyRaw        interface{}
	response       *http.Response
	Username       string
	Password       string
	BasicAuth      bool
	result         string
	Status         int
	AddCharInArray bool
	req            *http.Request
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

	if api.Method == MethodGET && api.ContentType == TypeJson {
		api.ContentType = TypeUrlEncode
	}

	//contentType
	switch api.ContentType {
	case TypeJson:
		err = api.jsonProcess()
	case TypeUrlEncode:
		err = api.urlEncodeProcess()
	case TypeMultipart:
		err = api.multipartProcess()
	case TypeText:
		err = api.textProcess()
	default:
		err = errors.New(api.ContentType + " doesn't support")
	}
	if err != nil {
		return err
	}

	//set content type
	if api.ContentType != TypeMultipart {
		api.req.Header.Set("Content-Type", api.ContentType)
	}

	//set basic auth
	if api.BasicAuth {
		api.req.SetBasicAuth(api.Username, api.Password)
	}

	//set header
	for i, v := range api.header {
		api.req.Header.Set(i, v)
	}

	client := &http.Client{}
	resp, err := client.Do(api.req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	api.result = string(body)
	api.Status, _ = strconv.Atoi(strings.Split(resp.Status, " ")[0])

	//clear
	api.response = resp
	api.bodyRaw = api.Body
	api.header = nil
	api.Body = nil

	return nil

}

func (api *Api) HeaderAdd(key, value string) {
	if api.header == nil {
		api.header = map[string]string{}
	}
	api.header[key] = value
}

func (api *Api) jsonProcess() error {
	r, err := json.Marshal(api.Body)
	if err != nil {
		return err
	}
	if api.Body == nil {
		api.req, err = http.NewRequest(api.Method, api.Url, nil)
	} else {
		reader := strings.NewReader(string(r))
		api.req, err = http.NewRequest(api.Method, api.Url, reader)
	}
	return err
}

func (api *Api) urlEncodeProcess() error {
	var err error
	param := url.Values{}
	if api.Body != nil {
		data, ok := api.Body.(map[string]interface{})
		if !ok {
			return errors.New("body must be map[string]interface{}")
		}
		for i, dt := range data {
			switch v := dt.(type) {
			case string:
				if v == "" {
					continue
				}
				param.Add(i, v)
			case []string:
				if len(v) == 0 {
					continue
				}
				for _, v2 := range v {
					if api.AddCharInArray {
						param.Add(i+"[]", v2)
					} else {
						param.Add(i, v2)
					}
				}
			case int, int8, int16, int32, int64:
				reflectValue := reflect.ValueOf(v)
				param.Add(i, strconv.Itoa(int(reflectValue.Int())))
			case []int:
				if len(v) == 0 {
					continue
				}
				for _, v2 := range v {
					if api.AddCharInArray {
						param.Add(i+"[]", strconv.Itoa(v2))
					} else {
						param.Add(i, strconv.Itoa(v2))
					}
				}
			case uint, uint8, uint16, uint32, uint64:
				reflectValue := reflect.ValueOf(v)
				param.Add(i, strconv.Itoa(int(reflectValue.Uint())))
			case []uint:
				if len(v) == 0 {
					continue
				}
				for _, v2 := range v {
					if api.AddCharInArray {
						param.Add(i+"[]", strconv.Itoa(int(v2)))
					} else {
						param.Add(i, strconv.Itoa(int(v2)))
					}
				}
			case float32, float64:
				reflectValue := reflect.ValueOf(v)
				param.Add(i, fmt.Sprintf("%f", reflectValue.Float()))
			case map[string]string:
				if len(v) == 0 {
					continue
				}
				for i2, v2 := range v {
					param.Add(i+"["+i2+"]", v2)
				}

			default:
				return errors.New(reflect.TypeOf(v).String() + " doesn't support")
			}
		}
	}
	if api.Method == MethodGET {
		char := "?"
		if strings.Contains(api.Url, "?") {
			char = "&"
		}
		api.Url += char + param.Encode()
		api.req, err = http.NewRequest(api.Method, api.Url, nil)
	} else {
		payload := strings.NewReader(param.Encode())
		api.req, err = http.NewRequest(api.Method, api.Url, payload)
	}

	return err
}

func (api *Api) multipartProcess() error {
	var err error
	payload := &bytes.Buffer{}
	param := multipart.NewWriter(payload)
	if api.Body != nil {
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
					if api.AddCharInArray {
						_ = param.WriteField(i+"[]", v2)
					} else {
						_ = param.WriteField(i, v2)
					}
				}
			case int, int8, int16, int32, int64:
				reflectValue := reflect.ValueOf(v)
				_ = param.WriteField(i, strconv.Itoa(int(reflectValue.Int())))
			case []int:
				for _, v2 := range v {
					if api.AddCharInArray {
						_ = param.WriteField(i+"[]", strconv.Itoa(v2))
					} else {
						_ = param.WriteField(i, strconv.Itoa(v2))
					}
				}
			case uint, uint8, uint16, uint32, uint64:
				reflectValue := reflect.ValueOf(v)
				_ = param.WriteField(i, strconv.Itoa(int(reflectValue.Uint())))
			case []uint:
				for _, v2 := range v {
					if api.AddCharInArray {
						_ = param.WriteField(i+"[]", strconv.Itoa(int(v2)))
					} else {
						_ = param.WriteField(i, strconv.Itoa(int(v2)))
					}
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
					if api.AddCharInArray {
						f, _ := param.CreateFormFile(i+"[]", v2.Filename)
						_, _ = io.Copy(f, file)
						_ = file.Close()
					} else {
						f, _ := param.CreateFormFile(i, v2.Filename)
						_, _ = io.Copy(f, file)
						_ = file.Close()
					}
				}

			default:
				return errors.New(reflect.TypeOf(v).String() + " doesn't support")
			}
		}
	}
	err = param.Close()
	if err != nil {
		return err
	}

	api.req, err = http.NewRequest(api.Method, api.Url, payload)
	api.req.Header.Set("Content-Type", param.FormDataContentType())

	return err
}

func (api *Api) textProcess() error {
	var err error
	if api.Body == nil {
		api.req, err = http.NewRequest(api.Method, api.Url, nil)
	} else {
		reader := strings.NewReader(api.Body.(string))
		api.req, err = http.NewRequest(api.Method, api.Url, reader)
	}
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

func (api *Api) GetRequest() *http.Request {
	return api.req
}

func (api *Api) GetHeader() http.Header {
	return api.req.Header
}

func (api *Api) GetBody() interface{} {
	return api.bodyRaw
}

func (api *Api) GetResponse() *http.Response {
	return api.response
}
