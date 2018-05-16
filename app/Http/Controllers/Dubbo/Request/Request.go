// Package gorequest inspired by Nodejs SuperChannel provides easy-way to write http client
package Request

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"mime/multipart"

	"net/textproto"

	"fmt"

	"path/filepath"

	"github.com/moul/http2curl"
	"golang.org/x/net/publicsuffix"
)

type Request *http.Request
type Response *http.Response

// HTTP methods we support
const (
	POST    = "POST"
	GET     = "GET"
	HEAD    = "HEAD"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
)

// A SuperChannel is a object storing all request data for client.
type SuperHttpClient struct {
	Url               string
	Method            string
	Header            map[string]string
	TargetType        string
	ForceType         string
	Data              map[string]interface{}
	SliceData         []interface{}
	FormData          url.Values
	QueryData         url.Values
	FileData          []File
	BounceToRawString bool
	RawString         string
	Client            *http.Client
	Transport         *http.Transport
	Cookies           []*http.Cookie
	Errors            []error
	BasicAuth         struct{ Username, Password string }
	Debug             bool
	CurlCommand       bool
	logger            *log.Logger
	Retryable         struct {
		RetryableStatus []int
		RetryerTime     time.Duration
		RetryerCount    int
		Attempt         int
		Enable          bool
	}
}

type SuperChannel struct {
	http   *SuperHttpClient

	Url               string
	Method            string
	Header            map[string]string
	TargetType        string
	ForceType         string
	Data              map[string]interface{}
	SliceData         []interface{}
	FormData          url.Values
	QueryData         url.Values
	FileData          []File
	BounceToRawString bool
	RawString         string
	Cookies           []*http.Cookie
	Errors            []error

 }

var DisableTransportSwap = false

// Used to create a new SuperChannel object.
func NewClient() *SuperHttpClient {
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)

	debug := os.Getenv("GOREQUEST_DEBUG") == "1"

	http := &SuperHttpClient{
		TargetType:        "json",
		Data:              make(map[string]interface{}),
		Header:            make(map[string]string),
		RawString:         "",
		SliceData:         []interface{}{},
		FormData:          url.Values{},
		QueryData:         url.Values{},
		FileData:          make([]File, 0),
		BounceToRawString: false,
		Client:            &http.Client{Jar: jar},
		Transport:         &http.Transport{},
		Cookies:           make([]*http.Cookie, 0),
		Errors:            nil,
		BasicAuth:         struct{ Username, Password string }{},
		Debug:             debug,
		CurlCommand:       false,
		logger:            log.New(os.Stderr, "[gorequest]", log.LstdFlags),
	}
	// disable keep alives by default, see this issue https://github.com/parnurzeal/gorequest/issues/75
	http.Transport.DisableKeepAlives = true
	return http
}

// Enable the debug mode which logs request/response detail
func (ch *SuperChannel) SetDebug(enable bool) *SuperChannel {
	ch.http.Debug = enable
	return ch
}

// Enable the curlcommand mode which display a CURL command line
func (ch *SuperChannel) SetCurlCommand(enable bool) *SuperChannel {
	ch.http.CurlCommand = enable
	return ch
}

func (ch *SuperChannel) SetLogger(logger *log.Logger) *SuperChannel {
	ch.http.logger = logger
	return ch
}

// Clear SuperChannel data for another new request.
func (sHttp *SuperHttpClient) NewChannel() *SuperChannel {
	ch := &SuperChannel{
		http:sHttp,
		Url : "",
		Method : "",
		Header : make(map[string]string),
		Data : make(map[string]interface{}),
		SliceData : []interface{}{},
		FormData : url.Values{},
		QueryData : url.Values{},
		FileData : make([]File, 0),
		BounceToRawString : false,
		RawString : "",
		ForceType : "",
		TargetType : "json",
		Cookies : make([]*http.Cookie, 0),
		Errors : nil,
	}
	return ch
}

// Just a wrapper to initialize SuperChannel instance by method string
func (sHttp *SuperHttpClient) CustomMethod(method, targetUrl string) *SuperChannel {
	switch method {
	case POST:
		return sHttp.Post(targetUrl)
	case GET:
		return sHttp.Get(targetUrl)
	case HEAD:
		return sHttp.Head(targetUrl)
	case PUT:
		return sHttp.Put(targetUrl)
	case DELETE:
		return sHttp.Delete(targetUrl)
	case PATCH:
		return sHttp.Patch(targetUrl)
	case OPTIONS:
		return sHttp.Options(targetUrl)
	default:
		ch := sHttp.NewChannel()
		ch.Method = method
		ch.Url = targetUrl
		ch.Errors = nil
		return ch
	}
}

