package proplo

import "net"

type Options struct {
	LocalAddr    string
	UpstreamAddr string
	IgnoreCIDR   string
	ignoreIPNet  *net.IPNet
}

func (o *Options) Validate() error {
	if o.IgnoreCIDR != "" {
		_, cidrNet, err := net.ParseCIDR(o.IgnoreCIDR)
		if err != nil {
			return err
		}
		o.ignoreIPNet = cidrNet
	}
	return nil
}

func (o *Options) Ignore(ip net.IP) bool {
	return o.ignoreIPNet != nil && o.ignoreIPNet.Contains(ip)
}
