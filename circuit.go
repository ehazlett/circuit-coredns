/*
  Copyright (c) Evan Hazlett

  Permission is hereby granted, free of charge, to any person
  obtaining a copy of this software and associated documentation
  files (the "Software"), to deal in the Software without
  restriction, including without limitation the rights to use, copy,
  modify, merge, publish, distribute, sublicense, and/or sell copies
  of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:
  The above copyright notice and this permission notice shall be
  included in all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
  OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE
  OR OTHER DEALINGS IN THE SOFTWARE.
*/

package circuit

import (
	"context"
	"io"
	"os"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/plugin/pkg/fall"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	cclient "github.com/ehazlett/circuit/client"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("circuit")

// Circuit is a circuit DNS resolver
type Circuit struct {
	client *cclient.Client

	Next plugin.Handler
	Fall fall.F
}

// RecordType is the type of DNS record
type RecordType string

// Record is a DNS record
type Record struct {
	Type  RecordType `json:"type,omitempty"`
	Name  string     `json:"name,omitempty"`
	Value string     `json:"value,omitempty"`
}

// New returns a new Circuit CoreDNS plugin
func New(ctx context.Context, socketPath string) (*Circuit, error) {
	log.Debugf("connecting to circuit on %s", socketPath)
	client, err := getClient(socketPath)
	if err != nil {
		return nil, err
	}
	return &Circuit{
		client: client,
	}, nil
}

// ServeDNS implements the plugin.Handler interface
func (c *Circuit) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	query := state.Name()

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = true
	// always return a response to prevent dns client lookup timeouts
	defer w.WriteMsg(m)

	records, err := c.lookup(ctx, query, state.QType())
	if err != nil {
		// log the error or it will be swallowed by ServeDNS
		log.Error(err)
		return -1, err
	}
	// no records found; pass through
	if len(records) == 0 {
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, w, r)
	}

	log.Debugf("answering: query=%s", query)
	m.Answer = records

	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	return dns.RcodeSuccess, nil
}

// Name returns the name of the plugin
func (c *Circuit) Name() string { return "circuit" }

func getClient(socketPath string) (*cclient.Client, error) {
	return cclient.NewClient(socketPath)
}

var out io.Writer = os.Stdout