func (sHttp *SuperHttpClient) Get(targetUrl string) *SuperChannel {
	ch:=sHttp.NewChannel()
	ch.Method = GET
	ch.Url = targetUrl
	ch.Errors = nil
	return ch
}

func (sHttp *SuperHttpClient) Post(targetUrl string) *SuperChannel {
	ch:=sHttp.NewChannel()
	ch.Method = POST
	ch.Url = targetUrl
	ch.Errors = nil
	return ch
}

func (sHttp *SuperHttpClient) Head(targetUrl string) *SuperChannel {
	ch:=sHttp.NewChannel()
	ch.Method = HEAD
	ch.Url = targetUrl
	ch.Errors = nil
	return ch
}

func (sHttp *SuperHttpClient) Put(targetUrl string) *SuperChannel {
	ch:=sHttp.NewChannel()
	ch.Method = PUT
	ch.Url = targetUrl
	ch.Errors = nil
	return ch
}

func (sHttp *SuperHttpClient) Delete(targetUrl string) *SuperChannel {
	ch:=sHttp.NewChannel()
	ch.Method = DELETE
	ch.Url = targetUrl
	ch.Errors = nil
	return ch
}

func (sHttp *SuperHttpClient) Patch(targetUrl string) *SuperChannel {
	ch:=sHttp.NewChannel()
	ch.Method = PATCH
	ch.Url = targetUrl
	ch.Errors = nil
	return ch
}

func (sHttp *SuperHttpClient) Options(targetUrl string) *SuperChannel {
	ch:=sHttp.NewChannel()
	ch.Method = OPTIONS
	ch.Url = targetUrl
	ch.Errors = nil
	return ch
}

// Set is used for setting header fields.
// Example. To set `Accept` as `application/json`
//
//    gorequest.New().
//      Post("/gamelist").
//      Set("Accept", "application/json").
//      End()
func (ch *SuperChannel) Set(param string, value string) *SuperChannel {
	ch.Header[param] = value
	return ch
}

// Retryable is used for setting a Retryer policy
// Example. To set Retryer policy with 5 seconds between each attempt.
//          3 max attempt.
//          And StatusBadRequest and StatusInternalServerError as RetryableStatus

//    gorequest.New().
//      Post("/gamelist").
//      Retry(3, 5 * time.seconds, http.StatusBadRequest, http.StatusInternalServerError).
//      End()
func (ch *SuperChannel) Retry(retryerCount int, retryerTime time.Duration, statusCode ...int) *SuperChannel {
	for _, code := range statusCode {
		statusText := http.StatusText(code)
		if len(statusText) == 0 {
			ch.Errors = append(ch.Errors, errors.New("StatusCode '"+strconv.Itoa(code)+"' doesn't exist in http package"))
		}
	}

	ch.http.Retryable = struct {
		RetryableStatus []int
		RetryerTime     time.Duration
		RetryerCount    int
		Attempt         int
		Enable          bool
	}{
		statusCode,
		retryerTime,
		retryerCount,
		0,
		true,
	}
	return ch
}

// SetBasicAuth sets the basic authentication header
// Example. To set the header for username "myuser" and password "mypass"
//
//    gorequest.New()
//      Post("/gamelist").
//      SetBasicAuth("myuser", "mypass").
//      End()
func (ch *SuperChannel) SetBasicAuth(username string, password string) *SuperChannel {
	ch.http.BasicAuth = struct{ Username, Password string }{username, password}
	return ch
}

// AddCookie adds a cookie to the request. The behavior is the same as AddCookie on Request from net/http
func (ch *SuperChannel) AddCookie(c *http.Cookie) *SuperChannel {
	ch.Cookies = append(ch.Cookies, c)
	return ch
}

// AddCookies is a convenient method to add multiple cookies
func (ch *SuperChannel) AddCookies(cookies []*http.Cookie) *SuperChannel {
	ch.Cookies = append(ch.Cookies, cookies...)
	return ch
}

var Types = map[string]string{
	"html":       "text/html",
	"json":       "application/json",
	"xml":        "application/xml",
	"text":       "text/plain",
	"urlencoded": "application/x-www-form-urlencoded",
	"form":       "application/x-www-form-urlencoded",
	"form-data":  "application/x-www-form-urlencoded",
	"multipart":  "multipart/form-data",
}

