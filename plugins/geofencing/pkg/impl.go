package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	envoycorev2 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoyauthv2 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	"github.com/solo-io/ext-auth-plugins/api"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

var (
	UnexpectedConfigError = func(typ interface{}) error {
		return errors.New(fmt.Sprintf("unexpected config type: %T", typ))
	}
	_ api.ExtAuthPlugin = new(GeoFencingPlugin)
)

type GeoFencingPlugin struct{}

type Config struct {
	Endpoint string
}

func (p *GeoFencingPlugin) NewConfigInstance(ctx context.Context) (interface{}, error) {
	return &Config{}, nil
}

func (p *GeoFencingPlugin) GetAuthService(ctx context.Context, configInstance interface{}) (api.AuthService, error) {
	config, ok := configInstance.(*Config)
	if !ok {
		return nil, UnexpectedConfigError(configInstance)
	}

	logger(ctx).Infow("Parsed GeoFencingAuthService config",
		zap.Any("Endpoint", config.Endpoint),
	)

	return &GeoFencingAuthService{
		Endpoint: config.Endpoint,
	}, nil
}

type GeoFencingAuthService struct {
	Endpoint string
}

func (c *GeoFencingAuthService) Start(context.Context) error {
	// no-op
	return nil
}

func (c *GeoFencingAuthService) Authorize(ctx context.Context, request *api.AuthorizationRequest) (*api.AuthorizationResponse, error) {
	for key, value := range request.CheckRequest.GetAttributes().GetRequest().GetHttp().GetHeaders() {
		if key == "X-Forwarded-For" || key == "x-forwarded-for" {
			var ip string

			if strings.Contains(value, ",") {
				ip = strings.Split(value, ",")[0]
			} else {
				ip = value
			}

			logger(ctx).Infof("ip address: %s", ip)

			logger(ctx).Infof("geofencing data: %+v", data)
			response := api.AuthorizedResponse()

			data, err := getGeoFencingData(c.Endpoint, ip)
			if data.Status == "success" && err == nil {
				response.CheckResponse.HttpResponse = &envoyauthv2.CheckResponse_OkResponse{
					OkResponse: &envoyauthv2.OkHttpResponse{
						Headers: []*envoycorev2.HeaderValueOption{
							{
								Header: &envoycorev2.HeaderValue{
									Key:   "x-geolocation",
									Value: fmt.Sprintf("lat=%v;lon=%v;country=%v;region=%v;", data.Lat, data.Lon, data.CountryCode, data.Region),
								},
							},
						},
					},
				}
			}

			logger(ctx).Infof("geofencing data: %+v", response.CheckResponse.GetOkResponse().String())

			return response, nil
		}
	}
	logger(ctx).Infow("Could not get GeoFencing data, denying access")
	return api.UnauthorizedResponse(), nil
}

type GeoFencingResponse struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}

func getGeoFencingData(endpoint, ip string) (*GeoFencingResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", endpoint, ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gfResp := GeoFencingResponse{}

	if err := json.Unmarshal(body, &gfResp); err != nil {
		return nil, err
	}

	return &gfResp, nil
}

func logger(ctx context.Context) *zap.SugaredLogger {
	return contextutils.LoggerFrom(contextutils.WithLogger(ctx, "geofencing_plugin"))
}
