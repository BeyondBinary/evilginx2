package core

import (
	"bufio"
	"net"
	"os"
	"strings"

	"github.com/kgretzky/evilginx2/log"
	"github.com/oschwald/geoip2-golang"
)

const (
	GEOIP_MODE_BLOCKLIST   = 0
	GEOIP_MODE_ALLOWLIST = 1
	GEOIP_MODE_OFF    = 2
)

type GeoIPList struct {
	countries	[]string
	mode		int
	lookup		*geoip2.Reader
}

func NewGeoIPList(path string, db_path string) (*GeoIPList, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lookup,_ := geoip2.Open(db_path)
	if lookup == nil {
		log.Error("Couldn't open GeoIP database");
	}

	geoip := &GeoIPList{
		mode:       GEOIP_MODE_ALLOWLIST,
		lookup:	    lookup,
	}

	fs := bufio.NewScanner(f)
	fs.Split(bufio.ScanLines)

	for fs.Scan() {
		l := fs.Text()
		l = strings.Trim(l, " ")
		geoip.countries = append(geoip.countries, l)
	}
		// remove comments
	log.Info("GeoIP list: loaded %d countries", len(geoip.countries))
	return geoip, nil
}

func (geoip *GeoIPList) SetGeoIPMode(mode string) {
	switch mode {
	case "block":
		geoip.mode = GEOIP_MODE_BLOCKLIST
	case "allow":
		geoip.mode = GEOIP_MODE_ALLOWLIST
	case "off":
		geoip.mode = GEOIP_MODE_OFF
	}
}

func (geoip *GeoIPList) IsAllowed(ip string) (bool, string) {
	if geoip.mode == GEOIP_MODE_OFF {
		return true, ""
	}

	ipv4 := net.ParseIP(ip)
	if ipv4 == nil {
		return false, "Invalid"
	}

	country,_ := geoip.lookup.Country(ipv4)
	if country == nil {
		log.Warning("Unknown country: %s", ip)
		return false, "Unknown"
	}

	found := false
	for _, c := range geoip.countries {
		if c == country.Country.IsoCode {
			found = true
			break
		}
	}
	return found == (geoip.mode == GEOIP_MODE_ALLOWLIST), country.Country.IsoCode
}
