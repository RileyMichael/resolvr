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
	ipDashRegex            = regexp.MustCompile(`(^|[.-])(((25[0-5]|(2[0-4]|1?[0-9])?[0-9])-){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9]))($|[.-])`)
	staticTypeARecords     map[string]dns.RR
	staticTypeAAAARecords  map[string]dns.RR
	staticTypeCNAMERecords map[string]dns.RR
	staticTypeNSRecords    []dns.RR
	staticTypeSOARecord    []dns.RR
	dnsRequests            = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_total",
		Help: "The total number of DNS requests",
	})
	typeAQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_type_a",
		Help: "The total number of Type A DNS Query requests not matching a static record",
	})
	unhandledQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dns_requests_unhandled",
		Help: "The total number of unhandled DNS Query requests",
	})
)

const (
	WeekTtl       = 60 * 60 * 24 * 7
	FiveMinuteTtl = 60 * 5
)

func ServeDns(config *Config) {
	initStaticRecords(config)
	dns.HandleFunc(config.Hostname, handle)
	server := &dns.Server{Addr: config.BindAddress, Net: "udp"}
	if err := server.ListenAndServe(); err != nil {
		zap.S().Panicw("failed to start server", "error", err.Error())
	}
}
func initStaticRecords(config *Config) {
	zap.S().Debugw("initializing static records", "config", config)
	initTypeARecords(config)
	initTypeAAAARecords(config)
	initTypeCNAMERecords(config)
	initRecordsForNameservers(config)
	initSoaRecord(config)
}

func initTypeARecords(config *Config) {
	staticTypeARecords = make(map[string]dns.RR, len(config.StaticTypeARecords))
	for _, record := range config.StaticTypeARecords {
		staticTypeARecords[record.First] = &dns.A{Hdr: dns.RR_Header{
			Name: record.First, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: WeekTtl,
		}, A: net.ParseIP(record.Second)}
	}
}

func initTypeAAAARecords(config *Config) {
	staticTypeAAAARecords = make(map[string]dns.RR, len(config.StaticTypeAAAARecords))
	for _, record := range config.StaticTypeAAAARecords {
		staticTypeAAAARecords[record.First] = &dns.AAAA{Hdr: dns.RR_Header{
			Name: record.First, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: WeekTtl,
		}, AAAA: net.ParseIP(record.Second)}
	}
}

func initTypeCNAMERecords(config *Config) {
	staticTypeCNAMERecords = make(map[string]dns.RR, len(config.StaticTypeCNAMERecords))
	for _, record := range config.StaticTypeCNAMERecords {
		staticTypeCNAMERecords[record.First] = &dns.CNAME{Hdr: dns.RR_Header{
			Name: record.First, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: WeekTtl,
		}, Target: record.Second}
	}
}

func initRecordsForNameservers(config *Config) {
	staticTypeNSRecords = make([]dns.RR, len(config.Nameservers))
	nsHeader := dns.RR_Header{
		Name: config.Hostname, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: WeekTtl,
	}
	for idx, record := range config.Nameservers {
		staticTypeNSRecords[idx] = &dns.NS{Hdr: nsHeader, Ns: record.First}
		staticTypeARecords[record.First] = &dns.A{Hdr: dns.RR_Header{
			Name: record.First, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: WeekTtl,
		}, A: net.ParseIP(record.Second)}
	}
}

func initSoaRecord(config *Config) {
	t := time.Now()
	yyyymmdd := (t.Year() * 10000) + (int(t.Month()) * 100) + (t.Day())
	serial := uint32(yyyymmdd * 100) // serial / last modification to zone should be everytime app starts
	staticTypeSOARecord = []dns.RR{
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
			reply.Answer = staticTypeSOARecord
			return
		}

		question := request.Question[0]
		name := strings.ToLower(question.Name)
		zap.S().Debugw("Attempting to handle query",
			"type", dns.TypeToString[question.Qtype],
			"name", question.Name,
		)
		switch question.Qtype {
		case dns.TypeA:
			if record, ok := staticTypeCNAMERecords[name]; ok {
				// if we've defined a static cname, resolve the A records for the canonical host
				reply.Answer = append(reply.Answer, record)
				m := new(dns.Msg)
				m.SetQuestion(record.(*dns.CNAME).Target, dns.TypeA)
				d := new(dns.Client)
				// todo: specify resolver(s) in config, and this could error
				response, _, _ := d.Exchange(m, "8.8.8.8:53")
				for _, a := range response.Answer {
					if r, ok := a.(*dns.A); ok {
						reply.Answer = append(reply.Answer, &dns.A{
							Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: FiveMinuteTtl},
							A:   r.A,
						})
					}
				}
			} else if record, ok := staticTypeARecords[name]; ok {
				reply.Answer = append(reply.Answer, record)
			} else {
				typeAQueries.Inc()
				reply.Answer = append(reply.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: WeekTtl},
					A:   ipFromName(name),
				})
			}
		case dns.TypeAAAA:
			if val, ok := staticTypeAAAARecords[name]; ok {
				reply.Answer = append(reply.Answer, val)
			} else {
				reply.Answer = staticTypeSOARecord
			}
		case dns.TypeCNAME:
			if val, ok := staticTypeCNAMERecords[name]; ok {
				reply.Answer = append(reply.Answer, val)
			} else {
				reply.Answer = staticTypeSOARecord
			}
		case dns.TypeNS:
			// should probably only return for the root Hostname..
			reply.Answer = staticTypeNSRecords
		case dns.TypeSOA:
			reply.Answer = staticTypeSOARecord
		default:
			unhandledQueries.Inc()
			reply.Answer = staticTypeSOARecord
		}
	}
}

func ipFromName(name string) net.IP {
	// ip extracted from name dash format, e.g. 10-10-10-1.rest.of.name
	if ipDashRegex.MatchString(name) {
		match := ipDashRegex.FindStringSubmatch(name)[2]
		ip := strings.Replace(match, "-", ".", -1)
		return net.ParseIP(ip)
	}

	// none of the above, just resolve to localhost
	return net.ParseIP("127.0.0.1")
}
