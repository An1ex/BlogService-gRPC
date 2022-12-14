package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	APP_KEY     = "an1ex_key"
	APP_SERCRET = "an1ex_secret"
)

type HttpResponse struct {
	Code    int               `json:"code"`
	Msg     string            `json:"msg"`
	Details map[string]string `json:"details"`
}

type ListContent struct {
	List  interface{}    `json:"list"`
	Pager map[string]int `json:"pager"`
}

type HttpResponseList struct {
	Code    int                    `json:"code"`
	Msg     string                 `json:"msg"`
	Details map[string]ListContent `json:"details"`
}

type API struct {
	URL string
}

func NewAPI(url string) *API {
	return &API{URL: url}
}

func (a *API) httpGet(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequest("GET", a.URL+path, nil)
	if err != nil {
		return nil, err
	}
	//// 方法二，自定义Client方法，实现超时
	// client := http.Client{
	//	Transport: http.DefaultTransport,
	//	Timeout:   60 * time.Second,
	//}
	//resp, err := client.Do(req)
	//if err != nil {
	//	return nil, err
	//}

	// HTTP追踪部分
	span, newCtx := opentracing.StartSpanFromContext(ctx, "HTTP GET:"+a.URL, opentracing.Tag{
		Key:   string(ext.Component),
		Value: "HTTP",
	})
	defer span.Finish()

	span.SetTag("url", path)
	_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	// HTTP追踪部分

	req = req.WithContext(newCtx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (a *API) getAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("/%s?app_key=%s&app_secret=%s", "auth", APP_KEY, APP_SERCRET)
	body, err := a.httpGet(ctx, url)
	if err != nil {
		return "", err
	}
	var accessToken HttpResponse
	err = json.Unmarshal(body, &accessToken)
	if err != nil {
		return "", err
	}
	return accessToken.Details["token"], nil
}

func (a *API) GetTagList(ctx context.Context, name string, state uint8) ([]byte, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	body, err := a.httpGet(ctx, fmt.Sprintf("/%s?token=%s&name=%s&state=%s", "api/v1/tags", token, name, strconv.Itoa(int(state))))
	if err != nil {
		return nil, err
	}
	return body, nil
}
