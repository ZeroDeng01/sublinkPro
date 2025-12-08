package geoip

import (
	"embed"
	"fmt"
	"log"
	"net/netip"
	"sync"

	"github.com/oschwald/geoip2-golang/v2"
)

//go:embed database/GeoLite2.mmdb
var geoIPDatabase embed.FS

var (
	geoIP *geoip2.Reader
	once  sync.Once
)

// InitGeoIP initializes the GeoIP database
func InitGeoIP() error {
	var err error
	once.Do(func() {
		data, readErr := geoIPDatabase.ReadFile("database/GeoLite2.mmdb")
		if readErr != nil {
			err = readErr
			log.Printf("Failed to read GeoIP database: %v", readErr)
			return
		}

		geoIP, err = geoip2.OpenBytes(data)
		if err != nil {
			log.Printf("Failed to initialize GeoIP reader: %v", err)
			return
		}
		log.Println("GeoIP database initialized successfully")
	})
	return err
}

// GetLocation returns the location information for the given IP address
func GetLocation(ipStr string) (string, error) {
	if geoIP == nil {
		if err := InitGeoIP(); err != nil {
			return "", err
		}
	}
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return "Unknown", nil
	}
	country := ""
	city := ""
	isocode := ""
	geoCountry, err := geoIP.Country(ip)
	if err != nil {
		log.Printf("Failed to get Country: %v", err)
	}
	if geoCountry.Country.HasData() {
		country = geoCountry.Country.Names.SimplifiedChinese
		isocode = geoCountry.Country.ISOCode
		flag := ISOCodeToFlag(isocode)
		if flag != "" {
			country = fmt.Sprintf("%s%s", flag, country)
		} else {
			country = fmt.Sprintf("(%s)%s", isocode, country)
		}
	}
	getCity, err := geoIP.City(ip)
	if err != nil {
		log.Printf("Failed to get City: %v", err)
	}
	if getCity.City.HasData() {
		city = getCity.City.Names.SimplifiedChinese
	}
	return fmt.Sprintf("%s%s", country, city), nil
}

// ISOCodeToFlag converts an ISO 3166-1 alpha-2 country code to a flag emoji
// Example: "CN" -> 🇨🇳, "US" -> 🇺🇸
func ISOCodeToFlag(isoCode string) string {
	if len(isoCode) != 2 {
		return ""
	}

	// Convert each letter to its corresponding Regional Indicator Symbol
	// Regional Indicator Symbols range from U+1F1E6 (A) to U+1F1FF (Z)
	flag := ""
	for _, char := range isoCode {
		if char >= 'A' && char <= 'Z' {
			// Convert A-Z to Regional Indicator Symbol
			flag += string(rune(0x1F1E6 + (char - 'A')))
		} else if char >= 'a' && char <= 'z' {
			// Convert a-z to Regional Indicator Symbol
			flag += string(rune(0x1F1E6 + (char - 'a')))
		}
	}
	return flag
}

// GetCountryISOCode returns only the ISO country code (e.g., "US", "CN", "JP") for the given IP address
func GetCountryISOCode(ipStr string) (string, error) {
	if geoIP == nil {
		if err := InitGeoIP(); err != nil {
			return "", err
		}
	}
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	geoCountry, err := geoIP.Country(ip)
	if err != nil {
		return "", fmt.Errorf("failed to get country: %v", err)
	}
	if geoCountry.Country.HasData() {
		return geoCountry.Country.ISOCode, nil
	}
	return "", nil
}

// Close closes the GeoIP reader
func Close() error {
	if geoIP != nil {
		return geoIP.Close()
	}
	return nil
}
