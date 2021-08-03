package test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	"github.ibm.com/mathewss/tf-helper/modules/terratest"
)

func TestMyTestCase(t *testing.T) {
	terratest.RunTestCase(t, &MyTestCase{}, &terraform.Options{
		// Where my module is
		TerraformDir: "..",

		// Variables supplied to module
		Vars: map[string]interface{}{
			"ibmcloud_api_key": os.Getenv("API_KEY"),

			// Application variables
			"vsi_name": "foobar",
		},
	})
}

type MyTestCase struct{}

func (m *MyTestCase) TestVSI(t *testing.T) {
	plan := terratest.GetPlan(t)
	vsi1 := terratest.GetResourcePlannedValues(t, plan, "ibm_is_instance.vsi1")

	assert.Equal(t, "cx2-2x4", vsi1["profile"])
	//	assert.NotEqual(t, "us-south-2", vsi1["zone"])
}

func (m *MyTestCase) GetOrderedTests() []string {
	return []string{
		"TestNIC",
		"TestVPC",
		"TestVSI",
	}
}

func (m *MyTestCase) TestNIC(t *testing.T) {
	plan := terratest.GetPlan(t)

	vsi1 := terratest.GetResourcePlannedValues(t, plan, "ibm_is_instance.vsi1")

	assert.Equal(t, "foobar", terratest.GetChildren(t, vsi1, "primary_network_interface")[0]["name"])
	assert.Equal(t, "foobar", terratest.GetFirstChild(t, vsi1, "primary_network_interface")["name"])
	assert.Equal(t, "foobar", terratest.GetResourcePlannedValues(t, plan, "ibm_is_instance.vsi1.primary_network_interface")["name"])
	assert.Equal(t, "foobar", terratest.GetResourcePlannedList(t, plan, "ibm_is_instance.vsi1.primary_network_interface")[0]["name"])
}

func (m *MyTestCase) TestVPC(t *testing.T) {
	plan := terratest.GetPlan(t)
	vpc := terratest.GetResourcePlannedValues(t, plan, "ibm_is_vpc.vpc")

	assert.Equal(t, "auto", vpc["address_prefix_management"])
	assert.Equal(t, false, vpc["classic_access"])
}

func (m *MyTestCase) TestVPCConnectivity(t *testing.T) {
	terratest.ApplyPlan(t)

	publicIp := terratest.Output(t, "public_ip")

	// This ping is an example, and should probably not be used in an actual test
	out, _ := exec.Command("ping", publicIp, "-c 5", "-i 3", "-w 10").Output()
	assert.NotContains(t, string(out), "0 received")
}

func (m *MyTestCase) TestVSIValues(t *testing.T) {
	state := terratest.GetState(t)
	vsi1 := terratest.GetResourceValues(t, state, "ibm_is_instance.vsi1")

	assert.Equal(t, 4, int(vsi1["memory"].(float64)))
	assert.Equal(t, "running", vsi1["status"])
	assert.NotEmpty(t, vsi1["id"])
}

func (m *MyTestCase) TestVPCValues(t *testing.T) {
	state := terratest.GetState(t)
	rules := terratest.GetResourceList(t, state, "ibm_is_vpc.vpc.security_group.rules")

	var outbound, inbound int = 0, 0
	for _, r := range rules {
		switch r["direction"] {
		case "outbound":
			outbound++
		case "inbound":
			inbound++
		}
	}

	assert.Equal(t, 1, outbound)
	assert.Equal(t, 1, inbound)
}

func (m *MyTestCase) TestMigrationChange(t *testing.T) {
	// Get the migration plan, send in options with new variables
	// than were used in our `prime` (TestMain) options
	migrationPlan := terratest.GetMigrationPlan(t, &terraform.Options{
		TerraformDir: "..",
		Vars: map[string]interface{}{
			"ibmcloud_api_key": os.Getenv("API_KEY"),
			"vsi_name":         "bazbar",
		},
	})

	resource := "ibm_is_instance.vsi1"

	// Values after the apply
	before := terratest.GetResourceBeforeChange(t, migrationPlan, resource)

	// Values with our migration plan run against the state
	after := terratest.GetResourceAfterChange(t, migrationPlan, resource)

	primaryNic := terratest.GetResourceAfterChange(t, migrationPlan, resource+".primary_network_interface")
	fmt.Printf("%+v\n", terratest.GetResourceListAfterChange(t, migrationPlan, resource+".primary_network_interface"))
	assert.Equal(t, "foobar", primaryNic["name"])

	// Assert some stuff
	assert.Equal(t, "tf-template-test-foobar", before["name"])
	assert.Equal(t, "tf-template-test-bazbar", after["name"])
}
