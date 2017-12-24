package packet

import (
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/packethost/packngo"
)

// Regexp vars for use with resource.ExpectError, resource.TestMatchResourceAttr, etc.
var matchErrMustBeProvided = regexp.MustCompile(".* must be provided when .*")
var matchErrShouldNotBeAnIPXE = regexp.MustCompile(`.*"user_data" should not be an iPXE.*`)
var matchErrShouldOnlyBeProvided = regexp.MustCompile(".* should only be provided when .*")
var matchErrOutOfRange = regexp.MustCompile(".* is out of range .*")
var matchErrIsNotValid = regexp.MustCompile(".* is not a valid value for.*")
var matchAttrDuration = regexp.MustCompile(`^\dh\d{1,2}m\d{1,2}s$`)

func TestAccPacketDeviceBasic(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigBasic, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					testAccCheckPacketDeviceAttributes(&device),
					resource.TestCheckResourceAttr(
						r, "public_ipv4_subnet_size", "31"),
					resource.TestCheckResourceAttr(
						r, "ipxe_script_url", ""),
					resource.TestCheckResourceAttr(
						r, "always_pxe", "false"),
					resource.TestCheckResourceAttrSet(
						r, "root_password"),
					resource.TestCheckResourceAttr(
						r, "spot_instance", "false"),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", ""),
					resource.TestCheckResourceAttr(
						r, "termination_time", ""),
				),
			},
		},
	})
}

func TestAccPacketDeviceRequestSubnet(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_subnet_29"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigRequestSubnet, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "public_ipv4_subnet_size", "29"),
				),
			},
		},
	})
}

func TestAccPacketDeviceIPXEScriptURL(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_ipxe_script_url"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigIpxeScriptURL, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "ipxe_script_url", "https://boot.netboot.xyz"),
					resource.TestCheckResourceAttr(
						r, "always_pxe", "true"),
				),
			},
		},
	})
}

func TestAccPacketDeviceIPXEConflictingFields(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_ipxe_conflict"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigIpxeConflict, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
				),
				ExpectError: matchErrShouldNotBeAnIPXE,
			},
		},
	})
}

func TestAccPacketDeviceIPXEConfigMissing(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_ipxe_config_missing"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigIpxeMissing, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
				),
				ExpectError: matchErrMustBeProvided,
			},
		},
	})
}

func TestAccPacketDeviceSpotInstance(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_spot_instance"
	si := "true"
	spm := "0.01"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigSpotInstance,
					rs, si, spm, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "spot_instance", si),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", spm),
				),
			},
		},
	})
}

func TestAccPacketDeviceSpotTermRFC3339(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_spot_instance"
	si := "true"
	spm := "0.01"
	ttd := time.Duration(time.Hour * 6).Round(terminationTimeRoundVal)
	tn := time.Now()
	tt := tn.Add(ttd).Round(terminationTimeRoundVal)
	ttRFC := tt.Format(time.RFC3339)

	// Test termination_time with RFC3339 format
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigSpotInstance,
					rs, si, spm, ttRFC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "spot_instance", si),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", spm),
					resource.TestCheckResourceAttr(
						r, "termination_time", ttRFC),
					resource.TestCheckResourceAttr(
						r, "termination_timestamp", ttRFC),
					resource.TestMatchResourceAttr(
						r, "termination_time_remaining",
						matchAttrDuration),
				),
			},
		},
	})
}

func TestAccPacketDeviceSpotTermDuration(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_spot_instance"
	si := "true"
	spm := "0.01"
	ttd := time.Duration(time.Hour * 6).Round(terminationTimeRoundVal)
	tn := time.Now()
	tt := tn.Add(ttd).Round(terminationTimeRoundVal)
	ttRFC := tt.Format(time.RFC3339)

	// Test termination_time with Duration format
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigSpotInstance,
					rs, si, spm, ttd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "spot_instance", si),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", spm),
					resource.TestCheckResourceAttr(
						r, "termination_time", ttd.String()),
					resource.TestCheckResourceAttr(
						r, "termination_timestamp", ttRFC),
					resource.TestMatchResourceAttr(
						r, "termination_time_remaining",
						matchAttrDuration),
				),
			},
		},
	})
}