// Type is a convenience function to specify the data type to send.
// For example, to send data as `application/x-www-form-urlencoded` :
//
//    gorequest.New().
//      Post("/recipe").
//      Type("form").
//      Send(`{ "name": "egg benedict", "category": "brunch" }`).
//      End()
//
// This will POST the body "name=egg benedict&category=brunch" to url /recipe
//
// GoRequest supports
//
//    "text/html" uses "html"
//    "application/json" uses "json"
//    "application/xml" uses "xml"
//    "text/plain" uses "text"
//    "application/x-www-form-urlencoded" uses "urlencoded", "form" or "form-data"
//
func (ch *SuperChannel) Type(typeStr string) *SuperChannel {
	if _, ok := Types[typeStr]; ok {
		ch.ForceType = typeStr
	} else {
		ch.Errors = append(ch.Errors, errors.New("Type func: incorrect type \""+typeStr+"\""))
	}
	return ch
}

// Query function accepts either json string or strings which will form a query-string in url of GET method or body of POST method.
// For example, making "/search?query=bicycle&size=50x50&weight=20kg" using GET method:
//
//      gorequest.New().
//        Get("/search").
//        Query(`{ query: 'bicycle' }`).
//        Query(`{ size: '50x50' }`).
//        Query(`{ weight: '20kg' }`).
//        End()
//
// Or you can put multiple json values:
//
//      gorequest.New().
//        Get("/search").
//        Query(`{ query: 'bicycle', size: '50x50', weight: '20kg' }`).
//        End()
//
// Strings are also acceptable:
//
//      gorequest.New().
//        Get("/search").
//        Query("query=bicycle&size=50x50").
//        Query("weight=20kg").
//        End()
//
// Or even Mixed! :)
//
//      gorequest.New().
//        Get("/search").
//        Query("query=bicycle").
//        Query(`{ size: '50x50', weight:'20kg' }`).
//        End()
//
func (ch *SuperChannel) Query(content interface{}) *SuperChannel {
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		ch.queryString(v.String())
	case reflect.Struct:
		ch.queryStruct(v.Interface())
	case reflect.Map:
		ch.queryMap(v.Interface())
	default:
	}
	return ch
}

func (ch *SuperChannel) queryStruct(content interface{}) *SuperChannel {
	if marshalContent, err := json.Marshal(content); err != nil {
		ch.Errors = append(ch.Errors, err)
	} else {
		var val map[string]interface{}
		if err := json.Unmarshal(marshalContent, &val); err != nil {
			ch.Errors = append(ch.Errors, err)
		} else {
			for k, v := range val {
				k = strings.ToLower(k)
				var queryVal string
				switch t := v.(type) {
				case string:
					queryVal = t
				case float64:
					queryVal = strconv.FormatFloat(t, 'f', -1, 64)
				case time.Time:
					queryVal = t.Format(time.RFC3339)
				default:
					j, err := json.Marshal(v)
					if err != nil {
						continue
					}
					queryVal = string(j)
				}
				ch.QueryData.Add(k, queryVal)
			}
		}
	}
	return ch
}

func (ch *SuperChannel) queryString(content string) *SuperChannel {
	var val map[string]string
	if err := json.Unmarshal([]byte(content), &val); err == nil {
		for k, v := range val {
			ch.QueryData.Add(k, v)
		}
	} else {
		if queryData, err := url.ParseQuery(content); err == nil {
			for k, queryValues := range queryData {
				for _, queryValue := range queryValues {
					ch.QueryData.Add(k, string(queryValue))
				}
			}
		} else {
			ch.Errors = append(ch.Errors, err)
		}
		// TODO: need to check correct format of 'field=val&field=val&...'
	}
	return ch
}

func (ch *SuperChannel) queryMap(content interface{}) *SuperChannel {
	return ch.queryStruct(content)
}

// As Go conventions accepts ; as a synonym for &. (https://github.com/golang/go/issues/2210)
// Thus, Query won't accept ; in a querystring if we provide something like fields=f1;f2;f3
// This Param is then created as an alternative method to solve this.
func (ch *SuperChannel) Param(key string, value string) *SuperChannel {
	ch.QueryData.Add(key, value)
	return ch
}

func (ch *SuperChannel) Timeout(timeout time.Duration) *SuperChannel {
	ch.http.Transport.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			ch.Errors = append(ch.Errors, err)
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))
		return conn, nil
	}
	return ch
}

// Set TLSClientConfig for underling Transport.
// One example is you can use it to disable security check (https):
//
//      gorequest.New().TLSClientConfig(&tls.Config{ InsecureSkipVerify: true}).
//        Get("https://disable-security-check.com").
//        End()
//
func (ch *SuperChannel) TLSClientConfig(config *tls.Config) *SuperChannel {
	ch.http.Transport.TLSClientConfig = config
	return ch
}

