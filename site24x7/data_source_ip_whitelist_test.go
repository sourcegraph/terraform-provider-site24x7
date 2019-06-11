package site24x7

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceSite24x7IpWhitelist_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSite24x7IpWhitelistConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccSote24x7IpWhitelistCheck("data.site24x7_ip_whitelist.test", "ipv4"),
					testAccSote24x7IpWhitelistCheck("data.site24x7_ip_whitelist.test", "ipv6"),
				),
			},
		},
	})
}

func TestAccDataSourceSite24x7IpWhitelist_filter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSite24x7IpWhitelistFilterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccSote24x7IpWhitelistCheck("data.site24x7_ip_whitelist.test", "ipv4"),
					testAccSote24x7IpWhitelistCheck("data.site24x7_ip_whitelist.test", "ipv6"),
				),
			},
		},
	})
}

func testAccSote24x7IpWhitelistCheck(name, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't access resource: %s", name)
		}

		attrs := rs.Primary.Attributes

		v, ok := attrs[fmt.Sprintf("%s.#", key)]
		if !ok {
			return fmt.Errorf("%s list is missing.", key)
		}

		qty, err := strconv.Atoi(v)
		if err != nil {
			return err
		}

		if qty < 1 {
			return fmt.Errorf("No %s addresses found.", key)
		}

		r := regexp.MustCompile(fmt.Sprintf("%s.[0-9]+", key))
		var count int

		for k, v := range attrs {
			if r.MatchString(k) {
				if strings.Contains(v, "/") {
					_, _, err := net.ParseCIDR(v)
					if err != nil {
						return fmt.Errorf("Error %v , '%s' is not a valid IP address", err, v)
					}
				} else {
					if net.ParseIP(v) == nil {
						return fmt.Errorf(" '%s' is not a valid IP address", v)
					}
				}
				count++
			}
		}

		fmt.Printf("Found %d %s addresses\n", count, key)
		return nil
	}
}

func testAccDataSourceSite24x7IpWhitelistConfig() string {
	return fmt.Sprintf(`
data "site24x7_ip_whitelist" "test" {}
`)
}

func testAccDataSourceSite24x7IpWhitelistFilterConfig() string {
	return fmt.Sprintf(`
data "site24x7_ip_whitelist" "test" {
  filter {
    city = "Denver"
    place = "US"
  }
}
`)
}
