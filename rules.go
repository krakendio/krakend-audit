package audit

import (
	botdetector "github.com/krakendio/krakend-botdetector/v2/krakend"
	cb "github.com/krakendio/krakend-circuitbreaker/v2/gobreaker"
	cors "github.com/krakendio/krakend-cors/v2"
	gelf "github.com/krakendio/krakend-gelf/v2"
	gologging "github.com/krakendio/krakend-gologging/v2"
	httpsecure "github.com/krakendio/krakend-httpsecure/v2"
	jose "github.com/krakendio/krakend-jose/v2"
	logstash "github.com/krakendio/krakend-logstash/v2"
	metrics "github.com/krakendio/krakend-metrics/v2"
	opencensus "github.com/krakendio/krakend-opencensus/v2"
	ratelimitProxy "github.com/krakendio/krakend-ratelimit/v2/juju/proxy"
	ratelimit "github.com/krakendio/krakend-ratelimit/v2/juju/router"
	router "github.com/luraproject/lura/v2/router/gin"
	server "github.com/luraproject/lura/v2/transport/http/server/plugin"
)

func hasBit(x float64, y int) bool {
	return (int(x)>>y)&1 == 1
}

func hasBasicAuth(s *Service) bool {
	return len(s.Components[server.Namespace]) > 0 && hasBit(float64(s.Components[server.Namespace][0]), parseServerPlugin("basic-auth"))
}

func hasApiKeys(s *Service) bool {
	_, ok := s.Components["auth/api-keys"]
	return ok
}

func hasNoJWT(s *Service) bool {
	for _, e := range s.Endpoints {
		if _, ok := e.Components[jose.ValidatorNamespace]; ok {
			return false
		}
	}
	return true
}

func hasInsecureConnections(s *Service) bool {
	return hasBit(float64(s.Details[0]), ServiceAllowInsecureConnections)
}

func hasNoTLS(s *Service) bool {
	return !hasBit(float64(s.Details[0]), ServiceHasTLS)
}

func hasTLSDisabled(s *Service) bool {
	return hasBit(float64(s.Details[0]), ServiceHasTLS) && !hasBit(float64(s.Details[0]), ServiceTLSEnabled)
}

func hasNoHTTPSecure(s *Service) bool {
	_, ok := s.Components[httpsecure.Namespace]
	return !ok
}

func hasNoObfuscatedVersionHeader(s *Service) bool {
	v, ok := s.Components[router.Namespace]
	if !ok || len(v) == 0 {
		return true
	}
	return !hasBit(float64(v[0]), RouterHideVersionHeader)
}

func hasNoCORS(s *Service) bool {
	_, ok := s.Components[cors.Namespace]
	return !ok
}

func hasBotdetectorDisabled(s *Service) bool {
	_, ok := s.Components[botdetector.Namespace]
	return !ok
}

func hasNoRatelimit(s *Service) bool {
	_, ok := s.Components[ratelimit.Namespace]
	if ok {
		return false
	}
	for _, e := range s.Endpoints {
		_, ok := e.Components[ratelimit.Namespace]
		if ok {
			return false
		}
		_, ok = e.Components[ratelimitProxy.Namespace]
		if ok {
			return false
		}
		for _, b := range e.Backends {
			_, ok := b.Components[ratelimitProxy.Namespace]
			if ok {
				return false
			}
		}
	}
	return true
}

func hasNoCB(s *Service) bool {
	for _, e := range s.Endpoints {
		_, ok := e.Components[cb.Namespace]
		if ok {
			return false
		}
		for _, b := range e.Backends {
			_, ok := b.Components[cb.Namespace]
			if ok {
				return false
			}
		}
	}
	return true
}

func hasTimeoutBiggerThan(d int) func(*Service) bool {
	return func(s *Service) bool {
		for _, e := range s.Endpoints {
			if e.Details[3] > d {
				return true
			}
		}
		return false
	}
}

func hasNoMetrics(s *Service) bool {
	for _, k := range []string{
		opencensus.Namespace,
		metrics.Namespace,
		"telemetry/newrelic",
		"telemetry/ganalytics",
		"telemetry/instana",
	} {
		if _, ok := s.Components[k]; ok {
			return false
		}
	}
	return true
}

func hasSeveralTelemetryComponents(s *Service) bool {
	tot := 0
	for _, k := range []string{
		opencensus.Namespace,
		metrics.Namespace,
		"telemetry/newrelic",
		"telemetry/ganalytics",
		"telemetry/instana",
	} {
		if _, ok := s.Components[k]; ok {
			tot++
		}
	}
	return tot > 1
}

func hasNoTracing(s *Service) bool {
	_, ok1 := s.Components[opencensus.Namespace]
	_, ok2 := s.Components["telemetry/newrelic"]
	_, ok3 := s.Components["telemetry/instana"]
	return !ok1 && !ok2 && !ok3
}

func hasNoLogging(s *Service) bool {
	_, ok1 := s.Components[gologging.Namespace]
	_, ok2 := s.Components[gelf.Namespace]
	_, ok3 := s.Components[logstash.Namespace]
	return !ok1 && !ok2 && !ok3
}

func hasRestfulDisabled(s *Service) bool {
	return hasBit(float64(s.Details[0]), ServiceDisableStrictREST)
}

func hasDebugEnabled(s *Service) bool {
	return hasBit(float64(s.Details[0]), ServiceDebug)
}

func hasEndpointWithoutBackends(s *Service) bool {
	for _, e := range s.Endpoints {
		if len(e.Backends) == 0 {
			return true
		}
	}
	return false
}

func hasASingleBackendPerEndpoint(s *Service) bool {
	for _, e := range s.Endpoints {
		if len(e.Backends) > 1 {
			return false
		}
	}
	return true
}

func hasAllEndpointsAsNoop(s *Service) bool {
	for _, e := range s.Endpoints {
		if !hasBit(float64(e.Details[0]), EncodingNOOP) {
			return false
		}
	}
	return true
}

func hasSequentialStart(s *Service) bool {
	return hasBit(float64(s.Details[0]), ServiceSequentialStart) && len(s.Agents) >= 10
}