// Proxy function accepts a proxy url string to setup proxy url for any request.
// It provides a convenience way to setup proxy which have advantages over usual old ways.
// One example is you might try to set `http_proxy` environment. This means you are setting proxy up for all the requests.
// You will not be able to send different request with different proxy unless you change your `http_proxy` environment again.
// Another example is using Golang proxy setting. This is normal prefer way to do but too verbase compared to GoRequest's Proxy:
//
//      gorequest.New().Proxy("http://myproxy:9999").
//        Post("http://www.google.com").
//        End()
//
// To set no_proxy, just put empty string to Proxy func:
//
//      gorequest.New().Proxy("").
//        Post("http://www.google.com").
//        End()
//
func (ch *SuperChannel) Proxy(proxyUrl string) *SuperChannel {
	parsedProxyUrl, err := url.Parse(proxyUrl)
	if err != nil {
		ch.Errors = append(ch.Errors, err)
	} else if proxyUrl == "" {
		ch.http.Transport.Proxy = nil
	} else {
		ch.http.Transport.Proxy = http.ProxyURL(parsedProxyUrl)
	}
	return ch
}

// RedirectPolicy accepts a function to define how to handle redirects. If the
// policy function returns an error, the next Request is not made and the previous
// request is returned.
//
// The policy function's arguments are the Request about to be made and the
// past requests in order of oldest first.
func (ch *SuperChannel) RedirectPolicy(policy func(req Request, via []Request) error) *SuperChannel {
	ch.http.Client.CheckRedirect = func(r *http.Request, v []*http.Request) error {
		vv := make([]Request, len(v))
		for i, r := range v {
			vv[i] = Request(r)
		}
		return policy(Request(r), vv)
	}
	return ch
}

// Send function accepts either json string or query strings which is usually used to assign data to POST or PUT method.
// Without specifying any type, if you give Send with json data, you are doing requesting in json format:
//
//      gorequest.New().
//        Post("/search").
//        Send(`{ query: 'sushi' }`).
//        End()
//
// While if you use at least one of querystring, GoRequest understands and automatically set the Content-Type to `application/x-www-form-urlencoded`
//
//      gorequest.New().
//        Post("/search").
//        Send("query=tonkatsu").
//        End()
//
// So, if you want to strictly send json format, you need to use Type func to set it as `json` (Please see more details in Type function).
// You can also do multiple chain of Send:
//
//      gorequest.New().
//        Post("/search").
//        Send("query=bicycle&size=50x50").
//        Send(`{ wheel: '4'}`).
//        End()
//
// From v0.2.0, Send function provide another convenience way to work with Struct type. You can mix and match it with json and query string:
//
//      type BrowserVersionSupport struct {
//        Chrome string
//        Firefox string
//      }
//      ver := BrowserVersionSupport{ Chrome: "37.0.2041.6", Firefox: "30.0" }
//      gorequest.New().
//        Post("/update_version").
//        Send(ver).
//        Send(`{"Safari":"5.1.10"}`).
//        End()
//
// If you have set Type to text or Content-Type to text/plain, content will be sent as raw string in body instead of form
//
//      gorequest.New().
//        Post("/greet").
//        Type("text").
//        Send("hello world").
//        End()
//
func (ch *SuperChannel) Send(content interface{}) *SuperChannel {
	// TODO: add normal text mode or other mode to Send func
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		ch.SendString(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // includes rune
		ch.SendString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // includes byte
		ch.SendString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float64:
		ch.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Float32:
		ch.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 32))
	case reflect.Bool:
		ch.SendString(strconv.FormatBool(v.Bool()))
	case reflect.Struct:
		ch.SendStruct(v.Interface())
	case reflect.Slice:
		ch.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Array:
		ch.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Ptr:
		ch.Send(v.Elem().Interface())
	case reflect.Map:
		ch.SendMap(v.Interface())
	default:
		// TODO: leave default for handling other types in the future, such as complex numbers, (nested) maps, etc
		return ch
	}
	return ch
}

func makeSliceOfReflectValue(v reflect.Value) (slice []interface{}) {

	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return slice
	}

	slice = make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		slice[i] = v.Index(i).Interface()
	}

	return slice
}

// SendSlice (similar to SendString) returns SuperChannel's itself for any next chain and takes content []interface{} as a parameter.
// Its duty is to append slice of interface{} into ch.SliceData ([]interface{}) which later changes into json array in the End() func.
func (ch *SuperChannel) SendSlice(content []interface{}) *SuperChannel {
	ch.SliceData = append(ch.SliceData, content...)
	return ch
}

