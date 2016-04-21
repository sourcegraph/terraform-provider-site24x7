package site24x7

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestWebsiteMonitor(t *testing.T) {
	const config1 = `
		resource "site24x7_website_monitor" "test" {
			display_name = "test"
			website = "https://www.sourcegraph.com"
		}
	`

	const config2 = `
		resource "site24x7_website_monitor" "test" {
			display_name = "new name"
			website = "https://www.sourcegraph.com/login"
			custom_headers { "foo" = "bar" }
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkWebsiteMonitorDestroyed,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					checkWebsiteMonitorExists,
				),
			},

			resource.TestStep{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					checkWebsiteMonitorExists,
					resource.TestCheckResourceAttr("site24x7_website_monitor.test", "display_name", "new name"),
				),
			},
		},
	})
}

func checkWebsiteMonitorExists(s *terraform.State) error {
	rs := s.RootModule().Resources["site24x7_website_monitor.test"]
	exists, err := fetchWebsiteMonitorExists(testAccProvider.Meta().(*http.Client), rs.Primary.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("monitor not found")
	}
	return nil
}

func checkWebsiteMonitorDestroyed(s *terraform.State) error {
	rs := s.RootModule().Resources["site24x7_website_monitor.test"]
	exists, err := fetchWebsiteMonitorExists(testAccProvider.Meta().(*http.Client), rs.Primary.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("monitor still exists")
	}
	return nil
}
