package packet

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccPacketPreCreatedIPBlockBasic(t *testing.T) {

	rs := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testPreCreatedIPBlockConfigBasic(rs),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.packet_precreated_ip_block.test", "cidr_notation"),
					resource.TestCheckResourceAttrPair(
						"packet_ip_attachment.test", "device_id",
						"packet_device.test", "id"),
				),
			},
			resource.TestStep{
				ResourceName:      "packet_ip_attachment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testPreCreatedIPBlockConfigBasic(name string) string {
	return fmt.Sprintf(`

resource "packet_project" "test" {
    name = "%s"
}

resource "packet_device" "test" {
  hostname         = "tftest"
  plan             = "baremetal_0"
  facility         = "ewr1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
}

data "packet_precreated_ip_block" "test" {
    facility         = "ewr1"
    project_id       = "${packet_device.test.project_id}"
    address_family   = 6
    public           = true
}

resource "packet_ip_attachment" "test" {
    device_id = "${packet_device.test.id}"
    cidr_notation = "${cidrsubnet(data.packet_precreated_ip_block.test.cidr_notation,8,2)}"
}
`, name)
}
