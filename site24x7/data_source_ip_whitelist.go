package site24x7

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceSite24x7IpWhitelist() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSite24x7IpWhitelistRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"city": {
							Type:     schema.TypeString,
							Required: true,
						},
						"place": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"ipv4": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ipv6": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSite24x7IpWhitelistRead(d *schema.ResourceData, meta interface{}) error {
	type Results struct {
		LocationDetails []struct {
			IPv6AddressExternal string `json:"IPv6_Address_External"`
			City                string `json:"City"`
			Place               string `json:"Place"`
			ExternalIP          string `json:"external_ip"`
		} `json:"LocationDetails"`
	}

	httpClient := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest(http.MethodGet, "https://creatorexport.zoho.com/site24x7/location-manager/json/IP_Address_View/C80EnP71mW2fDd60GaDgnPbVwMS8AGmP85vrN27EZ1CnCjPwnm0zPB5EX4Ct4q9n3rUnUgYwgwX0BW3KFtxnBqHt60Sz1Pgntgru", nil)
	if err != nil {
		return err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	results := Results{}
	err = json.Unmarshal(body, &results)
	if err != nil {
		return err
	}

	filters := d.Get("filter").(*schema.Set).List()
	ipv4 := make([]interface{}, 0)
	ipv6 := make([]interface{}, 0)

	if len(filters) != 0 {
		for f := range filters {
			city := filters[f].(map[string]interface{})["city"].(string)
			place := filters[f].(map[string]interface{})["place"].(string)

			log.Printf("City: %s, Place: %s\n", city, place)

			for _, r := range results.LocationDetails {
				if r.City == city && r.Place == place {
					ipv4 = appendValidate(ipv4, r.ExternalIP)
					ipv6 = appendValidate(ipv6, r.IPv6AddressExternal)
				}
			}
		}
	} else {
		for _, r := range results.LocationDetails {
			ipv4 = appendValidate(ipv4, r.ExternalIP)
			ipv6 = appendValidate(ipv6, r.IPv6AddressExternal)
		}

	}

	d.SetId(time.Now().UTC().String())
	d.Set("ipv4", ipToCidr(ipv4, 32))
	d.Set("ipv6", ipToCidr(ipv6, 128))

	return nil
}

func appendValidate(arr []interface{}, val string) []interface{} {
	if strings.Contains(val, "\n") {
		for _, ip := range strings.Split(val, "\n") {
			arr = append(arr, ip)
		}
		return arr
	}

	if strings.Contains(val, "\u200A") {
		return append(arr, strings.TrimLeft(val, "\u200A"))
	}

	if val != "" {
		return append(arr, val)
	}

	return arr
}
func ipToCidr(arr []interface{}, cidr int) (parsed []interface{}) {
	for _, ip := range arr {
		if strings.Contains(ip.(string), "/") {
			parsed = append(parsed, ip)
		} else {
			parsed = append(parsed, fmt.Sprintf("%s/%d", ip, cidr))
		}
	}

	return parsed
}
