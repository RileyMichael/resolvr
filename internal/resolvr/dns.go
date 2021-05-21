package resolvr

import (
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	// TODO: ipv6.?
	ipDashRegex = regexp.MustCompile(`(^|[.-])(((25[0-5]|(2[0-4]|1?[0-9])?[0-9])-){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9]))($|[.-])`)
	aRecords    map[string]net.IP
	nsRecords   []dns.RR
	soaRecord   []dns.RR
	dnsRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_total",
		Help: "The total number of DNS requests",
	})
	typeAQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_type_a",
		Help: "The total number of Type A DNS Query requests",
	})
	typeNsQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_type_ns",
		Help: "The total number of Type NS DNS Query requests",
	})
	typeSoaQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_type_soa",
		Help: "The total number of Type SOA DNS Query requests",
	})
	unhandledQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_unhandled",
		Help: "The total number of unhandled DNS Query requests",
	})
)

const (
	WeekTtl = 60 * 60 * 24 * 7
)

func ServeDns(config *Config) {
	initRecords(config)
	dns.HandleFunc(config.Hostname, handle)
	server := &dns.Server{Addr: config.BindAddress, Net: "udp"}
	if err := server.ListenAndServe(); err != nil {
		zap.S().Panicw("failed to start server", "error", err.Error())
	}
}

func initRecords(config *Config) {
	initARecords(config)
	initNsRecords(config)
	initSoaRecord(config)
}

func initARecords(config *Config) {
	aRecords = make(map[string]net.IP, len(config.Nameserver)+1)

	// create A records for all NS
	for _, ns := range config.Nameserver {
		aRecords[ns.Hostname] = net.ParseIP(ns.Address)
	}

	// create A record for base host
	aRecords[config.Hostname] = net.ParseIP(config.Address)
}

func initNsRecords(config *Config) {
	header := dns.RR_Header{
		Name: config.Hostname, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: WeekTtl,
	}
	nsRecords = make([]dns.RR, len(config.Nameserver))
	for idx, ns := range config.Nameserver {
		nsRecords[idx] = &dns.NS{Hdr: header, Ns: ns.Hostname}
	}
}

func initSoaRecord(config *Config) {
	t := time.Now()
	yyyymmdd := (t.Year() * 10000) + (int(t.Month()) * 100) + (t.Day())
	serial := uint32(yyyymmdd * 100) // serial / last modification to zone should be everytime app starts
	soaRecord = []dns.RR{
		&dns.SOA{
			Hdr: dns.RR_Header{
				Name:   config.Hostname,
				Rrtype: dns.TypeSOA,
				Class:  dns.ClassINET,
				Ttl:    WeekTtl,
			},
			Ns:      config.Hostname,
			Mbox:    "admin." + config.Hostname,
			Serial:  serial,
			Refresh: 86400,   // 24 hours
			Retry:   7200,    // 2 hours
			Expire:  3600000, // 1000 hours
			Minttl:  172800,  // 2 days
		},
	}
}

func handle(w dns.ResponseWriter, request *dns.Msg) {
	reply := new(dns.Msg)
	reply.SetReply(request)
	reply.Authoritative = true
	reply.RecursionDesired = false
	reply.RecursionAvailable = false

	dnsRequests.Inc()

	if request.Opcode == dns.OpcodeQuery {
		defer w.WriteMsg(reply)

		if len(request.Question) < 1 {
			reply.Answer = soaRecord
			return
		}

		question := request.Question[0]
		name := strings.ToLower(question.Name)

		switch question.Qtype {
		case dns.TypeA:
			zap.S().Debugf("'A' Query for %s", name)
			typeAQueries.Inc()
			reply.Answer = append(reply.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: WeekTtl},
				A:   ipFromName(name),
			})
		case dns.TypeNS:
			zap.S().Debug("'NS' Query")
			typeNsQueries.Inc()
			reply.Answer = nsRecords
		case dns.TypeSOA:
			zap.S().Debug("'SOA' Query")
			typeSoaQueries.Inc()
			reply.Answer = soaRecord
		default:
			zap.S().Debug("Unhandled query type")
			unhandledQueries.Inc()
			reply.Answer = soaRecord
		}
	}
}

func ipFromName(name string) net.IP {
	// name:ip matches from config
	if record, ok := aRecords[name]; ok {
		return record
	}

	// ip extracted from name dash format, e.g. 10-10-10-1.rest.of.name
	if ipDashRegex.MatchString(name) {
		match := ipDashRegex.FindStringSubmatch(name)[2]
		ip := strings.Replace(match, "-", ".", -1)
		return net.ParseIP(ip)
	}

	// none of the above, just resolve to localhost
	return net.ParseIP("127.0.0.1")
}
