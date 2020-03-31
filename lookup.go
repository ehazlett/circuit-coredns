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
	"fmt"
	"net"
	"strings"

	"github.com/containerd/containerd/errdefs"
	api "github.com/ehazlett/circuit/api/circuit/v1"
	"github.com/miekg/dns"
)

func (c *Circuit) lookup(ctx context.Context, query string, qtype uint16) ([]dns.RR, error) {
	// remove trailing dot
	q := strings.TrimSuffix(query, ".")
	// split into host / domain
	x := strings.SplitN(q, ".", 2)
	if len(x) != 2 {
		return nil, fmt.Errorf("invalid query %s", query)
	}

	host := x[0]
	domain := x[1]

	resp, err := c.client.GetContainerIPs(ctx, &api.GetContainerIPsRequest{
		Container: host,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// split domain to get network name
	n := strings.SplitN(domain, ".", 2)
	network := n[0]

	for _, cip := range resp.IPs {
		if cip.Network == network {
			return []dns.RR{
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   query,
						Ttl:    0,
						Class:  dns.ClassINET,
						Rrtype: dns.TypeA,
					},
					A: net.ParseIP(cip.IP),
				},
			}, nil
		}
	}

	return nil, nil

	//resp, err := d.client.GetZone(ctx, &api.GetZoneRequest{
	//	Name: domain,
	//})
	//if err != nil {
	//	return nil, err
	//}

	//zone := resp.Zone

	//records := []dns.RR{}
	//for _, r := range zone.Records {
	//	if r.Name != host {
	//		continue
	//	}
	//	var record dns.RR
	//	hdr := dns.RR_Header{
	//		Name:  query,
	//		Ttl:   0,
	//		Class: dns.ClassINET,
	//	}
	//	switch r.Type {
	//	case v1.Record_A:
	//		hdr.Rrtype = dns.TypeA
	//		record = &dns.A{
	//			Hdr: hdr,
	//			A:   net.ParseIP(r.Value),
	//		}
	//		records = append(records, record)
	//	case v1.Record_CNAME:
	//		hdr.Rrtype = dns.TypeCNAME
	//		record = &dns.CNAME{
	//			Hdr:    hdr,
	//			Target: r.Value + ".",
	//		}
	//		records = append(records, record)
	//		// lookup corresponding A records for target
	//		resolved, err := d.lookup(ctx, r.Value, dns.TypeA)
	//		if err != nil {
	//			return nil, err
	//		}

	//		for _, x := range resolved {
	//			v, ok := x.(*dns.A)
	//			if !ok {
	//				continue
	//			}
	//			records = append(records, &dns.A{
	//				Hdr: dns.RR_Header{
	//					Name:   r.Value + ".",
	//					Ttl:    0,
	//					Class:  dns.ClassINET,
	//					Rrtype: dns.TypeA,
	//				},
	//				A: v.A,
	//			})
	//		}
	//	default:
	//		log.Errorf("unsupported record type: %+v", r)
	//		continue
	//	}
	//}

	//return records, nil
}
