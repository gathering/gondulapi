package types

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

// IP represent an IP and an optional netmask/size. It is provided because
// while net/IP provides much of what is needed for marshalling to/from
// text and json, it doesn't provide a convenient netmask. And it doesn't
// implement database/sql's Scanner either. This type implements both
// MarshalText/UnmarshalText and Scan(), and to top it off: String(). this
// should ensure maximum usability.
type IP struct {
	IP   net.IP
	Mask int
}

func (i *IP) Scan(src interface{}) error {
	switch value := src.(type) {
	case string:
		return i.UnmarshalText([]byte(value))
	case []byte:
		return i.UnmarshalText(value)
	default:
		return fmt.Errorf("invalid IP")
	}
}
func (i *IP) String() string {
	if i.Mask != 0 {
		return fmt.Sprintf("%s/%d", i.IP.String(), i.Mask)
	}
	return i.IP.String()
}

func (i IP) MarshalText() ([]byte, error) {
	log.Printf("Marshalling i: %v", i)
	return []byte(i.String()), nil
}

func (i *IP) UnmarshalText(b []byte) error {
	s := string(b)
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		log.Printf("Failed to use ParseCIDR on %s, falling back...", s)
		ip = net.ParseIP(s)
		i.Mask = 0
		if ip == nil {
			return err
		}
	} else {
		i.Mask, _ = ipnet.Mask.Size()
	}
	i.IP = ip
	return nil
}