func (ch *SuperChannel) SendMap(content interface{}) *SuperChannel {
	return ch.SendStruct(content)
}

// SendStruct (similar to SendString) returns SuperChannel's itself for any next chain and takes content interface{} as a parameter.
// Its duty is to transfrom interface{} (implicitly always a struct) into ch.Data (map[string]interface{}) which later changes into appropriate format such as json, form, text, etc. in the End() func.
func (ch *SuperChannel) SendStruct(content interface{}) *SuperChannel {
	if marshalContent, err := json.Marshal(content); err != nil {
		ch.Errors = append(ch.Errors, err)
	} else {
		var val map[string]interface{}
		d := json.NewDecoder(bytes.NewBuffer(marshalContent))
		d.UseNumber()
		if err := d.Decode(&val); err != nil {
			ch.Errors = append(ch.Errors, err)
		} else {
			for k, v := range val {
				ch.Data[k] = v
			}
		}
	}
	return ch
}

// SendString returns SuperChannel's itself for any next chain and takes content string as a parameter.
// Its duty is to transform String into ch.Data (map[string]interface{}) which later changes into appropriate format such as json, form, text, etc. in the End func.
// Send implicitly uses SendString and you should use Send instead of this.
func (ch *SuperChannel) SendString(content string) *SuperChannel {
	if !ch.BounceToRawString {
		var val interface{}
		d := json.NewDecoder(strings.NewReader(content))
		d.UseNumber()
		if err := d.Decode(&val); err == nil {
			switch v := reflect.ValueOf(val); v.Kind() {
			case reflect.Map:
				for k, v := range val.(map[string]interface{}) {
					ch.Data[k] = v
				}
			// add to SliceData
			case reflect.Slice:
				ch.SendSlice(val.([]interface{}))
			// bounce to rawstring if it is arrayjson, or others
			default:
				ch.BounceToRawString = true
			}
		} else if formData, err := url.ParseQuery(content); err == nil {
			for k, formValues := range formData {
				for _, formValue := range formValues {
					// make it array if already have key
					if val, ok := ch.Data[k]; ok {
						var strArray []string
						strArray = append(strArray, string(formValue))
						// check if previous data is one string or array
						switch oldValue := val.(type) {
						case []string:
							strArray = append(strArray, oldValue...)
						case string:
							strArray = append(strArray, oldValue)
						}
						ch.Data[k] = strArray
					} else {
						// make it just string if does not already have same key
						ch.Data[k] = formValue
					}
				}
			}
			ch.TargetType = "form"
		} else {
			ch.BounceToRawString = true
		}
	}
	// Dump all contents to RawString in case in the end user doesn't want json or form.
	ch.RawString += content
	return ch
}

type File struct {
	Filename  string
	Fieldname string
	Data      []byte
}

