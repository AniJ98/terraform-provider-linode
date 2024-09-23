//go:build integration || instanceip

package instanceip_test

import (
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/acceptance"
	"github.com/linode/terraform-provider-linode/v2/linode/instanceip/tmpl"
)

const testInstanceIPResName = "linode_instance_ip.test"

var testRegion string

func init() {
	region, err := acceptance.GetRandomRegionWithCaps(nil, "core")
	if err != nil {
		log.Fatal(err)
	}

	testRegion = region
}

func TestAccInstanceIP_basic(t *testing.T) {
	t.Parallel()

	var instance linodego.Instance

	name := acctest.RandomWithPrefix("tf_test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV5ProviderFactories: acceptance.ProtoV5ProviderFactories,
		CheckDestroy:             acceptance.CheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Basic(t, name, testRegion, true),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckInstanceExists("linode_instance.foobar", &instance),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "address"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "gateway"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "prefix"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "rdns"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "subnet_mask"),
					resource.TestCheckResourceAttr(testInstanceIPResName, "region", testRegion),
					resource.TestCheckResourceAttr(testInstanceIPResName, "type", "ipv4"),
				),
			},
			{
				PreConfig: func() {
					acceptance.AssertInstanceReboot(t, true, &instance)
				},
				Config: tmpl.Basic(t, name, testRegion, true),
			},
		},
	})
}

func TestAccInstanceIP_noboot(t *testing.T) {
	t.Parallel()

	var instance linodego.Instance

	name := acctest.RandomWithPrefix("tf_test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV5ProviderFactories: acceptance.ProtoV5ProviderFactories,
		CheckDestroy:             acceptance.CheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.NoBoot(t, name, testRegion, true),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckInstanceExists("linode_instance.foobar", &instance),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "address"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "gateway"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "prefix"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "rdns"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "subnet_mask"),
					resource.TestCheckResourceAttr(testInstanceIPResName, "region", testRegion),
					resource.TestCheckResourceAttr(testInstanceIPResName, "type", "ipv4"),
				),
			},
			{
				Config: tmpl.NoBoot(t, name, testRegion, true),
				PreConfig: func() {
					acceptance.AssertInstanceReboot(t, false, &instance)
				},
			},
		},
	})
}

func TestAccInstanceIP_noApply(t *testing.T) {
	t.Parallel()

	var instance linodego.Instance

	name := acctest.RandomWithPrefix("tf_test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV5ProviderFactories: acceptance.ProtoV5ProviderFactories,
		CheckDestroy:             acceptance.CheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Basic(t, name, testRegion, false),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckInstanceExists("linode_instance.foobar", &instance),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "address"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "gateway"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "prefix"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "rdns"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "subnet_mask"),
					resource.TestCheckResourceAttr(testInstanceIPResName, "region", testRegion),
					resource.TestCheckResourceAttr(testInstanceIPResName, "type", "ipv4"),
				),
			},
			{
				PreConfig: func() {
					acceptance.AssertInstanceReboot(t, false, &instance)
				},
				Config: tmpl.Basic(t, name, testRegion, false),
			},
		},
	})
}

// func TestAccInstanceIP_addReservedIP(t *testing.T) {
// 	t.Parallel()

// 	var instance linodego.Instance
// 	name := acctest.RandomWithPrefix("tf_test")
// 	reservedIP := "45.33.74.65" // Replace with an actual reserved IP address

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { acceptance.PreCheck(t) },
// 		ProtoV5ProviderFactories: acceptance.ProtoV5ProviderFactories,
// 		CheckDestroy:             acceptance.CheckInstanceDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: tmpl.Basic(t, name, testRegion, false),
// 				Check: resource.ComposeTestCheckFunc(
// 					acceptance.CheckInstanceExists("linode_instance.foobar", &instance),
// 				),
// 			},
// 			{
// 				Config: tmpl.AddReservedIP(t, name, testRegion, reservedIP),
// 				Check: resource.ComposeTestCheckFunc(
// 					acceptance.CheckInstanceExists("linode_instance.foobar", &instance),
// 					resource.TestCheckResourceAttr(testInstanceIPResName, "address", reservedIP),
// 					resource.TestCheckResourceAttr(testInstanceIPResName, "public", "true"),
// 					resource.TestCheckResourceAttrSet(testInstanceIPResName, "linode_id"),
// 					resource.TestCheckResourceAttrSet(testInstanceIPResName, "gateway"),
// 					resource.TestCheckResourceAttrSet(testInstanceIPResName, "subnet_mask"),
// 					resource.TestCheckResourceAttrSet(testInstanceIPResName, "prefix"),
// 					resource.TestCheckResourceAttrSet(testInstanceIPResName, "rdns"),
// 					resource.TestCheckResourceAttr(testInstanceIPResName, "region", testRegion),
// 					resource.TestCheckResourceAttr(testInstanceIPResName, "type", "ipv4"),
// 				),
// 			},
// 		},
// 	})
// }

func TestAccInstanceIP_addReservedIP(t *testing.T) {
	t.Parallel()

	var instance linodego.Instance
	name := acctest.RandomWithPrefix("tf_test")
	reservedIP := "50.116.51.242" // Replace with your actual reserved IP address
	testRegion = "us-east"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV5ProviderFactories: acceptance.ProtoV5ProviderFactories,
		CheckDestroy:             acceptance.CheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.AddReservedIP(t, name, testRegion, reservedIP),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckInstanceExists("linode_instance.foobar", &instance),
					resource.TestCheckResourceAttr(testInstanceIPResName, "address", reservedIP),
					resource.TestCheckResourceAttr(testInstanceIPResName, "public", "true"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "linode_id"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "gateway"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "subnet_mask"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "prefix"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "rdns"),
					resource.TestCheckResourceAttr(testInstanceIPResName, "region", testRegion),
					resource.TestCheckResourceAttr(testInstanceIPResName, "type", "ipv4"),
				),
			},
		},
	})
}
