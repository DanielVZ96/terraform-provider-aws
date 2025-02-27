package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailDiskAttachment_basic(t *testing.T) {
	var disk lightsail.Disk
	resourceName := "aws_lightsail_disk_attachment.test"
	dName := sdkacctest.RandomWithPrefix("tf-acc-test")
	liName := sdkacctest.RandomWithPrefix("tf-acc-test")
	diskPath := "/dev/xvdf"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiskAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentConfig_basic(dName, liName, diskPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskAttachmentExists(resourceName, disk),
					resource.TestCheckResourceAttr(resourceName, "disk_name", dName),
					resource.TestCheckResourceAttr(resourceName, "disk_path", diskPath),
					resource.TestCheckResourceAttr(resourceName, "instance_name", liName),
				),
			},
		},
	})
}

func TestAccLightsailDiskAttachment_disappears(t *testing.T) {
	var disk lightsail.Disk
	resourceName := "aws_lightsail_disk_attachment.test"
	dName := sdkacctest.RandomWithPrefix("tf-acc-test")
	liName := sdkacctest.RandomWithPrefix("tf-acc-test")
	diskPath := "/dev/xvdf"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiskAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentConfig_basic(dName, liName, diskPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskAttachmentExists(resourceName, disk),
					acctest.CheckResourceDisappears(acctest.Provider, tflightsail.ResourceDiskAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDiskAttachmentExists(n string, disk lightsail.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailDiskAttachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		out, err := tflightsail.FindDiskAttachmentById(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Disk Attachment %q does not exist", rs.Primary.ID)
		}

		disk = *out

		return nil
	}
}

func testAccCheckDiskAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_disk_attachment" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		_, err := tflightsail.FindDiskAttachmentById(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResDiskAttachment, rs.Primary.ID, errors.New("still exists"))
	}

	return nil
}

func testAccDiskAttachmentConfig_basic(dName string, liName string, diskPath string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
resource "aws_lightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_lightsail_instance" "test" {
  name              = %[2]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_disk_attachment" "test" {
  disk_name     = aws_lightsail_disk.test.name
  instance_name = aws_lightsail_instance.test.name
  disk_path     = %[3]q
}
`, dName, liName, diskPath)
}
