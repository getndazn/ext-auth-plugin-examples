package impl

import (
	"context"
	"errors"
	"fmt"
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

type Config struct{}

func (p *GeoFencingPlugin) NewConfigInstance(ctx context.Context) (interface{}, error) {
	return &Config{}, nil
}

func (p *GeoFencingPlugin) GetAuthService(ctx context.Context, configInstance interface{}) (api.AuthService, error) {
	// config, ok := configInstance.(*Config)
	// if !ok {
	// 	return nil, UnexpectedConfigError(configInstance)
	// }

	return &GeoFencingAuthService{}, nil
}

type (
	GeoFencingAuthService       struct{}
	additionalHeaderBuilderFunc func(requestHeaders map[string]string) (*envoycorev2.HeaderValueOption, error)
)

func (c *GeoFencingAuthService) Start(context.Context) error {
	// no-op
	return nil
}

func (c *GeoFencingAuthService) Authorize(ctx context.Context, request *api.AuthorizationRequest) (*api.AuthorizationResponse, error) {
	var additionalHeaderBuilders []additionalHeaderBuilderFunc = []additionalHeaderBuilderFunc{
		getGeoLocationHeader,
	}

	requestHeaders := request.CheckRequest.GetAttributes().GetRequest().GetHttp().GetHeaders()
	responseHeaders := make([]*envoycorev2.HeaderValueOption, len(requestHeaders)+len(additionalHeaderBuilders))

	i := 0
	// 1: copy all the existing headers....
	for key, value := range requestHeaders {
		responseHeaders[i] = &envoycorev2.HeaderValueOption{
			Header: &envoycorev2.HeaderValue{
				Key:   key,
				Value: value,
			},
		}
		i++
	}

	for _, fn := range additionalHeaderBuilders {
		if r, err := fn(requestHeaders); err != nil {
			responseHeaders[i] = r
			i++
		} else {
			logger(ctx).Warn(err)
		}
	}

	response := api.AuthorizedResponse()
	response.CheckResponse.HttpResponse = &envoyauthv2.CheckResponse_OkResponse{
		OkResponse: &envoyauthv2.OkHttpResponse{Headers: responseHeaders},
	}

	return response, nil
}

func logger(ctx context.Context) *zap.SugaredLogger {
	return contextutils.LoggerFrom(contextutils.WithLogger(ctx, "geofencing_plugin"))
}

func getGeoLocationHeader(requestHeaders map[string]string) (*envoycorev2.HeaderValueOption, error) {
	for key, value := range requestHeaders {
		if key == "X-Forwarded-For" || key == "x-forwarded-for" {
			ip := strings.SplitN(value, ",", 1)[0]
			data, err := getGeoLocationData(ip)

			return &envoycorev2.HeaderValueOption{
				Header: &envoycorev2.HeaderValue{
					Key:   "x-geolocation",
					Value: data.headerString(),
				},
			}, err
		}
	}

	return nil, errors.New("no x-forwarded-for header")
}