// SendFile function works only with type "multipart". The function accepts one mandatory and up to two optional arguments. The mandatory (first) argument is the file.
// The function accepts a path to a file as string:
//
//      gorequest.New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile("./example_file.ext").
//        End()
//
// File can also be a []byte slice of a already file read by eg. ioutil.ReadFile:
//
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      gorequest.New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b).
//        End()
//
// Furthermore file can also be a os.File:
//
//      f, _ := os.Open("./example_file.ext")
//      gorequest.New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(f).
//        End()
//
// The first optional argument (second argument overall) is the filename, which will be automatically determined when file is a string (path) or a os.File.
// When file is a []byte slice, filename defaults to "filename". In all cases the automatically determined filename can be overwritten:
//
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      gorequest.New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b, "my_custom_filename").
//        End()
//
// The second optional argument (third argument overall) is the fieldname in the multipart/form-data request. It defaults to fileNUMBER (eg. file1), where number is ascending and starts counting at 1.
// So if you send multiple files, the fieldnames will be file1, file2, ... unless it is overwritten. If fieldname is set to "file" it will be automatically set to fileNUMBER, where number is the greatest exsiting number+1.
//
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      gorequest.New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b, "", "my_custom_fieldname"). // filename left blank, will become "example_file.ext"
//        End()
//
func (ch *SuperChannel) SendFile(file interface{}, args ...string) *SuperChannel {

	filename := ""
	fieldname := "file"

	if len(args) >= 1 && len(args[0]) > 0 {
		filename = strings.TrimSpace(args[0])
	}
	if len(args) >= 2 && len(args[1]) > 0 {
		fieldname = strings.TrimSpace(args[1])
	}
	if fieldname == "file" || fieldname == "" {
		fieldname = "file" + strconv.Itoa(len(ch.FileData)+1)
	}

	switch v := reflect.ValueOf(file); v.Kind() {
	case reflect.String:
		pathToFile, err := filepath.Abs(v.String())
		if err != nil {
			ch.Errors = append(ch.Errors, err)
			return ch
		}
		if filename == "" {
			filename = filepath.Base(pathToFile)
		}
		data, err := ioutil.ReadFile(v.String())
		if err != nil {
			ch.Errors = append(ch.Errors, err)
			return ch
		}
		ch.FileData = append(ch.FileData, File{
			Filename:  filename,
			Fieldname: fieldname,
			Data:      data,
		})
	case reflect.Slice:
		slice := makeSliceOfReflectValue(v)
		if filename == "" {
			filename = "filename"
		}
		f := File{
			Filename:  filename,
			Fieldname: fieldname,
			Data:      make([]byte, len(slice)),
		}
		for i := range slice {
			f.Data[i] = slice[i].(byte)
		}
		ch.FileData = append(ch.FileData, f)
	case reflect.Ptr:
		if len(args) == 1 {
			return ch.SendFile(v.Elem().Interface(), args[0])
		}
		if len(args) >= 2 {
			return ch.SendFile(v.Elem().Interface(), args[0], args[1])
		}
		return ch.SendFile(v.Elem().Interface())
	default:
		if v.Type() == reflect.TypeOf(os.File{}) {
			osfile := v.Interface().(os.File)
			if filename == "" {
				filename = filepath.Base(osfile.Name())
			}
			data, err := ioutil.ReadFile(osfile.Name())
			if err != nil {
				ch.Errors = append(ch.Errors, err)
				return ch
			}
			ch.FileData = append(ch.FileData, File{
				Filename:  filename,
				Fieldname: fieldname,
				Data:      data,
			})
			return ch
		}

		ch.Errors = append(ch.Errors, errors.New("SendFile currently only supports either a string (path/to/file), a slice of bytes (file content itself), or a os.File!"))
	}

	return ch
}

func changeMapToURLValues(data map[string]interface{}) url.Values {
	var newUrlValues = url.Values{}
	for k, v := range data {
		switch val := v.(type) {
		case string:
			newUrlValues.Add(k, val)
		case bool:
			newUrlValues.Add(k, strconv.FormatBool(val))
		// if a number, change to string
		// json.Number used to protect against a wrong (for GoRequest) default conversion
		// which always converts number to float64.
		// This type is caused by using Decoder.UseNumber()
		case json.Number:
			newUrlValues.Add(k, string(val))
		case int:
			newUrlValues.Add(k, strconv.FormatInt(int64(val), 10))
		// TODO add all other int-Types (int8, int16, ...)
		case float64:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		case float32:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		// following slices are mostly needed for tests
		case []string:
			for _, element := range val {
				newUrlValues.Add(k, element)
			}
		case []int:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatInt(int64(element), 10))
			}
		case []bool:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatBool(element))
			}
		case []float64:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		case []float32:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		// these slices are used in practice like sending a struct
		case []interface{}:

			if len(val) <= 0 {
				continue
			}

			switch val[0].(type) {
			case string:
				for _, element := range val {
					newUrlValues.Add(k, element.(string))
				}
			case bool:
				for _, element := range val {
					newUrlValues.Add(k, strconv.FormatBool(element.(bool)))
				}
			case json.Number:
				for _, element := range val {
					newUrlValues.Add(k, string(element.(json.Number)))
				}
			}
		default:
			// TODO add ptr, arrays, ...
		}
	}
	return newUrlValues
}

// End is the most important function that you need to call when ending the chain. The request won't proceed without calling it.
// End function returns Response which matchs the structure of Response type in Golang's http package (but without Body data). The body data itself returns as a string in a 2nd return value.
// Lastly but worth noticing, error array (NOTE: not just single error value) is returned as a 3rd value and nil otherwise.
//
// For example:
//
//    resp, body, errs := gorequest.New().Get("http://www.google.com").End()
//    if (errs != nil) {
//      fmt.Println(errs)
//    }
//    fmt.Println(resp, body)
//
// Moreover, End function also supports callback which you can put as a parameter.
// This extends the flexibility and makes GoRequest fun and clean! You can use GoRequest in whatever style you love!
//
// For example:
//
//    func printBody(resp gorequest.Response, body string, errs []error){
//      fmt.Println(resp.Status)
//    }
//    gorequest.New().Get("http://www..google.com").End(printBody)
//
func (ch *SuperChannel) End(callback ...func(response Response, body string, errs []error)) (Response, string, []error) {
	var bytesCallback []func(response Response, body []byte, errs []error)
	if len(callback) > 0 {
		bytesCallback = []func(response Response, body []byte, errs []error){
			func(response Response, body []byte, errs []error) {
				callback[0](response, string(body), errs)
			},
		}
	}

	resp, body, errs := ch.EndBytes(bytesCallback...)
	bodyString := string(body)

	return resp, bodyString, errs
}

