# Creating Tests for Your Module

Your Terraform module is tested in this template using the Golang
[Terratest library](https://github.com/gruntwork-io/terratest) from Gruntwork.io.

Unit tests consist of running a validate and plan against the module and asserting known values
without running an apply to create infrastructure on an account. These tests are run on any PR.

Acceptance tests are tests that rely on an apply of the module creating real infrastructure to
perform tests against. In this mode you can get all the values that would have been unknown with
only a plan, and also perform tests against the application layer of the deployment. These tests are
run when a PR against master is opened, or when creating a release (tag).

## TerraTest Helper Module

A helper module (tf-helper) is used in the test to help with setting up the test. It will help with
running and syncing the terraform plan and apply commands, so you don't have to worry about how many
times you call plan or apply. It provides resource helper methods to make it easier to get planned
and changed values. Will only run the tests that do not need an apply when in `-short` (PR) mode.
And it will automatically destroy resources spun up at the end of the test.

### Getting Started with Tf-Helper

To initialize tf-helper, you just need to hook it into the `TestMain` function like so:

```golang
func TestMain(m *testing.M) {
    terratest.TestMain(m, &terraform.Options{
        // required, directory where your module is
        // unless you have a sub-module, it should just be ".."
        TerraformDir: "..",

        // variables for you module, 
        Vars: map[string]interface{}{
            // pass in the Travis encrypted api key
            "ibmcloud_api_key":  os.Getenv("API_KEY"),

            // custom module vars below
            "vsi_name": "foobar",
        },
    })
}
```

The
[Terratest options](https://github.com/gruntwork-io/terratest/blob/master/modules/terraform/options.go#L40)
supplied to `TestMain` will become the `prime` options for the test. Each test file can only have
one set of `prime` options. These will be the options that are used for the plan and apply. If you
want to test more than one set of options, multiple variable sets, create additional test files. Travis is
setup to run all the test files in the `test/` directory. However, you may want to see how existing
infrastructure will change when a new set of variables are applied against it. You can do this with
migration plans.

### Creating Your First Test

Creating a unit test is straightforward, just run a plan and assert resource values:

```golang
func TestVSI(t *testing.T) {
    // Get the plan (plan struct from resulting terraform plan command)
    // tf-helper will only run the plan once, so no need to worry how
    // many times you call it in your test file
    plan := terratest.GetPlan(t)

    // Get the resources you want to assert against
    vsi1 := terratest.GetResourcePlannedValues(t, plan, "ibm_is_instance.vsi1")

    // Assert stuff
    assert.Equal(t, "cx2-2x4", vsi1["profile"])
    assert.NotEqual(t, "us-south-2", vsi1["zone"])
}
```

### Acceptance Test

Acceptance tests will run when a PR is opened against master. These are tests that create
infrastructure and allow you to run various tests against it. You can also get the output from the
apply to assert against.

In the test below, the public ip of the VSI created is used to ping and ensure that our
infrastructure was deployed successfully.

```golang
func TestVPCConnectivity(t *testing.T) {
    // does not matter how many times ApplyPlan is called in a test file, it 
    // will only run the Apply and create the infrastructure one time. Always
    // uses the `prime` options specified in TestMain
    terratest.ApplyPlan(t)

    // output from the terratest apply
    publicIp := terratest.Output(t, "public_ip")

    // This ping is an example, and should probably not be used in an actual test
    out, _ := exec.Command("ping", publicIp, "-c 5", "-i 3", "-w 10").Output()
    assert.NotContains(t, string(out), "0 received")
}
```

### Migration (Day 2) Plan Test

This is a more advance test where you can test to see what would happen if new variables were
used in a Terraform Plan against existing state (after an apply). Because this requires state
the `prime` options are applied and infrastructure is created first. This test will only be
run in non `-short` mode.

Because you may want to run many different variable sets, each time `GetMigrationPlan` is
called, the Terraform command `plan` will be run.

```golang
func TestMigrationChange(t *testing.T) {
    // Get the migration plan, send in options with new variables
    // than were used in our `prime` (TestMain) options
    migrationPlan := terratest.GetMigrationPlan(t, &terraform.Options{
        TerraformDir: "..",
        Vars: map[string]interface{}{
            "ibmcloud_api_key":  os.Getenv("API_KEY"),
            "vsi_name": "bazbar",
        },
    })

    resource := "ibm_is_instance.vsi1"

    // Values after the apply
    before := terratest.GetResourceBeforeChange(t, migrationPlan, resource)

    // Values with our migration plan run against the state
    after := terratest.GetResourceAfterChange(t, migrationPlan, resource)

    // Assert some stuff
    assert.Equal(t, "vsi-sandbox-foobar", before["name"])
    assert.Equal(t, "vsi-sandbox-bazbar", after["name"])
}
```

## Terratest Resources

To see example tests, including ones in this readme, see the Sandbox link below.

- [TerraTest Helper](https://github.ibm.com/mathewss/tf-helper/tree/master/modules/terratest)
- [Sandbox Tests](https://github.ibm.com/mathewss/tf-sandbox/tree/master/test)
- [Terratest](https://github.com/gruntwork-io/terratest/tree/master/modules/terraform)