func TestAccPacketDeviceSpotInstanceInvalid(t *testing.T) {
	rs := acctest.RandString(10)
	si := "true"
	ttd := time.Duration(time.Hour * 6)
	tt := time.Now().Add(ttd)
	ttRFC := tt.Format(time.RFC3339)
	spm := "0.01"

	// Invalid termination_time test: Wrong format
	badTime := tt.Format(time.Stamp)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigSpotInstance,
					rs, si, spm, badTime),
				ExpectError: matchErrIsNotValid,
			},
		},
	})

	// spot_instance false, but other spot fields set
	si = "false"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigSpotInstance,
					rs, si, spm, ttRFC),
				ExpectError: matchErrShouldOnlyBeProvided,
			},
		},
	})
	si = "true" // Reset

	// Missing spot_price_max
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfigSpotPriceMissing,
					rs, si, ttRFC),
				ExpectError: matchErrMustBeProvided,
			},
		},
	})
}

func testAccCheckPacketDeviceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*packngo.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "packet_device" {
			continue
		}
		if _, _, err := client.Devices.Get(rs.Primary.ID); err == nil {
			return fmt.Errorf("Device still exists")
		}
	}
	return nil
}

func testAccCheckPacketDeviceAttributes(device *packngo.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if device.Hostname != "test-device" {
			return fmt.Errorf("Bad name: %s", device.Hostname)
		}
		if device.State != "active" {
			return fmt.Errorf("Device should be 'active', not '%s'", device.State)
		}

		return nil
	}
}

func testAccCheckPacketDeviceExists(n string, device *packngo.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*packngo.Client)

		foundDevice, _, err := client.Devices.Get(rs.Primary.ID)
		if err != nil {
			return err
		}
		if foundDevice.ID != rs.Primary.ID {
			return fmt.Errorf("Record not found: %v - %v", rs.Primary.ID, foundDevice)
		}

		*device = *foundDevice

		return nil
	}
}

func testAccCheckPacketDeviceNetwork(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var ip net.IP
		var k, v string
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		k = "access_public_ipv6"
		v = rs.Primary.Attributes[k]
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("\"%s\" is not a valid IP address: %s",
				k, v)
		}

		k = "access_public_ipv4"
		v = rs.Primary.Attributes[k]
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("\"%s\" is not a valid IP address: %s",
				k, v)
		}

		k = "access_private_ipv4"
		v = rs.Primary.Attributes[k]
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("\"%s\" is not a valid IP address: %s",
				k, v)
		}

		return nil
	}
}

var testAccCheckPacketDeviceConfigBasic = `
resource "packet_project" "test" {
    name = "TerraformTestProject-%s"
}

resource "packet_device" "test" {
  hostname         = "test-device"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
}`

var testAccCheckPacketDeviceConfigRequestSubnet = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_subnet_29" {
  hostname         = "test-subnet-29"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  public_ipv4_subnet_size = 29
}`

var testAccCheckPacketDeviceConfigIpxeScriptURL = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_ipxe_script_url" {
  hostname         = "test-ipxe-script-url"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "custom_ipxe"
  user_data        = "#!/bin/sh\ntouch /tmp/test"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  ipxe_script_url  = "https://boot.netboot.xyz"
  always_pxe       = true
}`

var testAccCheckPacketDeviceConfigIpxeConflict = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_ipxe_conflict" {
  hostname         = "test-ipxe-conflict"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "custom_ipxe"
  user_data        = "#!ipxe\nset conflict ipxe_script_url"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  ipxe_script_url  = "https://boot.netboot.xyz"
  always_pxe       = true
}`

var testAccCheckPacketDeviceConfigIpxeMissing = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_ipxe_missing" {
  hostname         = "test-ipxe-missing"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  always_pxe       = true
}`

var testAccCheckPacketDeviceConfigSpotInstance = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}
resource "packet_device" "test_spot_instance" {
  hostname         = "test-spot-instance"
  plan             = "baremetal_0"
  facility         = "nrt1"
  operating_system = "coreos_stable"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  spot_instance    = %s
  spot_price_max   = %s
  termination_time = "%s"
}`

var testAccCheckPacketDeviceConfigSpotPriceMissing = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}
resource "packet_device" "test_spot_instance" {
  hostname         = "test-spot-instance"
  plan             = "baremetal_0"
  facility         = "nrt1"
  operating_system = "coreos_stable"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  spot_instance    = %s
  termination_time = "%s"
}`
