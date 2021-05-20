package resolvr

import (
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"net"
	"regexp"
	"strings"
)

// TODO: ipv6.?
var ipDashRegex = regexp.MustCompile(`(^|[.-])(((25[0-5]|(2[0-4]|1?[0-9])?[0-9])-){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9]))($|[.-])`)
var aRecords map[string]net.IP

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

func handle(w dns.ResponseWriter, request *dns.Msg) {
	reply := new(dns.Msg)
	reply.SetReply(request)

	if request.Opcode == dns.OpcodeQuery {
		switch request.Question[0].Qtype {
		case dns.TypeA:
			reply.Authoritative = true
			reply.RecursionDesired = false
			reply.RecursionAvailable = false

			name := request.Question[0].Name
			zap.S().Debugf("'A' Query for %s", name)

			// TODO: extract these to a single function

			// records from config
			if record, ok := aRecords[name]; ok {
				reply.Answer = append(reply.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   record,
				})
			}

			// record from ip contained in name
			if ipDashRegex.MatchString(name) {
				match := ipDashRegex.FindStringSubmatch(name)[2]
				ip := strings.Replace(match, "-", ".", -1)
				record := net.ParseIP(ip)
				reply.Answer = append(reply.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60 * 60 * 24 * 7},
					A:   record,
				})
			}
		}
		w.WriteMsg(reply)
	}
}