// EndBytes should be used when you want the body as bytes. The callbacks work the same way as with `End`, except that a byte array is used instead of a string.
func (ch *SuperChannel) EndBytes(callback ...func(response Response, body []byte, errs []error)) (Response, []byte, []error) {
	var (
		errs []error
		resp Response
		body []byte
	)

	for {
		resp, body, errs = ch.getResponseBytes()
		if errs != nil {
			return nil, nil, errs
		}
		if ch.isRetryableRequest(resp) {
			resp.Header.Set("Retry-Count", strconv.Itoa(ch.http.Retryable.Attempt))
			break
		}
	}

	respCallback := *resp
	if len(callback) != 0 {
		callback[0](&respCallback, body, ch.Errors)
	}
	return resp, body, nil
}

func (ch *SuperChannel) isRetryableRequest(resp Response) bool {
	if ch.http.Retryable.Enable && ch.http.Retryable.Attempt < ch.http.Retryable.RetryerCount && contains(resp.StatusCode, ch.http.Retryable.RetryableStatus) {
		time.Sleep(ch.http.Retryable.RetryerTime)
		ch.http.Retryable.Attempt++
		return false
	}
	return true
}

func contains(respStatus int, statuses []int) bool {
	for _, status := range statuses {
		if status == respStatus {
			return true
		}
	}
	return false
}

// EndStruct should be used when you want the body as a struct. The callbacks work the same way as with `End`, except that a struct is used instead of a string.
func (ch *SuperChannel) EndStruct(v interface{}, callback ...func(response Response, v interface{}, body []byte, errs []error)) (Response, []byte, []error) {
	resp, body, errs := ch.EndBytes()
	if errs != nil {
		return nil, body, errs
	}
	err := json.Unmarshal(body, &v)
	if err != nil {
		ch.Errors = append(ch.Errors, err)
		return resp, body, ch.Errors
	}
	respCallback := *resp
	if len(callback) != 0 {
		callback[0](&respCallback, v, body, ch.Errors)
	}
	return resp, body, nil
}

func (ch *SuperChannel) getResponseBytes() (Response, []byte, []error) {
	var (
		req  *http.Request
		err  error
		resp Response
	)
	// check whether there is an error. if yes, return all errors
	if len(ch.Errors) != 0 {
		return nil, nil, ch.Errors
	}
	// check if there is forced type
	switch ch.ForceType {
	case "json", "form", "xml", "text", "multipart":
		ch.TargetType = ch.ForceType
		// If forcetype is not set, check whether user set Content-Type header.
		// If yes, also bounce to the correct supported TargetType automatically.
	default:
		for k, v := range Types {
			if ch.Header["Content-Type"] == v {
				ch.TargetType = k
			}
		}
	}

	// if slice and map get mixed, let's bounce to rawstring
	if len(ch.Data) != 0 && len(ch.SliceData) != 0 {
		ch.BounceToRawString = true
	}

	// Make Request
	req, err = ch.MakeRequest()
	if err != nil {
		ch.Errors = append(ch.Errors, err)
		return nil, nil, ch.Errors
	}

	// Set Transport
	if !DisableTransportSwap {
		ch.http.Client.Transport = ch.http.Transport
	}

	// Log details of this request
	if ch.http.Debug {
		dump, err := httputil.DumpRequest(req, true)
		ch.http.logger.SetPrefix("[http] ")
		if err != nil {
			ch.http.logger.Println("Error:", err)
		} else {
			ch.http.logger.Printf("HTTP Request: %s", string(dump))
		}
	}

	// Display CURL command line
	if ch.http.CurlCommand {
		curl, err := http2curl.GetCurlCommand(req)
		ch.http.logger.SetPrefix("[curl] ")
		if err != nil {
			ch.http.logger.Println("Error:", err)
		} else {
			ch.http.logger.Printf("CURL command line: %s", curl)
		}
	}

	// Send request
	resp, err = ch.http.Client.Do(req)
	if err != nil {
		ch.Errors = append(ch.Errors, err)
		return nil, nil, ch.Errors
	}
	defer resp.Body.Close()

	// Log details of this response
	if ch.http.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if nil != err {
			ch.http.logger.Println("Error:", err)
		} else {
			ch.http.logger.Printf("HTTP Response: %s", string(dump))
		}
	}

	body, _ := ioutil.ReadAll(resp.Body)
	// Reset resp.Body so it can be use again
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return resp, body, nil
}

