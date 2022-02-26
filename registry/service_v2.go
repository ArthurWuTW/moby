package registry // import "github.com/docker/docker/registry"

import (
	"net/url"
	"strings"

	"github.com/docker/go-connections/tlsconfig"
)

func (s *defaultService) lookupV2Endpoints(hostname string) (endpoints []APIEndpoint, err error) {
	if hostname == DefaultNamespace || hostname == IndexHostname {
		for _, mirror := range s.config.Mirrors {
			if !strings.HasPrefix(mirror, "http://") && !strings.HasPrefix(mirror, "https://") {
				mirror = "https://" + mirror
			}
			mirrorURL, err := url.Parse(mirror)
			if err != nil {
				return nil, invalidParam(err)
			}
			mirrorTLSConfig, err := newTLSConfig(mirrorURL.Host, isSecureIndex(s.config, mirrorURL.Host))
			if err != nil {
				return nil, err
			}
			endpoints = append(endpoints, APIEndpoint{
				URL:          mirrorURL,
				Version:      APIVersion2,
				Mirror:       true,
				TrimHostname: true,
				TLSConfig:    mirrorTLSConfig,
			})
		}
		endpoints = append(endpoints, APIEndpoint{
			URL:          DefaultV2Registry,
			Version:      APIVersion2,
			Official:     true,
			TrimHostname: true,
			TLSConfig:    tlsconfig.ServerDefault(),
		})

		return endpoints, nil
	}

	tlsConfig, err := newTLSConfig(hostname, isSecureIndex(s.config, hostname))
	if err != nil {
		return nil, err
	}

	ana := allowNondistributableArtifacts(s.config, hostname)
	endpoints = []APIEndpoint{
		{
			URL: &url.URL{
				Scheme: "https",
				Host:   hostname,
			},
			Version:                        APIVersion2,
			AllowNondistributableArtifacts: ana,
			TrimHostname:                   true,
			TLSConfig:                      tlsConfig,
		},
	}

	if tlsConfig.InsecureSkipVerify {
		endpoints = append(endpoints, APIEndpoint{
			URL: &url.URL{
				Scheme: "http",
				Host:   hostname,
			},
			Version:                        APIVersion2,
			AllowNondistributableArtifacts: ana,
			TrimHostname:                   true,
			// used to check if supposed to be secure via InsecureSkipVerify
			TLSConfig: tlsConfig,
		})
	}

	return endpoints, nil
}