func (ch *SuperChannel) MakeRequest() (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	if ch.Method == "" {
		return nil, errors.New("No method specified")
	}

	if ch.TargetType == "json" {
		// If-case to give support to json array. we check if
		// 1) Map only: send it as json map from ch.Data
		// 2) Array or Mix of map & array or others: send it as rawstring from ch.RawString
		var contentJson []byte
		if ch.BounceToRawString {
			contentJson = []byte(ch.RawString)
		} else if len(ch.Data) != 0 {
			contentJson, _ = json.Marshal(ch.Data)
		} else if len(ch.SliceData) != 0 {
			contentJson, _ = json.Marshal(ch.SliceData)
		}
		contentReader := bytes.NewReader(contentJson)
		req, err = http.NewRequest(ch.Method, ch.Url, contentReader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else if ch.TargetType == "form" || ch.TargetType == "form-data" || ch.TargetType == "urlencoded" {
		var contentForm []byte
		if ch.BounceToRawString || len(ch.SliceData) != 0 {
			contentForm = []byte(ch.RawString)
		} else {
			formData := changeMapToURLValues(ch.Data)
			contentForm = []byte(formData.Encode())
		}
		contentReader := bytes.NewReader(contentForm)
		req, err = http.NewRequest(ch.Method, ch.Url, contentReader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if ch.TargetType == "text" {
		req, err = http.NewRequest(ch.Method, ch.Url, strings.NewReader(ch.RawString))
		req.Header.Set("Content-Type", "text/plain")
	} else if ch.TargetType == "xml" {
		req, err = http.NewRequest(ch.Method, ch.Url, strings.NewReader(ch.RawString))
		req.Header.Set("Content-Type", "application/xml")
	} else if ch.TargetType == "multipart" {

		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)

		if ch.BounceToRawString {
			fieldName, ok := ch.Header["data_fieldname"]
			if !ok {
				fieldName = "data"
			}
			fw, _ := mw.CreateFormField(fieldName)
			fw.Write([]byte(ch.RawString))
		}

		if len(ch.Data) != 0 {
			formData := changeMapToURLValues(ch.Data)
			for key, values := range formData {
				for _, value := range values {
					fw, _ := mw.CreateFormField(key)
					fw.Write([]byte(value))
				}
			}
		}

		if len(ch.SliceData) != 0 {
			fieldName, ok := ch.Header["json_fieldname"]
			if !ok {
				fieldName = "data"
			}
			// copied from CreateFormField() in mime/multipart/writer.go
			h := make(textproto.MIMEHeader)
			fieldName = strings.Replace(strings.Replace(fieldName, "\\", "\\\\", -1), `"`, "\\\"", -1)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, fieldName))
			h.Set("Content-Type", "application/json")
			fw, _ := mw.CreatePart(h)
			contentJson, err := json.Marshal(ch.SliceData)
			if err != nil {
				return nil, err
			}
			fw.Write(contentJson)
		}

		// add the files
		if len(ch.FileData) != 0 {
			for _, file := range ch.FileData {
				fw, _ := mw.CreateFormFile(file.Fieldname, file.Filename)
				fw.Write(file.Data)
			}
		}

		// close before call to FormDataContentType ! otherwise its not valid multipart
		mw.Close()

		req, err = http.NewRequest(ch.Method, ch.Url, &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
	} else {
		// let's return an error instead of an nil pointer exception here
		return nil, errors.New("TargetType '" + ch.TargetType + "' could not be determined")
	}

	for k, v := range ch.Header {
		req.Header.Set(k, v)
		// Setting the host header is a special case, see this issue: https://github.com/golang/go/issues/7682
		if strings.EqualFold(k, "host") {
			req.Host = v
		}
	}
	// Add all querystring from Query func
	q := req.URL.Query()
	for k, v := range ch.QueryData {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	req.URL.RawQuery = q.Encode()

	// Add basic auth
	if ch.http.BasicAuth != struct{ Username, Password string }{} {
		req.SetBasicAuth(ch.http.BasicAuth.Username, ch.http.BasicAuth.Password)
	}

	// Add cookies
	for _, cookie := range ch.Cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

// AsCurlCommand returns a string representing the runnable `curl' command
// version of the request.
func (ch *SuperChannel) AsCurlCommand() (string, error) {
	req, err := ch.MakeRequest()
	if err != nil {
		return "", err
	}
	cmd, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return "", err
	}
	return cmd.String(), nil
}
